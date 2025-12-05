package main

import (
	"fmt"
	"sort"
	"time"
)

// AddIssue creates a new issue with a unique ID and adds it to the store
func (s *Store) AddIssue(title string) (*Issue, error) {
	const maxRetries = 10

	length := CalculateIDLength(len(s.Issues))

	var id string
	for i := range maxRetries {
		id = GenerateID(s.Prefix, length)
		if _, exists := s.Issues[id]; !exists {
			break
		}
		// Collision detected, retry
		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to generate unique ID after %d attempts", maxRetries)
		}
	}

	now := time.Now()
	issue := &Issue{
		ID:        id,
		Title:     title,
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
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

// IsReady returns true if the issue has no open dependencies
func (s *Store) IsReady(issue *Issue) bool {
	for _, depID := range issue.DependsOn {
		dep := s.Issues[depID]
		if dep != nil && dep.Status == "open" {
			return false
		}
	}
	return true
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
	s.touch(issue)
	return nil
}
