// Package navigation provides festival-project linking and navigation state management.
package navigation

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FuzzyMatch represents a fuzzy match result
type FuzzyMatch struct {
	Path    string // Full path to the match
	Name    string // Display name
	Score   int    // Match score (higher is better)
	Indices []int  // Matched character positions (for highlighting)
}

// FuzzyTarget represents a target for fuzzy matching
type FuzzyTarget struct {
	Name string // Display name (used for matching)
	Path string // Full path (returned on match)
}

// FuzzyFinder provides fuzzy matching for festival navigation
type FuzzyFinder struct {
	targets   []FuzzyTarget // Available targets
	threshold int           // Minimum score threshold (0 = accept any)
}

// NewFuzzyFinder creates a finder for the given targets
func NewFuzzyFinder(targets []FuzzyTarget) *FuzzyFinder {
	return &FuzzyFinder{
		targets:   targets,
		threshold: 0,
	}
}

// WithThreshold sets the minimum score threshold
func (f *FuzzyFinder) WithThreshold(threshold int) *FuzzyFinder {
	f.threshold = threshold
	return f
}

// Find returns matches for the pattern, sorted by score descending
func (f *FuzzyFinder) Find(pattern string) []FuzzyMatch {
	// Handle multi-word patterns (AND logic)
	words := strings.Fields(pattern)
	if len(words) == 0 {
		return nil
	}

	// Score each target against all words
	var result []FuzzyMatch
	for _, target := range f.targets {
		// Target must match ALL words (AND logic)
		totalScore := 0
		var allIndices []int
		allMatch := true

		for _, word := range words {
			score, indices := Score(word, target.Name)
			if score == 0 {
				allMatch = false
				break
			}
			totalScore += score
			allIndices = append(allIndices, indices...)
		}

		if allMatch && totalScore >= f.threshold {
			result = append(result, FuzzyMatch{
				Path:    target.Path,
				Name:    target.Name,
				Score:   totalScore,
				Indices: allIndices,
			})
		}
	}

	// Sort by score descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result
}

// IsUnambiguous returns true if the top match is significantly better than alternatives
func IsUnambiguous(matches []FuzzyMatch) bool {
	if len(matches) <= 1 {
		return true
	}
	// Consider unambiguous if top score is 20% better than second
	threshold := float64(matches[0].Score) * 0.8
	return float64(matches[1].Score) < threshold
}

// CollectNavigationTargets gathers all possible navigation targets from a festivals directory
func CollectNavigationTargets(festivalsDir string) []FuzzyTarget {
	var targets []FuzzyTarget

	// Status directories to search
	statusDirs := []string{"active", "planned", "completed", "dungeon"}

	for _, status := range statusDirs {
		statusPath := filepath.Join(festivalsDir, status)
		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			festivalName := entry.Name()
			festivalPath := filepath.Join(statusPath, festivalName)

			// Add festival itself
			targets = append(targets, FuzzyTarget{
				Name: festivalName,
				Path: festivalPath,
			})

			// Add phases within festival
			phaseTargets := collectPhases(festivalPath, festivalName)
			targets = append(targets, phaseTargets...)
		}
	}

	return targets
}

// collectPhases gathers phases and sequences from a festival
func collectPhases(festivalPath, festivalName string) []FuzzyTarget {
	var targets []FuzzyTarget

	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip non-phase directories
		if !isPhaseDir(name) {
			continue
		}

		phasePath := filepath.Join(festivalPath, name)
		phaseDisplayName := festivalName + "/" + name

		targets = append(targets, FuzzyTarget{
			Name: phaseDisplayName,
			Path: phasePath,
		})

		// Also add phase name alone for simpler matching
		targets = append(targets, FuzzyTarget{
			Name: name,
			Path: phasePath,
		})

		// Add sequences within phase
		seqTargets := collectSequences(phasePath, festivalName, name)
		targets = append(targets, seqTargets...)
	}

	return targets
}

// collectSequences gathers sequences from a phase
func collectSequences(phasePath, festivalName, phaseName string) []FuzzyTarget {
	var targets []FuzzyTarget

	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip non-sequence directories
		if !isSequenceDir(name) {
			continue
		}

		seqPath := filepath.Join(phasePath, name)
		seqDisplayName := festivalName + "/" + phaseName + "/" + name

		targets = append(targets, FuzzyTarget{
			Name: seqDisplayName,
			Path: seqPath,
		})

		// Also add shorter names for simpler matching
		targets = append(targets, FuzzyTarget{
			Name: phaseName + "/" + name,
			Path: seqPath,
		})
	}

	return targets
}

// isPhaseDir checks if a directory name looks like a phase (NNN_name)
func isPhaseDir(name string) bool {
	if len(name) < 4 {
		return false
	}
	// Must start with 3 digits and underscore
	for i := 0; i < 3; i++ {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
	}
	return name[3] == '_'
}

// isSequenceDir checks if a directory name looks like a sequence (NN_name)
func isSequenceDir(name string) bool {
	if len(name) < 3 {
		return false
	}
	// Must start with 2 digits and underscore
	for i := 0; i < 2; i++ {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
	}
	return name[2] == '_'
}

// SortMatchesByScore sorts matches by score in descending order
func SortMatchesByScore(matches []FuzzyMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
}

// FormatMatchList formats matches for display in error messages
func FormatMatchList(matches []FuzzyMatch, limit int) []string {
	n := len(matches)
	if limit > 0 && n > limit {
		n = limit
	}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = matches[i].Name
	}
	return result
}
