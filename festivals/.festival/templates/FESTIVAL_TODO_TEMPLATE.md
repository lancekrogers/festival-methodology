---
id: festival-todo
aliases:
  - todo
  - progress
  - tracking
description: Unified project tracking template with checkboxes for manual progress tracking
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Festival TODO - Unified Project Tracking

## Festival: [REPLACE: Festival Name]

**Goal**: [REPLACE: Primary objective in one clear sentence]
**Status**: [REPLACE: Planning | Active | Review | Complete]
**Started**: [REPLACE: YYYY-MM-DD]
**Target**: [REPLACE: YYYY-MM-DD]

## Status Legend

- [ ] Not Started
- [x] In Progress
- [x] Completed
- [x] Blocked
- [x] Needs Review
- [x] Ready (dependencies met)

---

## Festival Progress Overview

### Phase Completion Status

- [ ] **001_PLAN** - Requirements and Architecture
- [ ] **002_DEFINE_INTERFACES** - System Contracts (Critical Gate)
- [ ] **003_IMPLEMENT** - Build Solution
- [ ] **004_REVIEW_AND_UAT** - User Acceptance

### Current Work Status

```
Active Phase: [REPLACE: Phase name]
Active Sequences: [REPLACE: Sequence names]
Active Tasks: [REPLACE: Task names]
Blockers: [REPLACE: List blockers or None]
Next Critical Gate: [REPLACE: Gate description]
```

---

## PHASE 001: PLAN

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Purpose**: Define requirements, architecture, and draft initial interfaces
**Gate Criteria**: Requirements approved, architecture documented, feasibility confirmed

### Sequence Progress

- [ ] **01_requirements_analysis** (Foundation for all other work)
- [ ] **02_architecture_design** (System blueprint)
- [ ] **03_feasibility_study** (Risk and resource validation)

#### 01_requirements_analysis

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]

**Tasks**:

- [ ] 01_user_research.md
- [ ] 01_security_requirements.md *(parallel)*
- [ ] 02_requirements_spec.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 02_architecture_design

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 01_requirements_analysis must be completed

**Tasks**:

- [ ] 01_system_architecture.md
- [ ] 01_technology_selection.md *(parallel)*
- [ ] 02_security_architecture.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 03_feasibility_study

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 02_architecture_design must be completed

**Tasks**:

- [ ] 01_technical_feasibility.md
- [ ] 01_resource_assessment.md *(parallel)*
- [ ] 02_risk_analysis.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

---

## PHASE 002: DEFINE_INTERFACES (CRITICAL GATE)

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Purpose**: Lock all system contracts, APIs, and data models before implementation
**Gate Criteria**: ALL interfaces FINALIZED, stakeholder sign-offs complete, no Phase 003 work until complete

### Sequence Progress

- [ ] **01_api_contracts** (External system interfaces)
- [ ] **02_data_schemas** (Data structure contracts)
- [ ] **03_integration_points** (Service boundaries)

#### 01_api_contracts

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]

**Tasks**:

- [ ] 01_rest_api_spec.md
- [ ] 01_graphql_schema.md *(parallel)*
- [ ] 02_error_handling_spec.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 02_data_schemas

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Can run parallel with**: 01_api_contracts

**Tasks**:

- [ ] 01_domain_models.md
- [ ] 01_database_schema.md *(parallel)*
- [ ] 02_validation_rules.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 03_integration_points

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 01_api_contracts and 02_data_schemas must be completed

**Tasks**:

- [ ] 01_external_services.md
- [ ] 01_event_contracts.md *(parallel)*
- [ ] 02_authentication_flow.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

---

## PHASE 003: IMPLEMENT

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Purpose**: Build solution based on locked interfaces with parallel development
**Gate Criteria**: All implementation complete, automated tests passing, interfaces maintained

### Sequence Progress

- [ ] **01_backend_foundation** (Core services and data)
- [ ] **02_frontend_integration** (User interface)
- [ ] **03_service_integration** (External connections)
- [ ] **04_performance_optimization** (Scale and efficiency)

