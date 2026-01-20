---
# Template metadata (for fest CLI discovery)
id: phase-goal-planning
aliases:
  - pgp
  - planning-phase
description: Phase goal template for planning phases with freeform exploration structure
phase_type: planning

# Fest document metadata (becomes document frontmatter)
fest_type: phase
fest_id: [REPLACE: PHASE_ID]
fest_name: [REPLACE: Phase Name]
fest_parent: [REPLACE: FESTIVAL_ID]
fest_order: [REPLACE: N]
fest_phase_type: planning
fest_status: pending
fest_tracking: true
fest_created: {{ .created_date }}
---

# Phase Goal: [REPLACE: Phase name like 001_PLANNING]

**Phase:** [REPLACE: Phase ID] | **Status:** Pending | **Type:** Planning

## Phase Objective

**Primary Goal:** [REPLACE: What this planning phase needs to figure out or decide]

**Context:** [REPLACE: Why this planning is needed before implementation can begin]

## Exploration Topics

What areas need to be explored during this phase:

- [REPLACE: Topic or area to investigate]

<!-- Add more exploration topics as identified -->

## Key Questions to Answer

Questions that must be answered before this phase is complete:

- [REPLACE: Critical question that needs resolution]

<!-- Add more questions as they emerge -->

## Expected Documents

Documents that will be produced during this phase:

- [REPLACE: Document name and purpose]

<!-- Add more documents as planning progresses -->

## Success Criteria

This planning phase is complete when:

- [ ] [REPLACE: Key planning outcome achieved]

<!-- Add more success criteria as they become clear -->

## Notes

[REPLACE: Any assumptions, constraints, or open items]

---

*Planning phases use freeform structure. Create topic directories as needed.*
