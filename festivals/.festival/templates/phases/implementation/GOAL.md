---
# Template metadata (for fest CLI discovery)
id: phase-goal-implementation
aliases:
  - pgi
  - implementation-phase
description: Phase goal template for implementation phases with sequential task structure
phase_type: implementation

# Fest document metadata (becomes document frontmatter)
fest_type: phase
fest_id: [REPLACE: PHASE_ID]
fest_name: [REPLACE: Phase Name]
fest_parent: [REPLACE: FESTIVAL_ID]
fest_order: [REPLACE: N]
fest_phase_type: implementation
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

# Phase Goal: [REPLACE: Phase name like 002_IMPLEMENTATION]

**Phase:** [REPLACE: Phase ID] | **Status:** Pending | **Type:** Implementation

## Phase Objective

**Primary Goal:** [REPLACE: What this implementation phase must deliver]

**Context:** [REPLACE: How this phase builds on planning and enables later phases]

## Required Outcomes

Deliverables this phase must produce:

- [ ] [REPLACE: Specific deliverable with acceptance criteria]

<!-- Add more required outcomes as needed -->

## Quality Standards

Quality criteria for all work in this phase:

- [ ] [REPLACE: Quality standard that applies to all sequences]

<!-- Add more quality standards as needed -->

## Sequence Alignment

| Sequence | Goal | Key Deliverable |
|----------|------|-----------------|
| [REPLACE: 01_sequence_name] | [REPLACE: Brief goal] | [REPLACE: Main output] |

<!-- Add rows as sequences are created -->

## Pre-Phase Checklist

Before starting implementation:

- [ ] Planning phase complete
- [ ] Architecture/design decisions documented
- [ ] Dependencies resolved
- [ ] Development environment ready

## Phase Progress

### Sequence Completion

- [ ] [REPLACE: First sequence name]

<!-- Track sequence completion here -->

## Notes

[REPLACE: Technical constraints, assumptions, or integration notes]

---

*Implementation phases use numbered sequences. Create sequences with `fest create sequence`.*