#### 01_backend_foundation

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]

**Tasks**:

- [ ] 01_database_setup.md
- [ ] 01_api_endpoints.md *(parallel)*
- [ ] 01_business_logic.md *(parallel)*
- [ ] 02_automated_testing.md
- [ ] 03_code_review.md
- [ ] 04_review_results_iterate.md

#### 02_frontend_integration

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Can run parallel with**: 01_backend_foundation (thanks to interface contracts)

**Tasks**:

- [ ] 01_ui_components.md
- [ ] 01_state_management.md *(parallel)*
- [ ] 02_api_integration.md
- [ ] 03_automated_testing.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 03_service_integration

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 01_backend_foundation must be completed

**Tasks**:

- [ ] 01_external_apis.md
- [ ] 01_message_queues.md *(parallel)*
- [ ] 02_error_handling.md
- [ ] 03_automated_testing.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 04_performance_optimization

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 01_backend_foundation and 02_frontend_integration completed

**Tasks**:

- [ ] 01_database_optimization.md
- [ ] 01_caching_strategy.md *(parallel)*
- [ ] 02_load_testing.md
- [ ] 03_code_review.md
- [ ] 04_review_results_iterate.md

---

## PHASE 004: REVIEW_AND_UAT

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Purpose**: Validate with users and stakeholders, ensure production readiness
**Gate Criteria**: User acceptance criteria met, stakeholder approval, production ready

### Sequence Progress

- [ ] **01_user_acceptance_testing** (Real user validation)
- [ ] **02_stakeholder_review** (Business validation)
- [ ] **03_production_readiness** (Deployment preparation)

#### 01_user_acceptance_testing

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]

**Tasks**:

- [ ] 01_uat_planning.md
- [ ] 01_test_scenarios.md *(parallel)*
- [ ] 02_user_testing_execution.md
- [ ] 03_feedback_analysis.md
- [ ] 04_testing_and_verify.md
- [ ] 05_code_review.md
- [ ] 06_review_results_iterate.md

#### 02_stakeholder_review

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Can run parallel with**: 01_user_acceptance_testing

**Tasks**:

- [ ] 01_business_validation.md
- [ ] 01_stakeholder_demos.md *(parallel)*
- [ ] 02_requirements_sign_off.md
- [ ] 03_testing_and_verify.md
- [ ] 04_code_review.md
- [ ] 05_review_results_iterate.md

#### 03_production_readiness

**Status**: [REPLACE: Not Started | In Progress | Completed | Blocked]
**Dependencies**: 01_user_acceptance_testing and 02_stakeholder_review completed

**Tasks**:

- [ ] 01_deployment_checklist.md
- [ ] 01_monitoring_setup.md *(parallel)*
- [ ] 02_runbook_creation.md
- [ ] 03_final_review.md
- [ ] 04_go_live_approval.md

---

## Progress Dashboard

### Overall Metrics

```
Festival Progress: [_______________] 0%

Phase Breakdown:
001_PLAN:              [_______________] 0/3 sequences
002_DEFINE_INTERFACES: [_______________] 0/3 sequences
003_IMPLEMENT:         [_______________] 0/4 sequences
004_REVIEW_AND_UAT:    [_______________] 0/3 sequences

Total: 0/13 sequences completed (0%)
Tasks: 0/72 completed (0%)
```

### Current Focus

```
Active Steps:
[REPLACE: Task name] - [REPLACE: Owner] - [REPLACE: Dependencies]
[REPLACE: Task name] - [REPLACE: Owner] - [REPLACE: Dependencies]
[REPLACE: Task name] - [REPLACE: Owner] - [REPLACE: Dependencies]

Recently Completed:
[REPLACE: Task name] - [REPLACE: Owner]
[REPLACE: Task name] - [REPLACE: Owner]

Blocked/At Risk:
[REPLACE: Task name] - [REPLACE: Blocker description]
[REPLACE: Task name] - [REPLACE: Review needed]
```

---

## Critical Dependencies & Gates

### Phase Gates (Must Complete Before Next Phase)

