---
# Template metadata (for fest CLI discovery)
id: phase-goal-non-coding-action
aliases:
  - pgnca
  - action-phase
description: Phase goal template for non-coding action phases (documentation, releases, configuration, etc.)
phase_type: non_coding_action

# Fest document metadata (becomes document frontmatter)
fest_type: phase
fest_id: [REPLACE: PHASE_ID]
fest_name: [REPLACE: Phase Name]
fest_parent: [REPLACE: FESTIVAL_ID]
fest_order: [REPLACE: N]
fest_phase_type: non_coding_action
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

# Phase Goal: [REPLACE: Phase name like 003_RELEASE]

**Phase:** [REPLACE: Phase ID] | **Status:** Pending | **Type:** Non-Coding Action

## Phase Objective

**Primary Goal:** [REPLACE: What non-coding outcome this phase must achieve]

**Context:** [REPLACE: Why this action phase is needed and what it enables]

## Action Items

Tasks to complete during this phase:

- [ ] [REPLACE: Action item to complete]

<!-- Add more action items as identified -->

## Prerequisites

What must be in place before starting:

- [ ] [REPLACE: Prerequisite condition]

<!-- Add more prerequisites as needed -->

## Verification Steps

How to verify each action was completed successfully:

- [REPLACE: Verification step for action item]

<!-- Add verification steps for each action -->

## Success Criteria

This action phase is complete when:

- [ ] [REPLACE: Key outcome achieved]
- [ ] All action items verified complete

<!-- Add more success criteria as they become clear -->

## Notes

[REPLACE: Any dependencies, constraints, or special instructions]

---

*Non-coding action phases handle documentation, releases, configuration, and other non-code tasks.*
