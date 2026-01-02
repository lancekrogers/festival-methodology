// Package gates provides template resolution with hierarchical precedence.
package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ResolveResult contains the resolved template path and its source
type ResolveResult struct {
	Path   string      // Full path to the template file
	Level  PolicyLevel // Hierarchy level where found
	Exists bool        // Whether the template was found
}

// TemplateNotFoundError is returned when a template cannot be found at any level
type TemplateNotFoundError struct {
	TemplateID string   // Template that was searched for
	Searched   []string // Paths that were searched
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template %q not found in: %s", e.TemplateID, strings.Join(e.Searched, ", "))
}

// TemplateResolver finds gate templates with hierarchical precedence
type TemplateResolver struct {
	festivalsRoot string
	cache         map[string]*ResolveResult
	mu            sync.RWMutex
}

// NewTemplateResolver creates a resolver for the given festivals root
func NewTemplateResolver(festivalsRoot string) *TemplateResolver {
	return &TemplateResolver{
		festivalsRoot: festivalsRoot,
		cache:         make(map[string]*ResolveResult),
	}
}

// gatesPrefixInfo holds parsed information about a template ID with optional gates/ prefix
type gatesPrefixInfo struct {
	originalID string // Original template ID (e.g., "gates/QUALITY_GATE_TESTING")
	resolvedID string // Template name without prefix (e.g., "QUALITY_GATE_TESTING")
	hasPrefix  bool   // Whether the original had "gates/" prefix
	gatesPath  string // Path to festival's gates/ directory (empty if no prefix or no festivalPath)
}

// parseGatesPrefix extracts gates/ prefix info from a template ID.
// If templateID starts with "gates/" and festivalPath is provided,
// returns info about where to look for the template.
func parseGatesPrefix(templateID, festivalPath string) gatesPrefixInfo {
	const gatesPrefix = "gates/"

	info := gatesPrefixInfo{
		originalID: templateID,
		resolvedID: templateID,
		hasPrefix:  strings.HasPrefix(templateID, gatesPrefix),
	}

	if info.hasPrefix {
		info.resolvedID = strings.TrimPrefix(templateID, gatesPrefix)
		if festivalPath != "" {
			info.gatesPath = filepath.Join(festivalPath, "gates", info.resolvedID+".md")
		}
	}

	return info
}

