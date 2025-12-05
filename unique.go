package main

import "sort"

// MinUniquePrefixLengths returns the minimum prefix length needed
// to uniquely identify each ID in the list
// Uses sort + adjacent comparison: O(n log n) instead of O(n²)
func MinUniquePrefixLengths(ids []string) map[string]int {
	result := make(map[string]int)
	if len(ids) == 0 {
		return result
	}

	// Sort IDs - after sorting, only need to compare adjacent pairs
	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)

	// For each ID, min prefix = max(commonPrefix with prev, commonPrefix with next) + 1
	for i, id := range sorted {
		minLen := 1

		// Compare with previous
		if i > 0 {
			commonLen := commonPrefixLen(sorted[i-1], id)
			if commonLen+1 > minLen {
				minLen = commonLen + 1
			}
		}

		// Compare with next
		if i < len(sorted)-1 {
			commonLen := commonPrefixLen(id, sorted[i+1])
			if commonLen+1 > minLen {
				minLen = commonLen + 1
			}
		}

		result[id] = minLen
	}

	return result
}

func commonPrefixLen(a, b string) int {
	maxCheck := min(len(a), len(b))
	for i := range maxCheck {
		if a[i] != b[i] {
			return i
		}
	}
	return maxCheck
}
