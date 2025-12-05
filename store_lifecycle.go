package main

import "fmt"

// AddComment adds a comment to an issue
func (s *Store) AddComment(id, comment string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return err
	}
	issue.Comments = append(issue.Comments, comment)
	s.touch(issue)
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
	s.touch(issue)
	return nil
}

// ReopenIssue reopens a closed issue
func (s *Store) ReopenIssue(id string) error {
	issue, err := s.GetIssue(id)
	if err != nil {
		return fmt.Errorf("issue not found: %s", id)
	}
	issue.Status = "open"
	s.touch(issue)
	return nil
}

// DeleteIssue deletes an issue and cleans up all references to it
func (s *Store) DeleteIssue(id string) error {
	fullID, err := s.ResolveIssueID(id)
	if err != nil {
		return fmt.Errorf("issue not found: %s", id)
	}

	issueToDelete := s.Issues[fullID]

	// Remove from issues that depend on this one
	for _, blockedID := range issueToDelete.Blocks {
		if blocked := s.Issues[blockedID]; blocked != nil {
			newDeps := make([]string, 0, len(blocked.DependsOn)-1)
			for _, depID := range blocked.DependsOn {
				if depID != fullID {
					newDeps = append(newDeps, depID)
				}
			}
			blocked.DependsOn = newDeps
		}
	}

	// Remove from issues this one depends on
	for _, depID := range issueToDelete.DependsOn {
		if dep := s.Issues[depID]; dep != nil {
			newBlocks := make([]string, 0, len(dep.Blocks)-1)
			for _, blockID := range dep.Blocks {
				if blockID != fullID {
					newBlocks = append(newBlocks, blockID)
				}
			}
			dep.Blocks = newBlocks
		}
	}

	delete(s.Issues, fullID)
	return nil
}
