package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()

	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Custom)
	assert.Empty(t, ctx.FestivalName)
	assert.Empty(t, ctx.PhaseNumber)
	assert.Equal(t, "", ctx.CurrentLevel)
}

func TestSetFestival(t *testing.T) {
	ctx := NewContext()
	ctx.SetFestival("my-festival", "Build awesome things", []string{"backend", "api"})

	assert.Equal(t, "my-festival", ctx.FestivalName)
	assert.Equal(t, "Build awesome things", ctx.FestivalGoal)
	assert.Equal(t, []string{"backend", "api"}, ctx.FestivalTags)
	assert.Equal(t, "festival", ctx.CurrentLevel)
}

func TestSetFestivalDescription(t *testing.T) {
	ctx := NewContext()
	ctx.SetFestivalDescription("Extended description of the festival")

	assert.Equal(t, "Extended description of the festival", ctx.FestivalDescription)
}

func TestSetPhase(t *testing.T) {
	ctx := NewContext()
	ctx.SetPhase(1, "PLANNING", "planning")

	assert.Equal(t, 1, ctx.PhaseNumber)
	assert.Equal(t, "PLANNING", ctx.PhaseName)
	assert.Equal(t, "001_PLANNING", ctx.PhaseID)
	assert.Equal(t, "planning", ctx.PhaseType)
	assert.Equal(t, "phase", ctx.CurrentLevel)
}

func TestSetPhaseStructure(t *testing.T) {
	ctx := NewContext()
	ctx.SetPhaseStructure("freeform")

	assert.Equal(t, "freeform", ctx.PhaseStructure)
}

func TestSetPhaseObjective(t *testing.T) {
	ctx := NewContext()
	ctx.SetPhaseObjective("Define comprehensive requirements")

	assert.Equal(t, "Define comprehensive requirements", ctx.PhaseObjective)
}

func TestSetSequence(t *testing.T) {
	ctx := NewContext()
	ctx.SetSequence(1, "requirements gathering")

	assert.Equal(t, 1, ctx.SequenceNumber)
	assert.Equal(t, "requirements gathering", ctx.SequenceName)
	assert.Equal(t, "01_requirements_gathering", ctx.SequenceID)
	assert.Equal(t, "sequence", ctx.CurrentLevel)
}

func TestSetSequenceObjective(t *testing.T) {
	ctx := NewContext()
	ctx.SetSequenceObjective("Gather all user requirements")

	assert.Equal(t, "Gather all user requirements", ctx.SequenceObjective)
}

func TestSetSequenceDependencies(t *testing.T) {
	ctx := NewContext()
	deps := []string{"01_requirements", "02_architecture"}
	ctx.SetSequenceDependencies(deps)

	assert.Equal(t, deps, ctx.SequenceDependencies)
}

func TestSetTask(t *testing.T) {
	ctx := NewContext()
	ctx.SetTask(1, "user research")

	assert.Equal(t, 1, ctx.TaskNumber)
	assert.Equal(t, "user research", ctx.TaskName)
	assert.Equal(t, "01_user_research.md", ctx.TaskID)
	assert.Equal(t, "task", ctx.CurrentLevel)
}

func TestSetTaskObjective(t *testing.T) {
	ctx := NewContext()
	ctx.SetTaskObjective("Research user needs and pain points")

	assert.Equal(t, "Research user needs and pain points", ctx.TaskObjective)
}

func TestSetTaskDeliverables(t *testing.T) {
	ctx := NewContext()
	deliverables := []string{"User interview notes", "Survey results", "Persona documents"}
	ctx.SetTaskDeliverables(deliverables)

	assert.Equal(t, deliverables, ctx.TaskDeliverables)
}

func TestSetTaskParallel(t *testing.T) {
	ctx := NewContext()
	ctx.SetTaskParallel(true)

	assert.True(t, ctx.TaskParallel)
}

