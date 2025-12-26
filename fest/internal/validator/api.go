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
