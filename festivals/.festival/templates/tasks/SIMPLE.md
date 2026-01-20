---
# Template metadata (for fest CLI discovery)
id: task-simple
aliases:
  - simple-task
  - ts
description: Streamlined task template for simple, well-understood tasks

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
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Use this for simple tasks; use TASK_TEMPLATE.md for complex tasks
- Remove this comment block when filling the template
-->

# Task: [REPLACE: Task_Name]

**Task Number:** [REPLACE: N] | **Parallel Group:** [REPLACE: N or None] | **Dependencies:** [REPLACE: Task numbers or None] | **Autonomy:** [REPLACE: high/medium/low]

## Objective

[REPLACE: ONE clear sentence describing what will be accomplished]

## Requirements

- [ ] [REPLACE: Specific, testable requirement 1]
- [ ] [REPLACE: Specific, testable requirement 2]
- [ ] [REPLACE: Specific, testable requirement 3]

## Implementation Steps

### 1. [REPLACE: Step 1 Title]

[REPLACE: Step 1 description and actions]

### 2. [REPLACE: Step 2 Title]

[REPLACE: Step 2 description and actions]

### 3. [REPLACE: Step 3 Title]

[REPLACE: Step 3 description and actions]

## Definition of Done

- [ ] All requirements implemented
- [ ] Tests pass (if applicable)
- [ ] Code reviewed (if applicable)
- [ ] Documentation updated (if applicable)

## Notes

[REPLACE: Any additional context, assumptions, or considerations]
