package festival

import "github.com/lancekrogers/festival-methodology/fest/internal/config"

// DefaultGateTemplates contains embedded default gate template content.
// These are used as fallback when templates don't exist in the template root.
var DefaultGateTemplates = map[string]string{
	"QUALITY_GATE_TESTING.md": `# Task: Testing and Verification

**Task Number:** 01 | **Parallel Group:** None | **Dependencies:** All implementation tasks | **Autonomy:** medium

## Objective

Verify all functionality implemented in this sequence works correctly through comprehensive testing.

## Requirements

- [ ] All unit tests pass
- [ ] Integration tests verify main workflows
- [ ] Manual testing confirms user stories work as expected
- [ ] Error cases are handled correctly
- [ ] Edge cases are addressed

## Test Categories

### Unit Tests

[REPLACE: Run your project's test command]

**Verify:**

- [ ] All new/modified code has test coverage
- [ ] Tests are meaningful (not just coverage padding)
- [ ] Test names describe what they verify

### Integration Tests

[REPLACE: Run your project's integration test command]

**Verify:**

- [ ] Components work together correctly
- [ ] External integrations function properly
- [ ] Data flows correctly through the system

### Manual Verification

Walk through each requirement from the sequence:

1. [ ] **Requirement 1**: [Describe manual test steps and expected result]
2. [ ] **Requirement 2**: [Describe manual test steps and expected result]
3. [ ] **Requirement 3**: [Describe manual test steps and expected result]

## Coverage Requirements

- Minimum coverage: [REPLACE: coverage threshold, e.g., 80%] for new code

[REPLACE: Run your project's coverage command]

## Error Handling Verification

- [ ] Invalid inputs are rejected gracefully
- [ ] Error messages are clear and actionable
- [ ] Errors don't expose sensitive information
- [ ] Recovery paths work correctly

## Definition of Done

- [ ] All automated tests pass
- [ ] Manual verification complete
- [ ] Coverage meets requirements
- [ ] Error handling verified
- [ ] No regressions introduced

## Notes

Document any test gaps, flaky tests, or areas needing future attention here.

---

**Test Results Summary:**

- Unit tests: [ ] Pass / [ ] Fail
- Integration tests: [ ] Pass / [ ] Fail
- Manual tests: [ ] Pass / [ ] Fail
- Coverage: ____%
`,
	"QUALITY_GATE_REVIEW.md": `# Task: Code Review

**Task Number:** 02 | **Parallel Group:** None | **Dependencies:** Testing and Verification | **Autonomy:** low

## Objective

Review all code changes in this sequence for quality, correctness, and adherence to project standards.

## Review Checklist

### Code Quality

- [ ] Code is readable and well-organized
- [ ] Functions/methods are focused (single responsibility)
- [ ] No unnecessary complexity
- [ ] Naming is clear and consistent
- [ ] Comments explain "why" not "what"

### Architecture & Design

- [ ] Changes align with project architecture
- [ ] No unnecessary coupling introduced
- [ ] Dependencies are appropriate
- [ ] Interfaces are clean and focused
- [ ] No code duplication

### Standards Compliance

[REPLACE: Run your project's lint command]

- [ ] Linting passes without warnings
- [ ] Formatting is consistent
- [ ] Project conventions are followed

### Error Handling

- [ ] Errors are handled appropriately
- [ ] Error messages are helpful
- [ ] No panic/crash scenarios
- [ ] Resources are properly cleaned up

### Security Considerations

- [ ] No secrets in code
- [ ] Input validation present
- [ ] No SQL injection risks
- [ ] No XSS vulnerabilities
- [ ] Proper authentication/authorization

### Performance

- [ ] No obvious performance issues
- [ ] Queries are efficient
- [ ] No memory leaks
- [ ] Appropriate caching used

### Testing

- [ ] Tests are meaningful
- [ ] Edge cases covered
- [ ] Test data is appropriate
- [ ] Mocks used correctly

## Review Process

1. **Read the sequence goal** - Understand what was being built
2. **Review file by file** - Check each modified file
3. **Run the code** - Verify functionality works
4. **Document findings** - Note issues and suggestions

## Findings

### Critical Issues (Must Fix)

1. [ ] [Issue description and recommendation]

### Suggestions (Should Consider)

1. [ ] [Suggestion and rationale]

### Positive Observations

- [Note good patterns or practices observed]

## Definition of Done

- [ ] All files reviewed
- [ ] Linting passes
- [ ] No critical issues remaining
- [ ] Suggestions documented
- [ ] Knowledge shared with team (if applicable)

## Review Summary

**Reviewer:** [Name/Agent]
**Date:** [Date]
**Verdict:** [ ] Approved / [ ] Needs Changes

**Notes:**
[Summary of the review and any outstanding concerns]
`,
	"QUALITY_GATE_ITERATE.md": `# Task: Review Results and Iterate

**Task Number:** 03 | **Parallel Group:** None | **Dependencies:** Code Review | **Autonomy:** medium

## Objective

Address all findings from code review and testing, iterate until the sequence meets quality standards.

## Review Findings to Address

### From Testing

| Finding | Priority | Status | Notes |
|---------|----------|--------|-------|
| [Finding 1] | [High/Medium/Low] | [ ] Fixed | |
| [Finding 2] | [High/Medium/Low] | [ ] Fixed | |

### From Code Review

| Finding | Priority | Status | Notes |
|---------|----------|--------|-------|
| [Finding 1] | [High/Medium/Low] | [ ] Fixed | |
| [Finding 2] | [High/Medium/Low] | [ ] Fixed | |

## Iteration Process

### Round 1

**Changes Made:**

- [ ] [Change 1 description]
- [ ] [Change 2 description]

**Verification:**

- [ ] Tests re-run and pass
- [ ] Linting passes
- [ ] Changes reviewed

### Round 2 (if needed)

**Changes Made:**

- [ ] [Change 1 description]

**Verification:**

- [ ] Tests re-run and pass
- [ ] Linting passes
- [ ] Changes reviewed

## Final Verification

After all iterations:

- [ ] All critical findings addressed
- [ ] All tests pass
- [ ] Linting passes
- [ ] Code review approved
- [ ] Sequence objectives met

## Lessons Learned

Document patterns or issues to avoid in future sequences:

### What Went Well

- [Positive observation]

### What Could Improve

- [Area for improvement]

### Process Improvements

- [Suggestion for future work]

## Definition of Done

- [ ] All critical findings fixed
- [ ] All tests pass
- [ ] Linting passes
- [ ] Code review approval received
- [ ] Lessons learned documented
- [ ] Ready to proceed to next sequence

## Sign-Off

**Sequence Complete:** [ ] Yes / [ ] No

**Final Status:**

- Tests: [ ] All Pass
- Review: [ ] Approved
- Quality: [ ] Meets Standards

**Notes:**
[Any final notes or observations about this sequence]

---

**Next Steps:**
[Identify what follows - next sequence, phase completion, etc.]
`,

	// Commit gate for sequence completion tracking
	"QUALITY_GATE_COMMIT.md": `# Task: Commit Changes

**Task Number:** 04 | **Parallel Group:** None | **Dependencies:** Review Results and Iterate | **Autonomy:** high

## Objective

Commit all changes from this sequence with proper documentation.

## Pre-Commit Checklist

- [ ] All tests pass
- [ ] Code review approved
- [ ] No uncommitted changes unrelated to this sequence
- [ ] Commit message follows project conventions

## Commit

Use fest commit so task references are preserved:

    fest commit -m "<type>: <summary>"
    # Optional: fest commit --task FEST-XXXX -m "<type>: <summary>"

## Post-Commit

- [ ] Changes pushed to remote
- [ ] CI/CD pipeline passed
- [ ] Sequence marked complete

## Definition of Done

- [ ] Changes committed with descriptive message
- [ ] Changes pushed to remote repository
- [ ] Sequence completion documented
`,
}

