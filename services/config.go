package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/types"
)

// ConfigService handles Samba configuration file management
type ConfigService struct{}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// GetSambaConfig reads and parses Samba configuration
func (s *ConfigService) GetSambaConfig() (*types.SambaConfigResponse, error) {
	smbConfPath := config.AppConfig.Samba.ConfigPath
	if smbConfPath == "" {
		smbConfPath = "/etc/samba/smb.conf"
	}

	file, err := os.Open(smbConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open smb.conf: %w", err)
	}
	defer file.Close()

	globalConfig := types.SambaGlobalConfig{}
	homesConfig := types.SambaHomesConfig{}

	var currentSection string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "global":
			s.parseGlobalConfig(&globalConfig, key, value)
		case "homes":
			s.parseHomesConfig(&homesConfig, key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read smb.conf: %w", err)
	}

	return &types.SambaConfigResponse{
		Global: globalConfig,
		Homes:  homesConfig,
	}, nil
}

func (s *ConfigService) parseGlobalConfig(cfg *types.SambaGlobalConfig, key, value string) {
	key = strings.ToLower(strings.ReplaceAll(key, " ", ""))
	switch key {
	case "workgroup":
		cfg.Workgroup = value
	case "serverstring":
		cfg.ServerString = value
	case "security":
		cfg.Security = value
	case "passdbbackend":
		cfg.PassdbBackend = value
	case "maptoguest":
		cfg.MapToGuest = value
	case "accessbasedshareenum":
		cfg.AccessBasedShareEnum = value
	}
}

func (s *ConfigService) parseHomesConfig(cfg *types.SambaHomesConfig, key, value string) {
	key = strings.ToLower(strings.ReplaceAll(key, " ", ""))
	switch key {
	case "comment":
		cfg.Comment = value
	case "browseable", "browsable":
		cfg.Browseable = value
	case "writable", "writeable":
		cfg.Writable = value
	case "validusers":
		cfg.ValidUsers = value
	case "forceuser":
		cfg.ForceUser = value
	case "forcegroup":
		cfg.ForceGroup = value
	case "createmask", "createmode":
		cfg.CreateMask = value
	case "directorymask", "directorymode":
		cfg.DirectoryMask = value
	}
}

// UpdateSambaConfig updates Samba configuration file
func (s *ConfigService) UpdateSambaConfig(req *types.UpdateSambaConfigRequest) error {
	smbConfPath := config.AppConfig.Samba.ConfigPath
	if smbConfPath == "" {
		smbConfPath = "/etc/samba/smb.conf"
	}

	// Read current config
	file, err := os.Open(smbConfPath)
	if err != nil {
		return fmt.Errorf("failed to open smb.conf: %w", err)
	}

	var lines []string
	var currentSection string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for section headers
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentSection = strings.ToLower(strings.Trim(trimmed, "[]"))
			lines = append(lines, line)
			continue
		}

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			lines = append(lines, line)
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			lines = append(lines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		keyLower := strings.ToLower(strings.ReplaceAll(key, " ", ""))

		replaced := false

		// Update global section
		if currentSection == "global" && req.Global != nil {
			if newValue, ok := s.getGlobalValue(req.Global, keyLower); ok {
				lines = append(lines, fmt.Sprintf("   %s = %s", key, newValue))
				replaced = true
			}
		}

		// Update homes section
		if currentSection == "homes" && req.Homes != nil {
			if newValue, ok := s.getHomesValue(req.Homes, keyLower); ok {
				lines = append(lines, fmt.Sprintf("   %s = %s", key, newValue))
				replaced = true
			}
		}

		if !replaced {
			lines = append(lines, line)
		}
	}

	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read smb.conf: %w", err)
	}

	// Write updated config
	output := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(smbConfPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write smb.conf: %w", err)
	}

	// Hot reload Samba configuration
	_ = ReloadSambaConfig()

	return nil
}

