---
# Template metadata (for fest CLI discovery)
id: QUALITY_GATE_ITERATE
aliases:
  - review-iterate
  - qg-iterate
description: Standard quality gate task for addressing review findings and iterating

# Fest document metadata (becomes document frontmatter)
fest_type: gate
fest_id: {{ .GateID }}
fest_name: Review Results and Iterate
fest_parent: {{ .SequenceID }}
fest_order: {{ .TaskNumber }}
fest_gate_type: iterate
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

# Task: Review Results and Iterate

**Task Number:** {{ .TaskNumber }} | **Parallel Group:** None | **Dependencies:** Code Review | **Autonomy:** medium

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