func TestSetTaskDependencies(t *testing.T) {
	ctx := NewContext()
	deps := []string{"01_user_research", "02_competitor_analysis"}
	ctx.SetTaskDependencies(deps)

	assert.Equal(t, deps, ctx.TaskDependencies)
}

func TestSetCustom(t *testing.T) {
	ctx := NewContext()
	ctx.SetCustom("custom_key", "custom_value")
	ctx.SetCustom("another_key", 123)

	assert.Equal(t, "custom_value", ctx.Custom["custom_key"])
	assert.Equal(t, 123, ctx.Custom["another_key"])
}

func TestComputeStructureVariables(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*Context)
		wantPath         string
		wantParentPhase  string
		wantParentSeq    string
		wantCurrentLevel string
		wantFestivalRoot string
	}{
		{
			name: "task level",
			setup: func(ctx *Context) {
				ctx.SetPhase(1, "PLANNING", "planning")
				ctx.SetSequence(1, "requirements")
				ctx.SetTask(1, "user research")
				ctx.ComputeStructureVariables()
			},
			wantPath:         "001_PLANNING/01_requirements/01_user_research.md",
			wantParentPhase:  "001_PLANNING",
			wantParentSeq:    "01_requirements",
			wantCurrentLevel: "task",
			wantFestivalRoot: "../..",
		},
		{
			name: "sequence level",
			setup: func(ctx *Context) {
				ctx.SetPhase(2, "IMPLEMENTATION", "implementation")
				ctx.SetSequence(3, "backend api")
				ctx.ComputeStructureVariables()
			},
			wantPath:         "002_IMPLEMENTATION/03_backend_api",
			wantParentPhase:  "002_IMPLEMENTATION",
			wantParentSeq:    "",
			wantCurrentLevel: "sequence",
			wantFestivalRoot: "../..",
		},
		{
			name: "phase level",
			setup: func(ctx *Context) {
				ctx.SetPhase(3, "VALIDATION", "validation")
				ctx.ComputeStructureVariables()
			},
			wantPath:         "003_VALIDATION",
			wantParentPhase:  "",
			wantParentSeq:    "",
			wantCurrentLevel: "phase",
			wantFestivalRoot: "..",
		},
		{
			name: "festival level",
			setup: func(ctx *Context) {
				ctx.SetFestival("my-project", "Build something", []string{"tag1"})
				ctx.ComputeStructureVariables()
			},
			wantPath:         "",
			wantParentPhase:  "",
			wantParentSeq:    "",
			wantCurrentLevel: "festival",
			wantFestivalRoot: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			tt.setup(ctx)

			assert.Equal(t, tt.wantPath, ctx.FullPath)
			assert.Equal(t, tt.wantParentPhase, ctx.ParentPhaseID)
			assert.Equal(t, tt.wantParentSeq, ctx.ParentSequenceID)
			assert.Equal(t, tt.wantCurrentLevel, ctx.CurrentLevel)
			assert.Equal(t, tt.wantFestivalRoot, ctx.FestivalRoot)
		})
	}
}

