---
id: research-phase-goal
aliases:
  - rpg
description: Defines research phase objective, scope, and expected findings
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template

PURPOSE: Use this template for research phases that:
- Explore unknown problem spaces
- Gather information before planning
- Investigate options and approaches
- Do NOT require structured sequences/tasks
-->

# Research Phase: {{.phase_id}}

**Phase:** {{.phase_id}} | **Type:** Research | **Status:** [REPLACE: Active/Complete]

## Research Objective

**Primary Question:** [REPLACE: The main question this research phase aims to answer]

**Context:** [REPLACE: Why this research is needed and how it supports the festival goal]

## Research Scope

### Topics to Investigate

- [REPLACE: Topic 1 - specific area of investigation]
- [REPLACE: Topic 2 - specific area of investigation]
- [REPLACE: Topic 3 - specific area of investigation]

### Out of Scope

- [REPLACE: What explicitly will NOT be covered]
- [REPLACE: Topics deferred to future research]

## Expected Deliverables

| Deliverable | Type | Purpose |
|-------------|------|---------|
| [REPLACE: Name] | Investigation | [REPLACE: What it provides] |
| [REPLACE: Name] | Comparison | [REPLACE: What it provides] |
| [REPLACE: Name] | Specification | [REPLACE: What it provides] |
| [REPLACE: Name] | Analysis | [REPLACE: What it provides] |

## Research Approach

### Methods

- [REPLACE: Research method 1 - e.g., code analysis, documentation review]
- [REPLACE: Research method 2 - e.g., benchmarking, prototype testing]
- [REPLACE: Research method 3 - e.g., stakeholder interviews]

### Sources

- [REPLACE: Source 1 - documentation, existing code, external resources]
- [REPLACE: Source 2]
- [REPLACE: Source 3]

## Directory Structure

Organize research documents in subdirectories by topic:

```
{{.phase_id}}/
├── PHASE_GOAL.md                    # This file
├── [topic_1]/
│   ├── investigation_[topic].md     # Investigation documents
│   ├── comparison_[options].md      # Comparison documents
│   └── analysis_[subject].md        # Analysis documents
├── [topic_2]/
│   └── ...
└── results/
    └── research_summary.md          # Synthesized findings
```

## Success Criteria

Research is complete when:

- [ ] Primary question is answered with evidence
- [ ] All expected deliverables are produced
- [ ] Findings are documented with sources
- [ ] Recommendations are actionable
- [ ] Next phase can be planned based on findings

## Findings Summary

**Status:** [REPLACE: Not Started/In Progress/Complete]

### Key Findings

1. [REPLACE: Finding 1 with supporting evidence]
2. [REPLACE: Finding 2 with supporting evidence]
3. [REPLACE: Finding 3 with supporting evidence]

### Recommendations

Based on research findings:

- [REPLACE: Recommendation for implementation phases]
- [REPLACE: Recommendation for architecture/design]
- [REPLACE: Recommendation for future research]

## Impact on Festival

### Informs These Phases

- [REPLACE: Phase that will use these findings]
- [REPLACE: Phase that depends on these decisions]

### Open Questions for Future Research

- [REPLACE: Question that needs deeper investigation]
- [REPLACE: Question that emerged during research]

## Stakeholder Review

| Reviewer | Role | Date | Outcome |
|----------|------|------|---------|
| [REPLACE: Name] | [REPLACE: Role] | [REPLACE: Date] | [REPLACE: Approved/Needs revision] |
