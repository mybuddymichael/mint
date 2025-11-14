package main

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Store represents the mint issue store
type Store struct {
	Prefix string            `yaml:"prefix"`
	Issues map[string]*Issue `yaml:"issues"`
}

// Issue represents a single issue
type Issue struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Status string `yaml:"status"`
}

// NewStore creates a new store with defaults
func NewStore() *Store {
	return &Store{
		Prefix: "mt-",
		Issues: make(map[string]*Issue),
	}
}

// LoadStore loads a store from a YAML file
// If the file doesn't exist, returns a new store with defaults
func LoadStore(filePath string) (*Store, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewStore(), nil
		}
		return nil, err
	}

	store := NewStore()
	if err := yaml.Unmarshal(data, store); err != nil {
		return nil, err
	}

	return store, nil
}

// Save saves the store to a YAML file
func (s *Store) Save(filePath string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
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
