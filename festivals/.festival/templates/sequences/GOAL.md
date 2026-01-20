---
# Template metadata (for fest CLI discovery)
id: sequence-goal
aliases:
  - sg
description: Defines sequence objective, deliverables, and quality standards

# Fest document metadata (becomes document frontmatter)
fest_type: sequence
fest_id: [REPLACE: SEQUENCE_ID]
fest_name: [REPLACE: Sequence Name]
fest_parent: [REPLACE: PHASE_ID]
fest_order: [REPLACE: N]
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Sequence Goal: [REPLACE: NN_sequence_name]

**Sequence:** [REPLACE: NN_sequence_name] | **Phase:** [REPLACE: NNN_PHASE_NAME] | **Status:** Pending | **Created:** {{ .created_date }}

## Sequence Objective

**Primary Goal:** [REPLACE: One clear sentence stating what this sequence must accomplish]

**Contribution to Phase Goal:** [REPLACE: How achieving this sequence goal directly supports the phase goal]

## Success Criteria

The sequence goal is achieved when:

### Required Deliverables

- [ ] **[REPLACE: Deliverable 1 name]**: [REPLACE: Deliverable 1 description]
- [ ] **[REPLACE: Deliverable 2 name]**: [REPLACE: Deliverable 2 description]
- [ ] **[REPLACE: Deliverable 3 name]**: [REPLACE: Deliverable 3 description]

### Quality Standards

- [ ] **[REPLACE: Quality standard 1]**: [REPLACE: Quality target 1]
- [ ] **[REPLACE: Quality standard 2]**: [REPLACE: Quality target 2]

### Completion Criteria

- [ ] All tasks in sequence completed successfully
- [ ] Quality verification tasks passed
- [ ] Code review completed and issues addressed
- [ ] Documentation updated

## Task Alignment

> **Note:** This table should be populated AFTER creating task files.
> SEQUENCE_GOAL.md defines WHAT to accomplish. Task files define HOW.
> Run `fest create task` to create tasks, then update this table.

| Task | Task Objective | Contribution to Sequence Goal |
|------|----------------|-------------------------------|
| [FILL: after creating tasks] | | |

## Dependencies

### Prerequisites (from other sequences)

- [REPLACE: Sequence X]: [REPLACE: What we need from it]

### Provides (to other sequences)

- [REPLACE: What this sequence produces]: Used by [REPLACE: Sequence Z]

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [REPLACE: Risk description] | [REPLACE: Low/Med/High] | [REPLACE: Low/Med/High] | [REPLACE: Prevention strategy] |

## Progress Tracking

### Milestones

- [ ] **Milestone 1**: [REPLACE: First key deliverable]
- [ ] **Milestone 2**: [REPLACE: Second key deliverable]
- [ ] **Milestone 3**: [REPLACE: Final key deliverable]

## Quality Gates

### Testing and Verification

- [ ] All unit tests pass
- [ ] Integration tests complete
- [ ] Performance benchmarks met

### Code Review

- [ ] Code review conducted
- [ ] Review feedback addressed
- [ ] Standards compliance verified

### Iteration Decision

- [ ] Need another iteration? [REPLACE: Yes/No]
- [ ] If yes, new tasks created: [REPLACE: List task numbers]
