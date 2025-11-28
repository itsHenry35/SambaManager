package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/types"
)

var (
	// Valid username pattern: alphanumeric, underscore, dash, 1-32 chars
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	// Valid share name pattern: alphanumeric and Chinese characters, no special symbols
	shareNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\p{Han}]+$`)
	// username-share-{anything} (timestamp or custom name)
	sharePattern = regexp.MustCompile(`^([a-zA-Z0-9_-]+)-share-(.+)$`)
)

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func isValidShareName(name string) bool {
	if name == "" {
		return true // Empty is allowed, will use timestamp
	}
	return shareNameRegex.MatchString(name)
}

// cleanSubPath cleans and validates a subdirectory path
// Returns cleaned path and error if invalid
func cleanSubPath(subPath string) (string, error) {
	if subPath == "" {
		return "", nil
	}

	// Clean the subpath
	cleaned := filepath.Clean(subPath)
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimPrefix(cleaned, "\\")

	// Prevent path traversal
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("invalid subdirectory path: path traversal not allowed")
	}

	// Empty or "." means no subdirectory
	if cleaned == "" || cleaned == "." {
		return "", nil
	}

	return cleaned, nil
}

type SambaService struct {
	mu sync.RWMutex
}

func NewSambaService() *SambaService {
	return &SambaService{}
}

// CreateUser creates a Samba user with Unix user in extrausers
func (s *SambaService) CreateUser(user *types.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate username (alphanumeric, underscore, dash only)
	if !isValidUsername(user.Username) {
		return fmt.Errorf("invalid username: must contain only letters, numbers, underscore, and dash")
	}

	// Create home directory for the user (owned by root)
	homeDir := filepath.Join(config.AppConfig.HomeDir, user.Username)
	if err := os.MkdirAll(homeDir, 0770); err != nil {
		return fmt.Errorf("failed to create home directory: %v", err)
	}

	// Set ownership to root:root
	cmd := exec.Command("chown", "-R", "root:root", homeDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set directory ownership: %v, output: %s", err, output)
	}

	// Set permissions to 770
	cmd = exec.Command("chmod", "-R", "770", homeDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set directory permissions: %v, output: %s", err, output)
	}

	// Create Unix user in extrausers with nologin shell
	// useradd with --extrausers flag to write to /var/lib/extrausers/passwd
	cmd = exec.Command("useradd",
		"--extrausers",
		"--no-create-home",
		"--shell", "/usr/sbin/nologin",
		"--home-dir", homeDir,
		"--badname",
		user.Username)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(homeDir)
		return fmt.Errorf("failed to create unix user: %v, output: %s", err, output)
	}

	// Create Samba user using stdin (smbpasswd -a creates the user in tdbsam)
	cmd = exec.Command("smbpasswd", "-a", "-s", user.Username)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		// Cleanup: remove unix user and home directory if Samba user creation fails
		_ = exec.Command("userdel", "--extrausers", user.Username).Run()
		_ = os.RemoveAll(homeDir)
		return fmt.Errorf("failed to create stdin pipe for smbpasswd: %v", err)
	}

	if err := cmd.Start(); err != nil {
		_ = exec.Command("userdel", "--extrausers", user.Username).Run()
		_ = os.RemoveAll(homeDir)
		return fmt.Errorf("failed to start smbpasswd: %v", err)
	}

	// smbpasswd expects password twice
	_, err = stdin.Write([]byte(user.Password + "\n" + user.Password + "\n"))
	stdin.Close()

	if err != nil {
		_ = exec.Command("userdel", "--extrausers", user.Username).Run()
		_ = os.RemoveAll(homeDir)
		return fmt.Errorf("failed to write samba password: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		_ = exec.Command("userdel", "--extrausers", user.Username).Run()
		_ = os.RemoveAll(homeDir)
		return fmt.Errorf("failed to create samba user: %v", err)
	}

	// Enable Samba user
	cmd = exec.Command("smbpasswd", "-e", user.Username)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = exec.Command("userdel", "--extrausers", user.Username).Run()
		return fmt.Errorf("failed to enable samba user: %v, output: %s", err, output)
	}

	return nil
}

// DeleteUser deletes a Samba user and optionally their home directory, and cleans up shares
func (s *SambaService) DeleteUser(username string, deleteHomeDir bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read config file once
	configPath := config.AppConfig.Samba.ConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read samba config: %v", err)
	}

	// Parse shares from content
	shares, err := s.parseSharesFromContent(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse shares: %v", err)
	}

	// Collect shares to delete and update
	sharesToDelete := make(map[string]bool)
	sharesToUpdate := make(map[string]*types.Share)

	for _, share := range shares {
		// If user is the owner, mark share for deletion
		if share.Owner == username {
			sharesToDelete[share.ID] = true
			continue
		}

		// If user is in shared_with list, remove them
		if contains(share.SharedWith, username) {
			newSharedWith := removeFromSlice(share.SharedWith, username)

			// If no users left, mark share for deletion
			if len(newSharedWith) == 0 {
				sharesToDelete[share.ID] = true
			} else {
				// Mark share for update
				sharesToUpdate[share.ID] = &types.Share{
					Owner:      share.Owner,
					SharedWith: newSharedWith,
					ReadOnly:   share.ReadOnly,
					Comment:    share.Comment,
				}
			}
		}
	}

	// Modify config in memory: delete and update shares
	if len(sharesToDelete) > 0 || len(sharesToUpdate) > 0 {
		newContent, err := s.modifySharesInContent(string(content), sharesToDelete, sharesToUpdate)
		if err != nil {
			return fmt.Errorf("failed to modify shares: %v", err)
		}

		// Write back to file once
		if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write samba config: %v", err)
		}
	}

	// Remove Samba user from tdbsam
	cmd := exec.Command("smbpasswd", "-x", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete samba user: %v, output: %s", err, output)
	}

	// Remove Unix user from extrausers
	cmd = exec.Command("userdel", "--extrausers", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete unix user: %v, output: %s", err, output)
	}

	// Delete user's home directory (optional)
	if deleteHomeDir {
		homeDir := filepath.Join(config.AppConfig.HomeDir, username)
		if err := os.RemoveAll(homeDir); err != nil {
			return fmt.Errorf("failed to delete home directory: %v", err)
		}
	}

	// Hot reload Samba configuration
	_ = ReloadSambaConfig()

	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// Helper function to remove an item from a string slice
func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// ListUsers lists all Samba users
func (s *SambaService) ListUsers() ([]types.UserResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cmd := exec.Command("pdbedit", "-L")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list samba users: %v, output: %s", err, output)
	}

	var users []types.UserResponse
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			username := parts[0]
			homeDir := filepath.Join(config.AppConfig.HomeDir, username)
			users = append(users, types.UserResponse{
				Username: username,
				HomeDir:  homeDir,
			})
		}
	}

	return users, nil
}

// CreateShare creates a Samba share for a user's directory (supports multiple shares per owner)
func (s *SambaService) CreateShare(share *types.Share) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate owner username
	if !isValidUsername(share.Owner) {
		return "", fmt.Errorf("invalid owner username")
	}

	// Validate custom share name if provided
	if !isValidShareName(share.Name) {
		return "", fmt.Errorf("invalid share name: must contain only alphanumeric characters or Chinese characters, no symbols")
	}

	// Validate shared_with usernames
	if len(share.SharedWith) == 0 {
		return "", fmt.Errorf("must share with at least one user")
	}
	for _, username := range share.SharedWith {
		if !isValidUsername(username) {
			return "", fmt.Errorf("invalid username in shared_with: %s", username)
		}
	}

	// Get owner's home directory
	ownerHome := filepath.Join(config.AppConfig.HomeDir, share.Owner)
	if _, err := os.Stat(ownerHome); os.IsNotExist(err) {
		return "", fmt.Errorf("owner's home directory does not exist")
	}

	// Validate and check subdirectory path if specified
	if share.SubPath != "" {
		cleanedSubPath, err := cleanSubPath(share.SubPath)
		if err != nil {
			return "", err
		}
		if cleanedSubPath != "" {
			fullSubPath := filepath.Join(ownerHome, cleanedSubPath)
			// Create the subdirectory if it doesn't exist
			if err := os.MkdirAll(fullSubPath, 0770); err != nil {
				return "", fmt.Errorf("failed to create subdirectory: %v", err)
			}
			// Set ownership to root:root
			cmd := exec.Command("chown", "-R", "root:root", fullSubPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("failed to set subdirectory ownership: %v, output: %s", err, output)
			}
			// Set permissions to 770
			cmd = exec.Command("chmod", "-R", "770", fullSubPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("failed to set subdirectory permissions: %v, output: %s", err, output)
			}
		}
	}

	// Read existing smb.conf once
	configPath := config.AppConfig.Samba.ConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read samba config: %v", err)
	}

	// Generate share name based on custom name or timestamp
	var shareName string
	if share.Name != "" {
		// Use custom name: username-share-customname
		shareName = fmt.Sprintf("%s-share-%s", share.Owner, share.Name)
		// Check if this share name already exists
		if strings.Contains(string(content), fmt.Sprintf("[%s]", shareName)) {
			return "", fmt.Errorf("share name '%s' already exists for this user", share.Name)
		}
	} else {
		// Use timestamp: username-share-YYYYMMDDHHMMSS
		timestamp := time.Now().Format("20060102150405")
		shareName = fmt.Sprintf("%s-share-%s", share.Owner, timestamp)
	}

	// Build new share configuration
	shareConfig := buildShareConfig(shareName, share)

	// Append to content in memory
	newContent := string(content) + shareConfig

	// Write once
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write samba config: %v", err)
	}

	// Hot reload Samba configuration (no service restart)
	_ = ReloadSambaConfig() // Ignore error as service might not be running

	return shareName, nil
}

// DeleteShare deletes a Samba share
func (s *SambaService) DeleteShare(shareName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configPath := config.AppConfig.Samba.ConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read samba config: %v", err)
	}

	sharesToDelete := map[string]bool{shareName: true}
	sharesToUpdate := make(map[string]*types.Share)

	newContent, err := s.modifySharesInContent(string(content), sharesToDelete, sharesToUpdate)
	if err != nil {
		return fmt.Errorf("failed to delete share: %v", err)
	}

	// Check if share was found
	if newContent == string(content) {
		return fmt.Errorf("share '%s' not found", shareName)
	}

	// Write updated config
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write samba config: %v", err)
	}

	// Hot reload Samba configuration (no service restart)
	_ = ReloadSambaConfig()

	return nil
}

// ListShares lists all Samba shares (only user-share# shares)
func (s *SambaService) ListShares() ([]types.ShareResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listSharesInternal()
}

// listSharesInternal lists shares without acquiring lock (for internal use)
func (s *SambaService) listSharesInternal() ([]types.ShareResponse, error) {
	configPath := config.AppConfig.Samba.ConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read samba config: %v", err)
	}

	return s.parseSharesFromContent(string(content))
}

// parseSharesFromContent parses shares from config content string
func (s *SambaService) parseSharesFromContent(content string) ([]types.ShareResponse, error) {
	var shares []types.ShareResponse
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentShare *types.ShareResponse

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for share section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous share if exists and it matches pattern
			if currentShare != nil && sharePattern.MatchString(currentShare.ID) {
				shares = append(shares, *currentShare)
			}

			// Start new share
			shareName := strings.Trim(line, "[]")

			// Extract owner from share name (e.g., "alice-share-ProjectFiles" -> "alice")
			matches := sharePattern.FindStringSubmatch(shareName)
			var owner string
			if len(matches) > 1 {
				owner = matches[1]
			}

			currentShare = &types.ShareResponse{
				ID:         shareName,
				Owner:      owner,
				SharedWith: []string{},
			}
			continue
		}

		// Parse share properties
		if currentShare != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "path":
					currentShare.Path = value
					// Extract SubPath by removing owner's home directory from the full path
					ownerHome := filepath.Join(config.AppConfig.HomeDir, currentShare.Owner)
					if strings.HasPrefix(value, ownerHome) {
						subPath := strings.TrimPrefix(value, ownerHome)
						subPath = strings.TrimPrefix(subPath, "/")
						subPath = strings.TrimPrefix(subPath, "\\")
						if subPath != "" {
							currentShare.SubPath = subPath
						}
					}
				case "read only":
					currentShare.ReadOnly = (value == "yes")
				case "comment":
					currentShare.Comment = value
				case "valid users":
					// Parse valid users list (owner is not included, only shared_with users)
					users := strings.Fields(value)
					currentShare.SharedWith = users
				}
			}
		}
	}

	// Add last share if it matches pattern
	if currentShare != nil && sharePattern.MatchString(currentShare.ID) {
		shares = append(shares, *currentShare)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse samba config: %v", err)
	}

	return shares, nil
}

// buildShareConfigLines builds a share configuration as a slice of lines
func buildShareConfigLines(shareID string, share *types.Share) []string {
	ownerHome := filepath.Join(config.AppConfig.HomeDir, share.Owner)

	// If SubPath is specified, append it to the owner's home directory
	sharePath := ownerHome
	if share.SubPath != "" {
		// Clean and validate the subpath (ignoring errors since this is just for config generation)
		cleanedSubPath, _ := cleanSubPath(share.SubPath)
		if cleanedSubPath != "" {
			sharePath = filepath.Join(ownerHome, cleanedSubPath)
		}
	}

	validUsersStr := strings.Join(share.SharedWith, " ")

	lines := []string{fmt.Sprintf("[%s]", shareID)}
	lines = append(lines, fmt.Sprintf("   path = %s", sharePath))
	lines = append(lines, "   browseable = yes")
	lines = append(lines, fmt.Sprintf("   valid users = %s", validUsersStr))
	lines = append(lines, "   force user = root")
	lines = append(lines, "   force group = root")

	if share.ReadOnly {
		lines = append(lines, "   read only = yes")
	} else {
		lines = append(lines, "   read only = no")
		lines = append(lines, "   writable = yes")
	}

	if share.Comment != "" {
		lines = append(lines, fmt.Sprintf("   comment = %s", share.Comment))
	}

	return lines
}

// buildShareConfig builds a share configuration string (with leading newline)
func buildShareConfig(shareID string, share *types.Share) string {
	lines := buildShareConfigLines(shareID, share)
	return "\n" + strings.Join(lines, "\n") + "\n"
}

// writeShareToLines writes a share configuration to lines (without leading newline)
func writeShareToLines(shareID string, share *types.Share) []string {
	return buildShareConfigLines(shareID, share)
}

// modifySharesInContent modifies shares in config content (delete and update)
func (s *SambaService) modifySharesInContent(content string, sharesToDelete map[string]bool, sharesToUpdate map[string]*types.Share) (string, error) {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	var inTargetShare bool
	var currentShareID string
	var currentShareLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check if this is the start of a share section
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			// If we were in a share that needs updating, write it now
			if inTargetShare && currentShareID != "" && sharesToUpdate[currentShareID] != nil {
				// Write updated share
				updatedShare := sharesToUpdate[currentShareID]
				lines = append(lines, writeShareToLines(currentShareID, updatedShare)...)
			} else if !inTargetShare || !sharesToDelete[currentShareID] {
				// Write accumulated lines if not deleted
				lines = append(lines, currentShareLines...)
			}

			// Start new share
			shareName := strings.Trim(trimmedLine, "[]")
			currentShareID = shareName
			currentShareLines = []string{line}

			// Check if this share should be deleted or updated
			if sharesToDelete[shareName] {
				inTargetShare = true
			} else if sharesToUpdate[shareName] != nil {
				inTargetShare = true
			} else {
				inTargetShare = false
			}
			continue
		}

		// Accumulate lines for current share
		if inTargetShare {
			// Skip lines for shares to delete, accumulate for shares to update
			if sharesToUpdate[currentShareID] != nil {
				// We'll rebuild this share, so skip original lines
				continue
			}
			// For delete, we skip everything
		} else {
			currentShareLines = append(currentShareLines, line)
		}
	}

	// Handle last share
	if inTargetShare && currentShareID != "" && sharesToUpdate[currentShareID] != nil {
		updatedShare := sharesToUpdate[currentShareID]
		lines = append(lines, writeShareToLines(currentShareID, updatedShare)...)
	} else if !inTargetShare || !sharesToDelete[currentShareID] {
		lines = append(lines, currentShareLines...)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to parse config: %v", err)
	}

	return strings.Join(lines, "\n"), nil
}

// UpdateShare updates an existing Samba share by ID
func (s *SambaService) UpdateShare(shareId string, share *types.Share) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate share ID format (owner-share#)
	if !sharePattern.MatchString(shareId) {
		return fmt.Errorf("invalid share ID format")
	}

	// Validate shared_with usernames
	if len(share.SharedWith) == 0 {
		return fmt.Errorf("must share with at least one user")
	}
	for _, username := range share.SharedWith {
		if !isValidUsername(username) {
			return fmt.Errorf("invalid username in shared_with: %s", username)
		}
	}

	// Get owner's home directory
	ownerHome := filepath.Join(config.AppConfig.HomeDir, share.Owner)
	if _, err := os.Stat(ownerHome); os.IsNotExist(err) {
		return fmt.Errorf("owner's home directory does not exist")
	}

	// Validate and check subdirectory path if specified
	if share.SubPath != "" {
		cleanedSubPath, err := cleanSubPath(share.SubPath)
		if err != nil {
			return err
		}
		if cleanedSubPath != "" {
			fullSubPath := filepath.Join(ownerHome, cleanedSubPath)
			// Create the subdirectory if it doesn't exist
			if err := os.MkdirAll(fullSubPath, 0770); err != nil {
				return fmt.Errorf("failed to create subdirectory: %v", err)
			}
			// Set ownership to root:root
			cmd := exec.Command("chown", "-R", "root:root", fullSubPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set subdirectory ownership: %v, output: %s", err, output)
			}
			// Set permissions to 770
			cmd = exec.Command("chmod", "-R", "770", fullSubPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set subdirectory permissions: %v, output: %s", err, output)
			}
		}
	}

	// Read config once
	configPath := config.AppConfig.Samba.ConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read samba config: %v", err)
	}

	// Update in memory
	sharesToDelete := make(map[string]bool)
	sharesToUpdate := map[string]*types.Share{
		shareId: share,
	}

	newContent, err := s.modifySharesInContent(string(content), sharesToDelete, sharesToUpdate)
	if err != nil {
		return fmt.Errorf("failed to update share: %v", err)
	}

	// Write once
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write samba config: %v", err)
	}

	// Hot reload Samba configuration
	_ = ReloadSambaConfig()

	return nil
}

// ChangePassword changes the password for an existing Samba user
func (s *SambaService) ChangePassword(username string, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate username
	if !isValidUsername(username) {
		return fmt.Errorf("invalid username: must contain only letters, numbers, underscore, and dash")
	}

	// Verify user exists by checking if they're in pdbedit list
	cmd := exec.Command("pdbedit", "-L", "-u", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("user does not exist or failed to verify user: %v, output: %s", err, output)
	}

	// Change Samba password using smbpasswd (without -a flag for existing users)
	cmd = exec.Command("smbpasswd", "-s", username)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe for smbpasswd: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start smbpasswd: %v", err)
	}

	// smbpasswd expects password twice
	_, err = stdin.Write([]byte(newPassword + "\n" + newPassword + "\n"))
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to write new password: %v", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to change password: %v", err)
	}

	return nil
}

// ReloadSambaConfig reloads Samba configuration without restarting the service
// This is a hot reload that tells smbd to re-read its configuration file
func ReloadSambaConfig() error {
	cmd := exec.Command("smbcontrol", "smbd", "reload-config")
	return cmd.Run()
}
