package main

import (
	"os"
	"path/filepath"
	"strings"
)

// SetPrefix changes the store prefix and updates all issue IDs
func (s *Store) SetPrefix(newPrefix string) error {
	oldPrefix := s.Prefix
	newPrefix = strings.TrimSuffix(newPrefix, "-")

	// Build mapping of old ID to new ID
	idMap := make(map[string]string)
	for oldID := range s.Issues {
		// Extract suffix after old prefix, stripping any leading hyphen
		suffix := strings.TrimPrefix(oldID[len(oldPrefix):], "-")
		newID := newPrefix + "-" + suffix
		idMap[oldID] = newID
	}

	// Create new issues map with updated IDs
	newIssues := make(map[string]*Issue)
	for oldID, issue := range s.Issues {
		newID := idMap[oldID]
		issue.ID = newID

		// Update DependsOn references
		for i, depID := range issue.DependsOn {
			if newDepID, ok := idMap[depID]; ok {
				issue.DependsOn[i] = newDepID
			}
		}

		// Update Blocks references
		for i, blockID := range issue.Blocks {
			if newBlockID, ok := idMap[blockID]; ok {
				issue.Blocks[i] = newBlockID
			}
		}

		newIssues[newID] = issue
	}

	s.Prefix = newPrefix
	s.Issues = newIssues
	return nil
}

// GetStoreFilePath returns the path to the mint-issues.yaml file
// Checks MINT_STORE_FILE env var first (for tests)
// Then looks for .git walking up from current directory
// Falls back to current directory if no .git found
func GetStoreFilePath() (string, error) {
	// Check env var first (for tests)
	if envPath := os.Getenv("MINT_STORE_FILE"); envPath != "" {
		return envPath, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up looking for .git
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return filepath.Join(dir, "mint-issues.yaml"), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return filepath.Join(cwd, "mint-issues.yaml"), nil
}