func TestToTemplateData(t *testing.T) {
	ctx := NewContext()
	ctx.SetFestival("test-festival", "Test goal", []string{"tag1", "tag2"})
	ctx.SetFestivalDescription("Extended description")
	ctx.SetPhase(1, "PLANNING", "planning")
	ctx.SetPhaseStructure("structured")
	ctx.SetPhaseObjective("Plan the project")
	ctx.SetSequence(2, "architecture design")
	ctx.SetSequenceObjective("Design system architecture")
	ctx.SetTask(3, "database schema")
	ctx.SetTaskObjective("Design database schema")
	ctx.SetTaskDeliverables([]string{"ERD diagram", "Migration scripts"})
	ctx.SetTaskParallel(true)
	ctx.SetCustom("custom_var", "custom_value")
	ctx.ComputeStructureVariables()

	data := ctx.ToTemplateData()

	// Festival-level
	assert.Equal(t, "test-festival", data["festival_name"])
	assert.Equal(t, "Test goal", data["festival_goal"])
	assert.Equal(t, []string{"tag1", "tag2"}, data["festival_tags"])
	assert.Equal(t, "Extended description", data["festival_description"])

	// Phase-level
	assert.Equal(t, 1, data["phase_number"])
	assert.Equal(t, "PLANNING", data["phase_name"])
	assert.Equal(t, "001_PLANNING", data["phase_id"])
	assert.Equal(t, "planning", data["phase_type"])
	assert.Equal(t, "structured", data["phase_structure"])
	assert.Equal(t, "Plan the project", data["phase_objective"])

	// Sequence-level
	assert.Equal(t, 2, data["sequence_number"])
	assert.Equal(t, "architecture design", data["sequence_name"])
	assert.Equal(t, "02_architecture_design", data["sequence_id"])
	assert.Equal(t, "Design system architecture", data["sequence_objective"])

	// Task-level
	assert.Equal(t, 3, data["task_number"])
	assert.Equal(t, "database schema", data["task_name"])
	assert.Equal(t, "03_database_schema.md", data["task_id"])
	assert.Equal(t, "Design database schema", data["task_objective"])
	assert.Equal(t, []string{"ERD diagram", "Migration scripts"}, data["task_deliverables"])
	assert.Equal(t, true, data["task_parallel"])

	// Structure
	assert.Equal(t, "task", data["current_level"])
	assert.Equal(t, "001_PLANNING", data["parent_phase_id"])
	assert.Equal(t, "02_architecture_design", data["parent_sequence_id"])
	assert.Equal(t, "001_PLANNING/02_architecture_design/03_database_schema.md", data["full_path"])
	assert.Equal(t, "../..", data["festival_root"])

	// Custom
	assert.Equal(t, "custom_value", data["custom_var"])
}

