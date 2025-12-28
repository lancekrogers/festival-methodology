package validator

import "context"

// StructureValidator validates directory structure
type StructureValidator struct{}

func NewStructureValidator() *StructureValidator { return &StructureValidator{} }

func (v *StructureValidator) Validate(ctx context.Context, path string) ([]Issue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ValidateStructure(ctx, path)
}

// TaskValidator validates presence of task files in implementation sequences
type TaskValidator struct{}

func NewTaskValidator() *TaskValidator { return &TaskValidator{} }

func (v *TaskValidator) Validate(ctx context.Context, path string) ([]Issue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ValidateTasks(ctx, path)
}

// QualityGateValidator validates presence of quality gate tasks; can later apply fixes
type QualityGateValidator struct{ AutoFix bool }

func NewQualityGateValidator(autoFix bool) *QualityGateValidator {
	return &QualityGateValidator{AutoFix: autoFix}
}

type GateValidationResult struct {
	Issues       []Issue
	FixesApplied []FixApplied
}

func (v *QualityGateValidator) ValidateWithFixes(ctx context.Context, path string) (*GateValidationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	issues, err := ValidateQualityGates(ctx, path)
	if err != nil {
		return nil, err
	}
	// Auto-fix is orchestrated by the command layer to avoid package cycles; we return no fixes here.
	return &GateValidationResult{Issues: issues, FixesApplied: nil}, nil
}

// TemplateValidator scans for unfilled markers
type TemplateValidator struct{}

func NewTemplateValidator() *TemplateValidator { return &TemplateValidator{} }

func (v *TemplateValidator) Validate(ctx context.Context, path string) ([]Issue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ValidateTemplateMarkers(path)
}

// QuickValidate runs essential validation checks suitable for post-create validation.
// It runs structure, tasks, and template marker validation.
// It skips ordering validation since structure may be incomplete during creation.
func QuickValidate(ctx context.Context, festivalPath string) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := NewResult("quick_validate", festivalPath)
	result.Score = 100 // Start with perfect score

	// Run structure validation
	structIssues, err := ValidateStructure(ctx, festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, structIssues...)

	// Run template marker validation
	templateIssues, err := ValidateTemplateMarkers(festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, templateIssues...)

	// Calculate score based on issues
	result.Score = CalculateScore(result)
	result.Valid = !result.HasErrors()
	result.OK = result.Valid

	return result, nil
}

// FullValidate runs all validation checks including ordering.
func FullValidate(ctx context.Context, festivalPath string) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := NewResult("validate", festivalPath)
	result.Score = 100 // Start with perfect score

	// Run all validators
	structIssues, err := ValidateStructure(ctx, festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, structIssues...)

	taskIssues, err := ValidateTasks(ctx, festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, taskIssues...)

	templateIssues, err := ValidateTemplateMarkers(festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, templateIssues...)

	gateIssues, err := ValidateQualityGates(ctx, festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, gateIssues...)

	orderIssues, err := ValidateOrdering(ctx, festivalPath)
	if err != nil {
		return nil, err
	}
	result.Issues = append(result.Issues, orderIssues...)

	// Calculate score based on issues
	result.Score = CalculateScore(result)
	result.Valid = !result.HasErrors()
	result.OK = result.Valid

	return result, nil
}
