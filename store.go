package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

// AddIssue creates a new issue with a unique ID and adds it to the store
func (s *Store) AddIssue(title string) (*Issue, error) {
	const maxRetries = 10

	var id string
	for i := 0; i < maxRetries; i++ {
		id = GenerateID(s.Prefix)
		if _, exists := s.Issues[id]; !exists {
			break
		}
		// Collision detected, retry
		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to generate unique ID after %d attempts", maxRetries)
		}
	}

	issue := &Issue{
		ID:     id,
		Title:  title,
		Status: "open",
	}

	s.Issues[id] = issue
	return issue, nil
}

// GetIssue retrieves an issue by ID
func (s *Store) GetIssue(id string) (*Issue, error) {
	issue, exists := s.Issues[id]
	if !exists {
		return nil, fmt.Errorf("issue %s not found", id)
	}
	return issue, nil
}

// ListIssues returns all issues sorted by ID
func (s *Store) ListIssues() []*Issue {
	issues := make([]*Issue, 0, len(s.Issues))
	for _, issue := range s.Issues {
		issues = append(issues, issue)
	}
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].ID < issues[j].ID
	})
	return issues
}

// UpdateIssueTitle updates an issue's title
func (s *Store) UpdateIssueTitle(id, title string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return err
	}
	issue.Title = title
	return nil
}

// AddDependency adds a dependency relationship (issue depends on dependsOnID)
func (s *Store) AddDependency(issueID, dependsOnID string) error {
	issue, err := s.GetIssue(issueID)
	if err != nil {
		return err
	}

	blocker, err := s.GetIssue(dependsOnID)
	if err != nil {
		return err
	}

	issue.DependsOn = append(issue.DependsOn, dependsOnID)
	blocker.Blocks = append(blocker.Blocks, issueID)
	return nil
}

// AddBlocker adds a blocks relationship (issue blocks blockedID)
func (s *Store) AddBlocker(issueID, blockedID string) error {
	issue, err := s.GetIssue(issueID)
	if err != nil {
		return err
	}

	blocked, err := s.GetIssue(blockedID)
	if err != nil {
		return err
	}

	issue.Blocks = append(issue.Blocks, blockedID)
	blocked.DependsOn = append(blocked.DependsOn, issueID)
	return nil
}

// AddComment adds a comment to an issue
func (s *Store) AddComment(id, comment string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return err
	}
	issue.Comments = append(issue.Comments, comment)
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