// Resolve finds a template by ID, searching the hierarchy from most to least specific.
// Search order: sequence → phase → festival → global gates
// If templateID starts with "gates/", it first looks in the festival's gates/ directory.
func (r *TemplateResolver) Resolve(
	templateID string,
	festivalPath, phasePath, sequencePath string,
) (*ResolveResult, error) {
	// Check cache first
	cacheKey := r.cacheKey(templateID, sequencePath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	// Parse gates/ prefix
	gatesInfo := parseGatesPrefix(templateID, festivalPath)

	// Build search paths from most to least specific
	searchPaths := r.buildSearchPaths(gatesInfo, festivalPath, phasePath, sequencePath)

	var searched []string
	for _, sp := range searchPaths {
		// Skip empty paths
		if sp.path == "" || sp.path == filepath.Join("", ".fest.templates", gatesInfo.resolvedID+".md") {
			continue
		}

		searched = append(searched, sp.path)
		if _, err := os.Stat(sp.path); err == nil {
			result := &ResolveResult{
				Path:   sp.path,
				Level:  sp.level,
				Exists: true,
			}
			r.setCache(cacheKey, result)
			return result, nil
		}
	}

	// Template not found
	result := &ResolveResult{Exists: false}
	return result, &TemplateNotFoundError{
		TemplateID: templateID,
		Searched:   searched,
	}
}

// searchPath represents a template search location with its hierarchy level
type searchPath struct {
	level PolicyLevel
	path  string
}

// buildSearchPaths creates the ordered list of paths to search for a template
func (r *TemplateResolver) buildSearchPaths(
	gatesInfo gatesPrefixInfo,
	festivalPath, phasePath, sequencePath string,
) []searchPath {
	var paths []searchPath

	// If gates/ prefix, prioritize festival's gates/ directory
	if gatesInfo.hasPrefix && gatesInfo.gatesPath != "" {
		paths = append(paths, searchPath{PolicyLevelFestival, gatesInfo.gatesPath})
	}

	// Standard hierarchy paths
	paths = append(paths,
		searchPath{PolicyLevelSequence, filepath.Join(sequencePath, ".fest.templates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelPhase, filepath.Join(phasePath, ".fest.templates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
	)

	return paths
}

// ResolveForPhase finds a template, stopping at phase level (no sequence search)
// If templateID starts with "gates/", it first looks in the festival's gates/ directory.
func (r *TemplateResolver) ResolveForPhase(
	templateID string,
	festivalPath, phasePath string,
) (*ResolveResult, error) {
	cacheKey := r.cacheKey(templateID, phasePath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	// Parse gates/ prefix
	gatesInfo := parseGatesPrefix(templateID, festivalPath)

	// Build search paths (phase level and below)
	var paths []searchPath
	if gatesInfo.hasPrefix && gatesInfo.gatesPath != "" {
		paths = append(paths, searchPath{PolicyLevelFestival, gatesInfo.gatesPath})
	}
	paths = append(paths,
		searchPath{PolicyLevelPhase, filepath.Join(phasePath, ".fest.templates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
	)

	var searched []string
	for _, sp := range paths {
		if sp.path == "" {
			continue
		}
		searched = append(searched, sp.path)
		if _, err := os.Stat(sp.path); err == nil {
			result := &ResolveResult{
				Path:   sp.path,
				Level:  sp.level,
				Exists: true,
			}
			r.setCache(cacheKey, result)
			return result, nil
		}
	}

	return &ResolveResult{Exists: false}, &TemplateNotFoundError{
		TemplateID: templateID,
		Searched:   searched,
	}
}

// ResolveForFestival finds a template, stopping at festival level
// If templateID starts with "gates/", it first looks in the festival's gates/ directory.
func (r *TemplateResolver) ResolveForFestival(
	templateID string,
	festivalPath string,
) (*ResolveResult, error) {
	cacheKey := r.cacheKey(templateID, festivalPath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	// Parse gates/ prefix
	gatesInfo := parseGatesPrefix(templateID, festivalPath)

	// Build search paths (festival level and below)
	var paths []searchPath
	if gatesInfo.hasPrefix && gatesInfo.gatesPath != "" {
		paths = append(paths, searchPath{PolicyLevelFestival, gatesInfo.gatesPath})
	}
	paths = append(paths,
		searchPath{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
		searchPath{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "templates", "gates", gatesInfo.resolvedID+".md")},
	)

	var searched []string
	for _, sp := range paths {
		if sp.path == "" {
			continue
		}
		searched = append(searched, sp.path)
		if _, err := os.Stat(sp.path); err == nil {
			result := &ResolveResult{
				Path:   sp.path,
				Level:  sp.level,
				Exists: true,
			}
			r.setCache(cacheKey, result)
			return result, nil
		}
	}

	return &ResolveResult{Exists: false}, &TemplateNotFoundError{
		TemplateID: templateID,
		Searched:   searched,
	}
}

// ClearCache clears the template cache
func (r *TemplateResolver) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]*ResolveResult)
}

// cacheKey creates a unique cache key for a template lookup
func (r *TemplateResolver) cacheKey(templateID, contextPath string) string {
	return templateID + ":" + contextPath
}

// getCached retrieves a cached result
func (r *TemplateResolver) getCached(key string) *ResolveResult {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache[key]
}

// setCache stores a result in the cache
func (r *TemplateResolver) setCache(key string, result *ResolveResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[key] = result
}

// FestivalsRoot returns the configured festivals root path
func (r *TemplateResolver) FestivalsRoot() string {
	return r.festivalsRoot
}