func TestGet(t *testing.T) {
	ctx := NewContext()
	ctx.SetFestival("test-festival", "Test goal", []string{"tag1"})
	ctx.SetCustom("custom_key", "custom_value")

	// Test getting festival variable
	val, ok := ctx.Get("festival_name")
	assert.True(t, ok)
	assert.Equal(t, "test-festival", val)

	// Test getting custom variable
	val, ok = ctx.Get("custom_key")
	assert.True(t, ok)
	assert.Equal(t, "custom_value", val)

	// Test nonexistent variable
	val, ok = ctx.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestFormatPhaseID(t *testing.T) {
	tests := []struct {
		number int
		name   string
		want   string
	}{
		{1, "PLANNING", "001_PLANNING"},
		{2, "IMPLEMENTATION", "002_IMPLEMENTATION"},
		{10, "REVIEW", "010_REVIEW"},
		{99, "FINAL PHASE", "099_FINAL_PHASE"},
		{1, "planning", "001_PLANNING"},         // lowercase converted to uppercase
		{5, "User Testing", "005_USER_TESTING"}, // spaces to underscores
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPhaseID(tt.number, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatSequenceID(t *testing.T) {
	tests := []struct {
		number int
		name   string
		want   string
	}{
		{1, "requirements", "01_requirements"},
		{2, "architecture design", "02_architecture_design"},
		{15, "integration testing", "15_integration_testing"},
		{3, "BACKEND API", "03_backend_api"},       // uppercase converted to lowercase
		{8, "User Interface", "08_user_interface"}, // spaces to underscores
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSequenceID(tt.number, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTaskID(t *testing.T) {
	tests := []struct {
		number int
		name   string
		want   string
	}{
		{1, "user research", "01_user_research.md"},
		{2, "competitor analysis", "02_competitor_analysis.md"},
		{25, "code review", "25_code_review.md"},
		{5, "Database Schema", "05_database_schema.md"}, // mixed case normalized
		{10, "API Design", "10_api_design.md"},          // spaces to underscores
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTaskID(tt.number, tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test complete workflow: create full context hierarchy
func TestCompleteWorkflow(t *testing.T) {
	ctx := NewContext()

	// Set festival
	ctx.SetFestival("ecommerce-platform", "Build a scalable e-commerce platform", []string{"backend", "api", "payments"})
	ctx.SetFestivalDescription("A complete e-commerce solution with payment processing")

	// Set phase
	ctx.SetPhase(1, "PLANNING", "planning")
	ctx.SetPhaseStructure("structured")
	ctx.SetPhaseObjective("Define requirements and architecture")

	// Set sequence
	ctx.SetSequence(2, "architecture design")
	ctx.SetSequenceObjective("Design system architecture and data models")
	ctx.SetSequenceDependencies([]string{"01_requirements_gathering"})

	// Set task
	ctx.SetTask(3, "database schema design")
	ctx.SetTaskObjective("Design complete database schema")
	ctx.SetTaskDeliverables([]string{
		"ERD diagram",
		"Migration scripts",
		"Index strategy document",
	})
	ctx.SetTaskParallel(false)
	ctx.SetTaskDependencies([]string{"01_requirements_definition", "02_data_modeling"})

	// Add custom variable
	ctx.SetCustom("project_lead", "Alice")

	// Compute structure
	ctx.ComputeStructureVariables()

	// Verify complete context
	assert.Equal(t, "ecommerce-platform", ctx.FestivalName)
	assert.Equal(t, "001_PLANNING", ctx.PhaseID)
	assert.Equal(t, "02_architecture_design", ctx.SequenceID)
	assert.Equal(t, "03_database_schema_design.md", ctx.TaskID)
	assert.Equal(t, "task", ctx.CurrentLevel)
	assert.Equal(t, "001_PLANNING", ctx.ParentPhaseID)
	assert.Equal(t, "02_architecture_design", ctx.ParentSequenceID)
	assert.Equal(t, "001_PLANNING/02_architecture_design/03_database_schema_design.md", ctx.FullPath)
	assert.Equal(t, "../..", ctx.FestivalRoot)

	// Verify ToTemplateData includes everything
	data := ctx.ToTemplateData()
	assert.Equal(t, "ecommerce-platform", data["festival_name"])
	assert.Equal(t, "Alice", data["project_lead"])
	assert.Equal(t, 3, len(data["task_deliverables"].([]string)))
}

// Test empty/zero values
func TestEmptyValues(t *testing.T) {
	ctx := NewContext()

	data := ctx.ToTemplateData()

	// Check zero values are present
	assert.Equal(t, "", data["festival_name"])
	assert.Equal(t, 0, data["phase_number"])
	assert.Equal(t, "", data["sequence_id"])
	assert.Nil(t, data["festival_tags"])
}

// Test multiple custom variables
func TestMultipleCustomVariables(t *testing.T) {
	ctx := NewContext()

	ctx.SetCustom("string_var", "value")
	ctx.SetCustom("int_var", 42)
	ctx.SetCustom("bool_var", true)
	ctx.SetCustom("slice_var", []string{"a", "b", "c"})
	ctx.SetCustom("map_var", map[string]int{"key": 123})

	data := ctx.ToTemplateData()

	assert.Equal(t, "value", data["string_var"])
	assert.Equal(t, 42, data["int_var"])
	assert.Equal(t, true, data["bool_var"])
	assert.Equal(t, []string{"a", "b", "c"}, data["slice_var"])
	assert.Equal(t, map[string]int{"key": 123}, data["map_var"])
}

// Test overwriting values
func TestOverwritingValues(t *testing.T) {
	ctx := NewContext()

	// Set festival twice
	ctx.SetFestival("first-name", "first goal", []string{"tag1"})
	ctx.SetFestival("second-name", "second goal", []string{"tag2"})

	assert.Equal(t, "second-name", ctx.FestivalName)
	assert.Equal(t, "second goal", ctx.FestivalGoal)
	assert.Equal(t, []string{"tag2"}, ctx.FestivalTags)

	// Set custom variable twice
	ctx.SetCustom("key", "value1")
	ctx.SetCustom("key", "value2")

	assert.Equal(t, "value2", ctx.Custom["key"])
}
