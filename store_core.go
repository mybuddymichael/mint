package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

// Store represents the mint issue store
type Store struct {
	Prefix string            `yaml:"prefix"`
	Issues map[string]*Issue `yaml:"issues"`
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
		Prefix: "mint",
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
