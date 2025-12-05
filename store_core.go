package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

// Store represents the mint issue store
type Store struct {
	SchemaVersion int               `yaml:"schema_version"`
	Prefix        string            `yaml:"prefix"`
	Issues        map[string]*Issue `yaml:"issues"`
}

// Issue represents a single issue
type Issue struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	DependsOn []string `yaml:"depends_on,omitempty"`
	Blocks    []string `yaml:"blocks,omitempty"`
	Comments  []string `yaml:"comments,omitempty"`
}

// NewStore creates a new store with defaults
func NewStore() *Store {
	return &Store{
		SchemaVersion: CurrentSchemaVersion,
		Prefix:        "mint",
		Issues:        make(map[string]*Issue),
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

	// Use zero-value Store instead of NewStore() so missing fields
	// (like schema_version in v0 files) default to 0, triggering migrations
	store := &Store{}
	if err := yaml.Unmarshal(data, store); err != nil {
		return nil, err
	}

	// Initialize maps if nil (for empty files)
	if store.Issues == nil {
		store.Issues = make(map[string]*Issue)
	}

	// Run migrations if needed
	migrated, err := RunMigrations(store)
	if err != nil {
		return nil, err
	}

	// Save and notify if migrated
	if migrated {
		if err := store.Save(filePath); err != nil {
			return nil, err
		}
		println("\x1b[1;33mYou're using a new version of mint. Issue file updated.\x1b[0m\n")
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
