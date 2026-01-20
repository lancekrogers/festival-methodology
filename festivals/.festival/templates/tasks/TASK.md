---
# Template metadata (for fest CLI discovery)
id: TASK_TEMPLATE
aliases:
  - TASK TEMPLATE
  - TASK-TEMPLATE
description: Full task template with all sections and quality gates

# Fest document metadata (becomes document frontmatter)
fest_type: task
fest_id: [REPLACE: TASK_ID]
fest_name: [REPLACE: Task Name]
fest_parent: [REPLACE: SEQUENCE_ID]
fest_order: [REPLACE: N]
fest_status: pending
fest_autonomy: [REPLACE: high|medium|low]
fest_tracking: true
fest_created: {{ .created_date }}
# Future task routing fields (reserved):
# fest_agent: null
# fest_complexity: medium
# fest_estimated_tokens: null
# fest_requires_human: false
# fest_requires_context: false
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Task: [REPLACE: NN_task_name]

> **Task Number**: [REPLACE: NN] | **Parallel Execution**: [REPLACE: Yes/No] | **Dependencies**: [REPLACE: Prior tasks] | **Autonomy Level**: [REPLACE: high|medium|low]

## Objective

[REPLACE: One clear sentence describing what will be accomplished with specific deliverables]

## Rules Compliance

Before starting this task, review [FESTIVAL_RULES.md]({{.festival_root}}/FESTIVAL_RULES.md), particularly:

- [REPLACE: Relevant section 1]
- [REPLACE: Relevant section 2]

## Context

[REPLACE: Why this task is needed, dependencies, background information]

## Requirements

- [ ] [REPLACE: Specific requirement 1]
- [ ] [REPLACE: Specific requirement 2]
- [ ] [REPLACE: Specific requirement 3]

## Deliverables

- [REPLACE: Specific file or artifact 1]
- [REPLACE: Specific file or artifact 2]
- [REPLACE: Specific file or artifact 3]

## Definition of Done

- [ ] [REPLACE: Completion criteria 1]
- [ ] [REPLACE: Quality criteria 1]
- [ ] [REPLACE: Acceptance criteria 1]

## Pre-Task Checklist

- [ ] Read [FESTIVAL_RULES.md]({{.festival_root}}/FESTIVAL_RULES.md) completely
- [ ] Understand task requirements
- [ ] Review existing code/content and patterns
- [ ] Verify dependencies are complete
- [ ] Plan approach

## Implementation Steps

### 1. [REPLACE: Step 1 Title]

[REPLACE: Step 1 description and actions]

### 2. [REPLACE: Step 2 Title]

[REPLACE: Step 2 description and actions]

### 3. [REPLACE: Step 3 Title]

[REPLACE: Step 3 description and actions]

## Technical Notes

[REPLACE: Technical considerations, constraints, or important information]

## Testing

[REPLACE: How to verify the implementation works correctly]

## Completion Checklist

- [ ] All requirements met
- [ ] Tests pass (if applicable)
- [ ] Documentation updated
- [ ] Quality standards met
- [ ] Self-review completed

## Notes

[REPLACE: Additional information, assumptions, or considerations]
