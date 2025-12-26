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

// Resolve finds a template by ID, searching the hierarchy from most to least specific.
// Search order: sequence → phase → festival → global gates → built-in
func (r *TemplateResolver) Resolve(
	templateID string,
	festivalPath, phasePath, sequencePath string,
) (*ResolveResult, error) {
	// Check cache first
	cacheKey := r.cacheKey(templateID, sequencePath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	// Build search paths from most to least specific
	searchPaths := []struct {
		level PolicyLevel
		path  string
	}{
		// 1. Sequence level
		{PolicyLevelSequence, filepath.Join(sequencePath, ".fest.templates", templateID+".md")},

		// 2. Phase level
		{PolicyLevelPhase, filepath.Join(phasePath, ".fest.templates", templateID+".md")},

		// 3. Festival level
		{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", templateID+".md")},

		// 4. Global gates level
		{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "gates", "templates", templateID+".md")},

		// 5. Built-in templates
		{PolicyLevelBuiltin, filepath.Join(r.festivalsRoot, ".festival", "templates", templateID+".md")},
	}

	var searched []string
	for _, sp := range searchPaths {
		// Skip empty paths
		if sp.path == "" || sp.path == filepath.Join("", ".fest.templates", templateID+".md") {
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

// ResolveForPhase finds a template, stopping at phase level (no sequence search)
func (r *TemplateResolver) ResolveForPhase(
	templateID string,
	festivalPath, phasePath string,
) (*ResolveResult, error) {
	cacheKey := r.cacheKey(templateID, phasePath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	searchPaths := []struct {
		level PolicyLevel
		path  string
	}{
		{PolicyLevelPhase, filepath.Join(phasePath, ".fest.templates", templateID+".md")},
		{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", templateID+".md")},
		{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "gates", "templates", templateID+".md")},
		{PolicyLevelBuiltin, filepath.Join(r.festivalsRoot, ".festival", "templates", templateID+".md")},
	}

	var searched []string
	for _, sp := range searchPaths {
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
func (r *TemplateResolver) ResolveForFestival(
	templateID string,
	festivalPath string,
) (*ResolveResult, error) {
	cacheKey := r.cacheKey(templateID, festivalPath)
	if cached := r.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	searchPaths := []struct {
		level PolicyLevel
		path  string
	}{
		{PolicyLevelFestival, filepath.Join(festivalPath, ".festival", "templates", templateID+".md")},
		{PolicyLevelGlobal, filepath.Join(r.festivalsRoot, ".festival", "gates", "templates", templateID+".md")},
		{PolicyLevelBuiltin, filepath.Join(r.festivalsRoot, ".festival", "templates", templateID+".md")},
	}

	var searched []string
	for _, sp := range searchPaths {
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
