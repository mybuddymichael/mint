package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// ResolveIssueID resolves a partial ID to a full ID
// Returns error if ambiguous or not found
func (s *Store) ResolveIssueID(partialID string) (string, error) {
	// Check exact match first
	if _, exists := s.Issues[partialID]; exists {
		return partialID, nil
	}

	// Check for prefix matches
	var matches []string
	for id := range s.Issues {
		if len(id) >= len(partialID) && id[:len(partialID)] == partialID {
			matches = append(matches, id)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("issue %s not found", partialID)
	}

	if len(matches) > 1 {
		sort.Strings(matches)
		return "", fmt.Errorf("ambiguous ID %s matches: %v", partialID, matches)
	}

	return matches[0], nil
}

// GetIssue retrieves an issue by ID or partial ID
func (s *Store) GetIssue(id string) (*Issue, error) {
	resolvedID, err := s.ResolveIssueID(id)
	if err != nil {
		return nil, err
	}
	return s.Issues[resolvedID], nil
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

// FormatID computes minimum unique prefix length across all store issues,
// then calls id.FormatID to apply ANSI formatting (underline + gray suffix)
func (s *Store) FormatID(id string) string {
	issues := s.ListIssues()
	ids := make([]string, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}
	uniqueLengths := MinUniquePrefixLengths(ids)
	return FormatID(id, uniqueLengths[id])
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

// CloseIssue closes an issue, optionally with a reason
func (s *Store) CloseIssue(id, reason string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return fmt.Errorf("issue not found: %s", id)
	}
	issue.Status = "closed"
	if reason != "" {
		issue.Comments = append(issue.Comments, fmt.Sprintf("Closed with reason: %s", reason))
	}
	return nil
}

// ReopenIssue reopens a closed issue
func (s *Store) ReopenIssue(id string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return fmt.Errorf("issue not found: %s", id)
	}
	issue.Status = "open"
	return nil
}

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
