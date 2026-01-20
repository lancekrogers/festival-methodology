---
# Template metadata (for fest CLI discovery)
id: QUALITY_GATE_REVIEW
aliases:
  - code-review
  - qg-review
description: Standard quality gate task for code review

# Fest document metadata (becomes document frontmatter)
fest_type: gate
fest_id: {{ .GateID }}
fest_name: Code Review
fest_parent: {{ .SequenceID }}
fest_order: {{ .TaskNumber }}
fest_gate_type: review
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

# Task: Code Review

**Task Number:** {{ .TaskNumber }} | **Parallel Group:** None | **Dependencies:** Testing and Verification | **Autonomy:** low

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
