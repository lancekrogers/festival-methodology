package template

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// Context holds all variables available for template rendering
type Context struct {
	// Automatic variables (fest provides)
	Auto map[string]interface{}

	// User-provided variables (from forms/CLI)
	User map[string]interface{}

	// Computed variables (derived from other values)
	Computed map[string]interface{}
}

// ContextBuilder builds variable contexts for template rendering
type ContextBuilder interface {
	WithUser(key string, value interface{}) ContextBuilder
	WithUserMap(values map[string]interface{}) ContextBuilder
	Build() *Context
	BuildForFestival(festivalName, goal string) *Context
	BuildForPhase(festivalName string, phaseNumber int, phaseName string) *Context
}

type contextBuilderImpl struct {
	userVars map[string]interface{}
}

// NewContextBuilder creates a new context builder
func NewContextBuilder() ContextBuilder {
	return &contextBuilderImpl{
		userVars: make(map[string]interface{}),
	}
}

// WithUser adds a user-provided variable
func (b *contextBuilderImpl) WithUser(key string, value interface{}) ContextBuilder {
	b.userVars[key] = value
	return b
}

// WithUserMap adds multiple user-provided variables
func (b *contextBuilderImpl) WithUserMap(values map[string]interface{}) ContextBuilder {
	for k, v := range values {
		b.userVars[k] = v
	}
	return b
}

// Build creates a context with automatic variables
func (b *contextBuilderImpl) Build() *Context {
	ctx := &Context{
		Auto:     buildAutomaticVariables(),
		User:     b.userVars,
		Computed: make(map[string]interface{}),
	}

	return ctx
}

// BuildForFestival creates a context for festival creation
func (b *contextBuilderImpl) BuildForFestival(festivalName, goal string) *Context {
	// Add festival-specific user variables
	b.WithUser("festival_name", festivalName)
	b.WithUser("festival_goal", goal)

	ctx := b.Build()

	// Add computed variables
	ctx.Computed["festival_id"] = generateFestivalID(festivalName)
	ctx.Computed["festival_path"] = festivalName

	return ctx
}

// BuildForPhase creates a context for phase creation
func (b *contextBuilderImpl) BuildForPhase(festivalName string, phaseNumber int, phaseName string) *Context {
	// Add phase-specific user variables
	b.WithUser("festival_name", festivalName)
	b.WithUser("phase_number", phaseNumber)
	b.WithUser("phase_name", phaseName)

	ctx := b.Build()

	// Add computed variables
	ctx.Computed["phase_id"] = fmt.Sprintf("%03d_%s", phaseNumber, phaseName)
	ctx.Computed["phase_path"] = filepath.Join(festivalName, fmt.Sprintf("%03d_%s", phaseNumber, phaseName))

	return ctx
}

// buildAutomaticVariables creates the automatic variables available in all templates
func buildAutomaticVariables() map[string]interface{} {
	auto := make(map[string]interface{})

	// Time-based variables
	now := time.Now()
	auto["now"] = now.Format(time.RFC3339)
	auto["today"] = now.Format("2006-01-02")
	auto["current_year"] = now.Year()
	auto["current_month"] = now.Format("January")
	auto["current_date"] = now.Format("January 2, 2006")
	auto["created_date"] = now.Format("2006-01-02")
	auto["timestamp"] = now.Unix()

	// User information
	if currentUser, err := user.Current(); err == nil {
		auto["user_name"] = currentUser.Username
		auto["user_home"] = currentUser.HomeDir

		// Try to get display name from full name
		if currentUser.Name != "" {
			auto["user_full_name"] = currentUser.Name
		}
	}

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		auto["hostname"] = hostname
	}

	// Fest version (placeholder - would be injected at build time)
	auto["fest_version"] = "2.0.0-dev"

	// Working directory
	if wd, err := os.Getwd(); err == nil {
		auto["working_dir"] = wd
		auto["working_dir_name"] = filepath.Base(wd)
	}

	return auto
}

// generateFestivalID creates a unique festival ID from the name
func generateFestivalID(name string) string {
	// Convert to lowercase, replace spaces/underscores with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")

	// Add timestamp for uniqueness
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", id, timestamp)
}

// ToFlatMap converts a Context to a flat map for template rendering
func (c *Context) ToFlatMap() map[string]interface{} {
	flat := make(map[string]interface{})

	// Copy automatic variables
	for k, v := range c.Auto {
		flat[k] = v
	}

	// Copy user variables (override auto if conflicts)
	for k, v := range c.User {
		flat[k] = v
	}

	// Copy computed variables (override all if conflicts)
	for k, v := range c.Computed {
		flat[k] = v
	}

	return flat
}

// Get retrieves a variable value by key, searching in order: Computed, User, Auto
func (c *Context) Get(key string) (interface{}, bool) {
	// Check computed first (highest priority)
	if val, ok := c.Computed[key]; ok {
		return val, true
	}

	// Check user second
	if val, ok := c.User[key]; ok {
		return val, true
	}

	// Check auto last
	if val, ok := c.Auto[key]; ok {
		return val, true
	}

	return nil, false
}
