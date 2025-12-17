---
id: phase-goal
aliases:
  - pg
description: Defines phase objective, success criteria, and quality metrics
---

# Phase Goal: [NNN_PHASE_NAME]

**Phase:** [NNN_PHASE_NAME] | **Status:** [Planning/Active/Complete] | **Created:** [Date] | **Target Completion:** [Date]

## Phase Objective

**Primary Goal:** [One clear sentence stating what this phase must accomplish]

**Context:** [Why this phase is critical to the festival's success and how it enables subsequent phases]

## Success Criteria

The phase goal is achieved when:

### Required Outcomes

- [ ] **[Outcome 1]**: [Specific, measurable deliverable or milestone]
- [ ] **[Outcome 2]**: [Specific, measurable deliverable or milestone]
- [ ] **[Outcome 3]**: [Specific, measurable deliverable or milestone]

### Quality Metrics

- [ ] **[Metric 1]**: [Quantifiable quality measure with target]
- [ ] **[Metric 2]**: [Quantifiable quality measure with target]
- [ ] **[Metric 3]**: [Quantifiable quality measure with target]

### Validation Gates

- [ ] All sequence goals within this phase achieved
- [ ] Stakeholder review and approval completed
- [ ] Documentation updated and complete
- [ ] Next phase prerequisites satisfied

## Key Deliverables

| Deliverable | Description | Acceptance Criteria |
|-------------|-------------|-------------------|
| [Deliverable 1] | [What it is] | [How to verify completion] |
| [Deliverable 2] | [What it is] | [How to verify completion] |
| [Deliverable 3] | [What it is] | [How to verify completion] |

## Risk Factors

| Risk | Impact on Goal | Mitigation Strategy |
|------|---------------|-------------------|
| [Risk 1] | [How it affects phase goal] | [Prevention/response plan] |
| [Risk 2] | [How it affects phase goal] | [Prevention/response plan] |

## Sequence Goal Alignment

Verify that sequence goals support this phase goal:

| Sequence | Sequence Goal | Contribution to Phase Goal |
|----------|--------------|---------------------------|
| [01_sequence] | [Brief goal statement] | [How it helps achieve phase goal] |
| [02_sequence] | [Brief goal statement] | [How it helps achieve phase goal] |
| [03_sequence] | [Brief goal statement] | [How it helps achieve phase goal] |

## Pre-Phase Checklist

Before starting this phase:

- [ ] Previous phase goals fully achieved
- [ ] Resources and team available
- [ ] Dependencies resolved
- [ ] Stakeholders aligned on goals

## Evaluation Framework

### During Execution

Track progress weekly/daily against:

- Sequence completion rate
- Quality metric trends
- Risk emergence
- Blocker resolution time

### Post-Completion Assessment

**Date Completed:** [Date]

**Goal Achievement Score:** [X/Y criteria met]

### What Worked Well

- [Success factor 1]
- [Success factor 2]
- [Success factor 3]

### What Could Be Improved

- [Improvement area 1]
- [Improvement area 2]
- [Improvement area 3]

### Lessons Learned

- [Key learning 1]
- [Key learning 2]
- [Key learning 3]

### Recommendations for Future Phases

- [Recommendation 1]
- [Recommendation 2]
- [Recommendation 3]

## Stakeholder Sign-off

| Stakeholder | Role | Sign-off Date | Notes |
|-------------|------|---------------|-------|
| [Name] | [Role] | [Date] | [Any conditions or notes] |

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

**Phase:** 002_DEFINE_INTERFACES | **Status:** Planning | **Created:** 2024-01-15 | **Target Completion:** 2024-01-22

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
|-------------|-------------|-------------------|
| API Documentation | Complete OpenAPI 3.0 specification | Validates without errors, includes all endpoints |
| Data Model Schemas | JSON Schema definitions for all models | All required fields defined, relationships clear |
| Integration Specs | Third-party API usage documentation | Authentication, rate limits, error handling defined |

## Risk Factors

| Risk | Impact on Goal | Mitigation Strategy |
|------|---------------|-------------------|
| Incomplete requirements | Missing interfaces discovered later | Multiple review cycles with stakeholders |
| Changing requirements | Interface rework during implementation | Lock interfaces with formal change process |

## Sequence Goal Alignment

Verify that sequence goals support this phase goal:

| Sequence | Sequence Goal | Contribution to Phase Goal |
|----------|--------------|---------------------------|
| 01_api_design | Define all REST API endpoints | Provides 60% of interface definitions |
| 02_data_models | Define all data structures | Provides 30% of interface definitions |
| 03_integration_contracts | Define external integrations | Provides 10% of interface definitions |
