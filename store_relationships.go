package main

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
	s.touch(issue, blocker)
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
	s.touch(issue, blocked)
	return nil
}

// RemoveDependency removes a dependency relationship (issue depends on dependsOnID)
func (s *Store) RemoveDependency(issueID, dependsOnID string) error {
	issue, err := s.GetIssue(issueID)
	if err != nil {
		return err
	}

	blocker, err := s.GetIssue(dependsOnID)
	if err != nil {
		return err
	}

	// Remove dependsOnID from issue.DependsOn
	newDeps := make([]string, 0, len(issue.DependsOn)-1)
	for _, depID := range issue.DependsOn {
		if depID != dependsOnID {
			newDeps = append(newDeps, depID)
		}
	}
	issue.DependsOn = newDeps

	// Remove issueID from blocker.Blocks
	newBlocks := make([]string, 0, len(blocker.Blocks)-1)
	for _, blockID := range blocker.Blocks {
		if blockID != issueID {
			newBlocks = append(newBlocks, blockID)
		}
	}
	blocker.Blocks = newBlocks

	s.touch(issue, blocker)
	return nil
}

// RemoveBlocker removes a blocks relationship (issue blocks blockedID)
func (s *Store) RemoveBlocker(issueID, blockedID string) error {
	issue, err := s.GetIssue(issueID)
	if err != nil {
		return err
	}

	blocked, err := s.GetIssue(blockedID)
	if err != nil {
		return err
	}

	// Remove blockedID from issue.Blocks
	newBlocks := make([]string, 0, len(issue.Blocks)-1)
	for _, blockID := range issue.Blocks {
		if blockID != blockedID {
			newBlocks = append(newBlocks, blockID)
		}
	}
	issue.Blocks = newBlocks

	// Remove issueID from blocked.DependsOn
	newDeps := make([]string, 0, len(blocked.DependsOn)-1)
	for _, depID := range blocked.DependsOn {
		if depID != issueID {
			newDeps = append(newDeps, depID)
		}
	}
	blocked.DependsOn = newDeps

	s.touch(issue, blocked)
	return nil
}
