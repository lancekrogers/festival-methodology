---
id: sequence-goal
aliases:
  - sg
description: Defines sequence objective, deliverables, and quality standards
---

---
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

- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]
- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]
- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]

### Quality Standards

- [ ] **[REPLACE: Standard name]**: [REPLACE: Quality measure with specific target]
- [ ] **[REPLACE: Standard name]**: [REPLACE: Quality measure with specific target]

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

- [ ] **Milestone 1**: [REPLACE: Key deliverable achieved]
- [ ] **Milestone 2**: [REPLACE: Key deliverable achieved]
- [ ] **Milestone 3**: [REPLACE: Key deliverable achieved]

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
