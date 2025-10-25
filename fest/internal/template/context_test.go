package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextBuilder_Build(t *testing.T) {
	builder := NewContextBuilder()

	ctx := builder.
		WithUser("custom_key", "custom_value").
		WithUser("another_key", 123).
		Build()

	// Check automatic variables exist
	assert.NotNil(t, ctx.Auto)
	assert.Contains(t, ctx.Auto, "today")
	assert.Contains(t, ctx.Auto, "now")
	assert.Contains(t, ctx.Auto, "fest_version")

	// Check user variables
	assert.Equal(t, "custom_value", ctx.User["custom_key"])
	assert.Equal(t, 123, ctx.User["another_key"])

	// Check computed variables
	assert.NotNil(t, ctx.Computed)
}

func TestContextBuilder_BuildForFestival(t *testing.T) {
	builder := NewContextBuilder()

	ctx := builder.BuildForFestival("my-festival", "Build awesome things")

	// Check user variables
	assert.Equal(t, "my-festival", ctx.User["festival_name"])
	assert.Equal(t, "Build awesome things", ctx.User["festival_goal"])

	// Check computed variables
	assert.Contains(t, ctx.Computed, "festival_id")
	assert.Equal(t, "my-festival", ctx.Computed["festival_path"])
}

func TestContextBuilder_BuildForPhase(t *testing.T) {
	builder := NewContextBuilder()

	ctx := builder.BuildForPhase("my-festival", 1, "PLANNING")

	// Check user variables
	assert.Equal(t, "my-festival", ctx.User["festival_name"])
	assert.Equal(t, 1, ctx.User["phase_number"])
	assert.Equal(t, "PLANNING", ctx.User["phase_name"])

	// Check computed variables
	assert.Equal(t, "001_PLANNING", ctx.Computed["phase_id"])
	assert.Contains(t, ctx.Computed["phase_path"], "001_PLANNING")
}

func TestContextBuilder_WithUserMap(t *testing.T) {
	builder := NewContextBuilder()

	userVars := map[string]interface{}{
		"var1": "value1",
		"var2": 42,
		"var3": true,
	}

	ctx := builder.WithUserMap(userVars).Build()

	assert.Equal(t, "value1", ctx.User["var1"])
	assert.Equal(t, 42, ctx.User["var2"])
	assert.Equal(t, true, ctx.User["var3"])
}

func TestContext_ToFlatMap(t *testing.T) {
	ctx := &Context{
		Auto: map[string]interface{}{
			"auto_var": "auto_value",
		},
		User: map[string]interface{}{
			"user_var": "user_value",
		},
		Computed: map[string]interface{}{
			"computed_var": "computed_value",
		},
	}

	flat := ctx.ToFlatMap()

	// All variables should be in flat map
	assert.Equal(t, "auto_value", flat["auto_var"])
	assert.Equal(t, "user_value", flat["user_var"])
	assert.Equal(t, "computed_value", flat["computed_var"])
}

func TestContext_Get(t *testing.T) {
	ctx := &Context{
		Auto: map[string]interface{}{
			"auto_var": "auto_value",
			"override": "auto",
		},
		User: map[string]interface{}{
			"user_var": "user_value",
			"override": "user",
		},
		Computed: map[string]interface{}{
			"computed_var": "computed_value",
			"override":     "computed",
		},
	}

	// Test priority: Computed > User > Auto
	val, ok := ctx.Get("override")
	assert.True(t, ok)
	assert.Equal(t, "computed", val)

	val, ok = ctx.Get("auto_var")
	assert.True(t, ok)
	assert.Equal(t, "auto_value", val)

	val, ok = ctx.Get("user_var")
	assert.True(t, ok)
	assert.Equal(t, "user_value", val)

	val, ok = ctx.Get("computed_var")
	assert.True(t, ok)
	assert.Equal(t, "computed_value", val)

	val, ok = ctx.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestBuildAutomaticVariables(t *testing.T) {
	auto := buildAutomaticVariables()

	// Check required automatic variables
	assert.Contains(t, auto, "now")
	assert.Contains(t, auto, "today")
	assert.Contains(t, auto, "current_year")
	assert.Contains(t, auto, "fest_version")
	assert.Contains(t, auto, "created_date")
	assert.Contains(t, auto, "timestamp")

	// Verify types
	assert.IsType(t, "", auto["today"])
	assert.IsType(t, 0, auto["current_year"])
	assert.IsType(t, int64(0), auto["timestamp"])
}

func TestGenerateFestivalID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "lowercase with hyphens",
			input:    "My Festival",
			contains: "my-festival",
		},
		{
			name:     "underscores to hyphens",
			input:    "my_festival_name",
			contains: "my-festival-name",
		},
		{
			name:     "already lowercase",
			input:    "simple-name",
			contains: "simple-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := generateFestivalID(tt.input)
			assert.Contains(t, id, tt.contains)
			// Should also contain timestamp
			assert.Greater(t, len(id), len(tt.contains))
		})
	}
}
