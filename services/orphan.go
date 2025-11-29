package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/types"
)

// ListOrphanedDirectories lists home directories that don't have corresponding Samba users
func (s *SambaService) ListOrphanedDirectories() ([]types.OrphanedDirectory, error) {
	// Get all Samba users
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}

	// Create a map of existing usernames
	userMap := make(map[string]bool)
	for _, user := range users {
		userMap[user.Username] = true
	}

	// List all directories in home directory
	homeDir := config.AppConfig.HomeDir
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return nil, err
	}

	var orphanedDirs []types.OrphanedDirectory
	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			// If directory name doesn't match any existing user, it's orphaned
			if !userMap[dirName] {
				dirPath := filepath.Join(homeDir, dirName)

				// Get directory size
				var size int64
				err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						// Skip files/dirs that can't be accessed
						return nil
					}
					if !info.IsDir() {
						size += info.Size()
					}
					return nil
				})
				// If walk fails, set size to 0 and continue
				if err != nil {
					size = 0
				}

				orphanedDirs = append(orphanedDirs, types.OrphanedDirectory{
					Name: dirName,
					Path: dirPath,
					Size: size,
				})
			}
		}
	}

	return orphanedDirs, nil
}

// DeleteOrphanedDirectory deletes an orphaned home directory
func (s *SambaService) DeleteOrphanedDirectory(dirName string) error {
	// Ensure the directory is actually orphaned
	users, err := s.ListUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Username == dirName {
			return fmt.Errorf("cannot delete directory: user '%s' still exists", dirName)
		}
	}

	// Delete the directory
	dirPath := filepath.Join(config.AppConfig.HomeDir, dirName)
	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	return nil
}
