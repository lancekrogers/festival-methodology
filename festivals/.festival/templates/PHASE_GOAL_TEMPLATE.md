---
id: phase-goal
aliases:
  - pg
description: Defines phase objective, success criteria, and quality metrics
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Phase Goal: [REPLACE: Phase name like 002_IMPLEMENT]

**Phase:** [REPLACE: Phase ID] | **Status:** [REPLACE: Planning/Active/Complete] | **Sequences:** [REPLACE: X total, Y completed]

## Phase Objective

**Primary Goal:** [REPLACE: One clear sentence stating what this phase must accomplish]

**Context:** [REPLACE: Why this phase is critical to the festival's success and how it enables subsequent phases]

## Success Criteria

The phase goal is achieved when:

### Required Outcomes

- [ ] **[REPLACE: Outcome name]**: [REPLACE: Specific, measurable deliverable or milestone]
- [ ] **[REPLACE: Outcome name]**: [REPLACE: Specific, measurable deliverable or milestone]
- [ ] **[REPLACE: Outcome name]**: [REPLACE: Specific, measurable deliverable or milestone]

### Quality Metrics

- [ ] **[REPLACE: Metric name]**: [REPLACE: Quantifiable quality measure with target]
- [ ] **[REPLACE: Metric name]**: [REPLACE: Quantifiable quality measure with target]
- [ ] **[REPLACE: Metric name]**: [REPLACE: Quantifiable quality measure with target]

### Validation Gates

- [ ] All sequence goals within this phase achieved
- [ ] Stakeholder review and approval completed
- [ ] Documentation updated and complete
- [ ] Next phase prerequisites satisfied

## Key Deliverables

| Deliverable | Description | Acceptance Criteria |
|-------------|-------------|---------------------|
| [REPLACE: Deliverable name] | [REPLACE: What it is] | [REPLACE: How to verify completion] |
| [REPLACE: Deliverable name] | [REPLACE: What it is] | [REPLACE: How to verify completion] |
| [REPLACE: Deliverable name] | [REPLACE: What it is] | [REPLACE: How to verify completion] |

## Risk Factors

| Risk | Impact on Goal | Mitigation Strategy |
|------|----------------|---------------------|
| [REPLACE: Risk description] | [REPLACE: How it affects phase goal] | [REPLACE: Prevention/response plan] |
| [REPLACE: Risk description] | [REPLACE: How it affects phase goal] | [REPLACE: Prevention/response plan] |

## Sequence Goal Alignment

Verify that sequence goals support this phase goal:

| Sequence | Sequence Goal | Contribution to Phase Goal |
|----------|---------------|----------------------------|
| [REPLACE: 01_sequence] | [REPLACE: Brief goal statement] | [REPLACE: How it helps achieve phase goal] |
| [REPLACE: 02_sequence] | [REPLACE: Brief goal statement] | [REPLACE: How it helps achieve phase goal] |
| [REPLACE: 03_sequence] | [REPLACE: Brief goal statement] | [REPLACE: How it helps achieve phase goal] |

## Pre-Phase Checklist

Before starting this phase:

- [ ] Previous phase goals fully achieved
- [ ] Resources and team available
- [ ] Dependencies resolved
- [ ] Stakeholders aligned on goals

## Evaluation Framework

### During Execution

Track step completion against:

- Sequence completion rate
- Quality metric trends
- Risk emergence
- Active blockers count

### Post-Completion Assessment

**Date Completed:** [REPLACE: Date when phase was completed]

**Goal Achievement Score:** [REPLACE: X/Y criteria met]

### What Worked Well

- [REPLACE: Success factor]
- [REPLACE: Success factor]
- [REPLACE: Success factor]

### What Could Be Improved

- [REPLACE: Improvement area]
- [REPLACE: Improvement area]
- [REPLACE: Improvement area]

### Lessons Learned

- [REPLACE: Key learning]
- [REPLACE: Key learning]
- [REPLACE: Key learning]

### Recommendations for Future Phases

- [REPLACE: Recommendation]
- [REPLACE: Recommendation]
- [REPLACE: Recommendation]

## Stakeholder Sign-off

| Stakeholder | Role | Sign-off Date | Notes |
|-------------|------|---------------|-------|
| [REPLACE: Name] | [REPLACE: Role] | [REPLACE: Date] | [REPLACE: Any conditions or notes] |

---

## Usage Guide

This PHASE_GOAL.md file should be:

1. Created at the start of phase planning
2. Placed in the phase directory (e.g., `001_PLAN/PHASE_GOAL.md`)
3. Referenced by all sequences within the phase
4. Updated during execution to track progress
5. Completed with assessment after phase completion

The goal serves as:

- North star for all phase activities
- Evaluation criteria for phase completion
- Communication tool with stakeholders
- Learning capture mechanism

### Example (Filled Out)

# Phase Goal: 002_DEFINE_INTERFACES

**Phase:** 002_DEFINE_INTERFACES | **Status:** Planning | **Sequences:** 3 total, 0 completed

## Phase Objective

**Primary Goal:** Define all system interfaces, contracts, and specifications to enable parallel development in Phase 003.

**Context:** This phase is critical because it establishes the contracts that allow multiple teams to work independently without integration conflicts. Without complete interface definitions, Phase 003 will face constant rework and integration issues.

## Success Criteria

The phase goal is achieved when:

### Required Outcomes

- [ ] **API Specification**: All 47 REST endpoints fully documented with OpenAPI
- [ ] **Data Models**: All 12 data models defined with complete schemas
- [ ] **Integration Contracts**: All 5 third-party integrations specified

### Quality Metrics

- [ ] **Completeness**: 100% of identified interfaces have full specifications
- [ ] **Clarity**: 0 ambiguous interface definitions (validated by review)
- [ ] **Examples**: Every interface includes at least 2 usage examples

### Validation Gates

- [ ] All sequence goals within this phase achieved
- [ ] Technical review with all development teams passed
- [ ] Stakeholder approval on all public interfaces received
- [ ] Interface freeze agreed and documented

## Key Deliverables

| Deliverable | Description | Acceptance Criteria |
|-------------|-------------|---------------------|
| API Documentation | Complete OpenAPI 3.0 specification | Validates without errors, includes all endpoints |
| Data Model Schemas | JSON Schema definitions for all models | All required fields defined, relationships clear |
| Integration Specs | Third-party API usage documentation | Authentication, rate limits, error handling defined |

## Risk Factors

| Risk | Impact on Goal | Mitigation Strategy |
|------|----------------|---------------------|
| Incomplete requirements | Missing interfaces discovered later | Multiple review cycles with stakeholders |
| Changing requirements | Interface rework during implementation | Lock interfaces with formal change process |

## Sequence Goal Alignment

Verify that sequence goals support this phase goal:

| Sequence | Sequence Goal | Contribution to Phase Goal |
|----------|---------------|----------------------------|
| 01_api_design | Define all REST API endpoints | Provides 60% of interface definitions |
| 02_data_models | Define all data structures | Provides 30% of interface definitions |
| 03_integration_contracts | Define external integrations | Provides 10% of interface definitions |