// DefaultFestivalGatesConfig creates a festival config with gates/ prefixed template paths.
// This is used when creating a new festival to set up default quality gates.
func DefaultFestivalGatesConfig() *config.FestivalConfig {
	return &config.FestivalConfig{
		Version: "1.0",
		QualityGates: config.QualityGatesConfig{
			Enabled:    true,
			AutoAppend: true,
			Tasks: []config.QualityGateTask{
				{
					ID:       "testing_and_verify",
					Template: "gates/QUALITY_GATE_TESTING",
					Name:     "Testing and Verification",
					Enabled:  true,
				},
				{
					ID:       "code_review",
					Template: "gates/QUALITY_GATE_REVIEW",
					Name:     "Code Review",
					Enabled:  true,
				},
				{
					ID:       "review_results_iterate",
					Template: "gates/QUALITY_GATE_ITERATE",
					Name:     "Review Results and Iterate",
					Enabled:  true,
				},
				{
					ID:       "commit",
					Template: "gates/QUALITY_GATE_COMMIT",
					Name:     "Commit Changes",
					Enabled:  true,
				},
			},
		},
		ExcludedPatterns: []string{
			"*_planning",
			"*_research",
			"*_requirements",
			"*_docs",
		},
		Templates: config.TemplatePrefs{
			TaskDefault:  "TASK_TEMPLATE_SIMPLE",
			PreferSimple: true,
		},
		Tracking: config.TrackingConfig{
			Enabled:      true,
			ChecksumFile: ".festival-checksums.json",
		},
	}
}