1. **001 → 002**: [ ] Requirements documented, [ ] Architecture approved, [ ] Team aligned
2. **002 → 003**: [ ] ALL interfaces FINALIZED, [ ] Stakeholder sign-offs, [ ] COMMON_INTERFACES.md status = FINALIZED
3. **003 → 004**: [ ] Implementation complete, [ ] Tests passing, [ ] Integration working
4. **004 → DONE**: [ ] User acceptance passed, [ ] Production deployment ready

### External Dependencies

```
Waiting For:
[REPLACE: External system/person] - [REPLACE: What needed] - [REPLACE: Priority]
[REPLACE: External system/person] - [REPLACE: What needed] - [REPLACE: Priority]

Provides To Others:
[REPLACE: System/person] depends on our [REPLACE: deliverable] - [REPLACE: Status]
```

---

## Blockers & Risks

### Active Blockers

```
BLOCKER_001: [REPLACE: Description]
   Impact: [REPLACE: How this blocks progress]
   Owner: [REPLACE: Who is resolving]
   Next Step: [REPLACE: What needs to happen to unblock]

BLOCKER_002: [REPLACE: Description]
   Impact: [REPLACE: How this blocks progress]
   Owner: [REPLACE: Who is resolving]
   Next Step: [REPLACE: What needs to happen to unblock]
```

### Risk Register

```
HIGH: [REPLACE: Risk description] - Mitigation: [REPLACE: Strategy]
MED:  [REPLACE: Risk description] - Mitigation: [REPLACE: Strategy]
LOW:  [REPLACE: Risk description] - Mitigation: [REPLACE: Strategy]
```

---

## Decision Log

### Recent Decisions

```
[REPLACE: YYYY-MM-DD] DECISION: [REPLACE: What was decided]
  Rationale: [REPLACE: Why this decision]
  Impact: [REPLACE: How this affects festival]
  Made by: [REPLACE: Person/Agent]

[REPLACE: YYYY-MM-DD] DECISION: [REPLACE: What was decided]
  Rationale: [REPLACE: Why this decision]
  Impact: [REPLACE: How this affects festival]
  Made by: [REPLACE: Person/Agent]
```

---

## Usage Instructions

### How to Use This Festival TODO System

**1. Daily Updates**

- Update task checkboxes as work progresses: [ ] → [x] In Progress → [x] Completed
- Update sequence status when all tasks in sequence complete
- Update phase status when all sequences in phase complete
- Note any blockers immediately with [x] Blocked

**2. Progress Reviews**

- Review overall progress metrics
- Update active steps section
- Assess risks and dependencies
- Identify next priority steps

**3. Phase Gate Reviews**

- Before moving to next phase, ensure ALL criteria met
- **Phase 002 → 003 is CRITICAL**: No implementation until interfaces are FINALIZED
- Document gate decisions in decision log

**4. Status Meanings**

- **Phase Status**: Overall phase health and completion
- **Sequence Status**: All tasks in sequence completed and reviewed
- **Task Status**: Individual deliverable completion

### Methodology Reminders

- **Interface-First**: Phase 002 gates all implementation
- **Parallel Work**: Tasks with same numbers (01_, 01_) can run simultaneously
- **Quality Gates**: Every sequence ends with test → review → iterate
- **Step-Based Progress**: Track completed items, not time
- **Three-Level Tracking**: Phase → Sequence → Task status all matter

### Automation Opportunities

This markdown format enables:

- **CI/CD Integration**: Parse checkboxes for automated reporting
- **Progress Tracking**: Calculate completion percentages
- **Dependency Management**: Validate prerequisite completion
- **Status Dashboards**: Convert to visual project boards
- **Risk Monitoring**: Alert on blocked items or missed gates

---

**Last Updated**: [REPLACE: YYYY-MM-DD HH:MM]
**Updated By**: [REPLACE: Agent/User Name]
**Next Review**: [REPLACE: YYYY-MM-DD]

**Festival Methodology Compliance**: This TODO system implements the three-level hierarchy (Phases → Sequences → Tasks) with proper gate controls and interface-first development principles.
