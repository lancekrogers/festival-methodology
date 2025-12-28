---
id: sequence
aliases:
  - seq
description: Plan a sequence of related tasks within a phase
---

# Sequence: [NN_sequence_name]

**Sequence Number:** [NN] | **Phase:** [Phase Name] | **Parallel Sequences:** [List or None]

> 📋 **Important**: Create a `SEQUENCE_GOAL.md` file in this sequence directory using the SEQUENCE_GOAL_TEMPLATE.md to define specific goals and evaluation criteria for this sequence.

## Objective

[Clear description of what this sequence will accomplish and its role in the phase]

## Context

[Why this sequence is needed, its dependencies on other sequences, and what it enables]

## Success Criteria

- [ ] [Specific measurable outcome 1]
- [ ] [Specific measurable outcome 2]
- [ ] [Specific measurable outcome 3]

## Planned Tasks

### Core Tasks

1. **[01_task_name]** - [Brief description]
   - Dependencies: None
   - Parallel Group: 1

2. **[02_task_name]** - [Brief description]
   - Dependencies: Task 01
   - Parallel Group: None

3. **[03_task_name]** - [Brief description]
   - Dependencies: Task 02
   - Parallel Group: None

### Quality Verification Tasks

1. **[04_testing_and_verify]** - Validate all sequence deliverables
   - Dependencies: All core tasks
   - Parallel Group: None

2. **[05_code_review]** - Review implementation quality
   - Dependencies: Task 04
   - Parallel Group: None

3. **[06_review_results_iterate]** - Address findings and iterate if needed
   - Dependencies: Task 05
   - Parallel Group: None

## Interfaces Produced

[List any interfaces, contracts, or specifications this sequence will define]

- Interface/Contract 1
- Interface/Contract 2

## Dependencies

**Requires from other sequences:**

- [Sequence/Task]: [What is needed]

**Provides to other sequences:**

- [What this sequence produces that others need]

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [Risk description] | Low/Med/High | Low/Med/High | [Mitigation strategy] |

## Estimated Complexity

**Effort:** [Low/Medium/High]
**Technical Difficulty:** [Low/Medium/High]
**Coordination Required:** [Low/Medium/High]

## Notes

[Additional context, assumptions, constraints, or considerations]
