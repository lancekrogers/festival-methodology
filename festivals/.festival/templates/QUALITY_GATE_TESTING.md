---
id: QUALITY_GATE_TESTING
aliases:
  - testing-verify
  - qg-test
description: Standard quality gate task for testing and verification
---

# Task: Testing and Verification

**Task Number:** {{ .TaskNumber }} | **Parallel Group:** None | **Dependencies:** All implementation tasks | **Autonomy:** medium

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

```bash
{{ if .Customizations.test_command }}{{ .Customizations.test_command }}{{ else }}# Run unit tests
make test
# or
go test ./...
# or
npm test{{ end }}
```

**Verify:**

- [ ] All new/modified code has test coverage
- [ ] Tests are meaningful (not just coverage padding)
- [ ] Test names describe what they verify

### Integration Tests

```bash
# Run integration tests
make test-integration
```

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

{{ if .Customizations.coverage_threshold }}

- Minimum coverage: {{ .Customizations.coverage_threshold }}%
{{ else }}
- Minimum coverage: 80% for new code
{{ end }}

```bash
# Generate coverage report
make coverage
```

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