func (s *ConfigService) getGlobalValue(cfg *types.SambaGlobalConfig, key string) (string, bool) {
	switch key {
	case "workgroup":
		if cfg.Workgroup != "" {
			return cfg.Workgroup, true
		}
	case "serverstring":
		if cfg.ServerString != "" {
			return cfg.ServerString, true
		}
	case "security":
		if cfg.Security != "" {
			return cfg.Security, true
		}
	case "passdbbackend":
		if cfg.PassdbBackend != "" {
			return cfg.PassdbBackend, true
		}
	case "maptoguest":
		if cfg.MapToGuest != "" {
			return cfg.MapToGuest, true
		}
	case "accessbasedshareenum":
		if cfg.AccessBasedShareEnum != "" {
			return cfg.AccessBasedShareEnum, true
		}
	}
	return "", false
}

func (s *ConfigService) getHomesValue(cfg *types.SambaHomesConfig, key string) (string, bool) {
	switch key {
	case "comment":
		if cfg.Comment != "" {
			return cfg.Comment, true
		}
	case "browseable", "browsable":
		if cfg.Browseable != "" {
			return cfg.Browseable, true
		}
	case "writable", "writeable":
		if cfg.Writable != "" {
			return cfg.Writable, true
		}
	case "validusers":
		if cfg.ValidUsers != "" {
			return cfg.ValidUsers, true
		}
	case "forceuser":
		if cfg.ForceUser != "" {
			return cfg.ForceUser, true
		}
	case "forcegroup":
		if cfg.ForceGroup != "" {
			return cfg.ForceGroup, true
		}
	case "createmask", "createmode":
		if cfg.CreateMask != "" {
			return cfg.CreateMask, true
		}
	case "directorymask", "directorymode":
		if cfg.DirectoryMask != "" {
			return cfg.DirectoryMask, true
		}
	}
	return "", false
}

// ValidateSambaConfig validates Samba configuration using testparm
func ValidateSambaConfig(configPath string) error {
	cmd := exec.Command("testparm", "-s", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("testparm validation failed: %s (error: %w)", string(output), err)
	}
	return nil
}

// GetSambaConfigFile reads the raw smb.conf file
func (s *ConfigService) GetSambaConfigFile() (*types.SambaConfigFileResponse, error) {
	smbConfPath := config.AppConfig.Samba.ConfigPath
	if smbConfPath == "" {
		smbConfPath = "/etc/samba/smb.conf"
	}

	content, err := os.ReadFile(smbConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read smb.conf: %w", err)
	}

	return &types.SambaConfigFileResponse{
		Content: string(content),
		Path:    smbConfPath,
	}, nil
}

// UpdateSambaConfigFile writes the raw smb.conf file
func (s *ConfigService) UpdateSambaConfigFile(req *types.UpdateSambaConfigFileRequest) error {
	smbConfPath := config.AppConfig.Samba.ConfigPath
	if smbConfPath == "" {
		smbConfPath = "/etc/samba/smb.conf"
	}

	// Backup the original config
	backupPath := smbConfPath + ".backup"
	originalContent, err := os.ReadFile(smbConfPath)
	if err != nil {
		return fmt.Errorf("failed to read original smb.conf: %w", err)
	}

	if err := os.WriteFile(backupPath, originalContent, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write the new content
	if err := os.WriteFile(smbConfPath, []byte(req.Content), 0644); err != nil {
		return fmt.Errorf("failed to write smb.conf: %w", err)
	}

	// Validate the new configuration using testparm
	if err := ValidateSambaConfig(smbConfPath); err != nil {
		// Revert to backup if validation fails
		if revertErr := os.WriteFile(smbConfPath, originalContent, 0644); revertErr != nil {
			return fmt.Errorf("validation failed and revert failed: %v (revert error: %v)", err, revertErr)
		}
		// Remove the backup file after successful revert
		os.Remove(backupPath)
		return fmt.Errorf("configuration validation failed, reverted to original: %w", err)
	}

	// Remove backup file after successful validation
	os.Remove(backupPath)

	// Hot reload Samba configuration
	_ = ReloadSambaConfig()

	return nil
}
