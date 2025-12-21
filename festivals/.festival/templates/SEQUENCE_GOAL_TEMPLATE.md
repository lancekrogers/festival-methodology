---
id: sequence-goal
aliases:
  - sg
description: Defines sequence objective, deliverables, and quality standards
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Sequence Goal: [REPLACE: NN_sequence_name like 02_user_authentication]

**Sequence:** [REPLACE: NN_sequence_name] | **Phase:** [REPLACE: NNN_PHASE_NAME] | **Status:** [REPLACE: Planning/Active/Complete] | **Created:** [REPLACE: Date]

## Sequence Objective

**Primary Goal:** [REPLACE: One clear sentence stating what this sequence must accomplish]

**Contribution to Phase Goal:** [REPLACE: How achieving this sequence goal directly supports the phase goal]

## Success Criteria

The sequence goal is achieved when:

### Required Deliverables

- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]
- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]
- [ ] **[REPLACE: Deliverable name]**: [REPLACE: Specific output or artifact produced]

### Quality Standards

- [ ] **[REPLACE: Standard name]**: [REPLACE: Quality measure with specific target]
- [ ] **[REPLACE: Standard name]**: [REPLACE: Quality measure with specific target]
- [ ] **[REPLACE: Standard name]**: [REPLACE: Quality measure with specific target]

### Completion Criteria

- [ ] All tasks in sequence completed successfully
- [ ] Quality verification tasks passed
- [ ] Code review completed and issues addressed
- [ ] Documentation updated

## Task Alignment

Verify tasks support this sequence goal:

| Task | Task Objective | Contribution to Sequence Goal |
|------|----------------|-------------------------------|
| [REPLACE: 01_task] | [REPLACE: Brief objective] | [REPLACE: How it helps achieve sequence goal] |
| [REPLACE: 02_task] | [REPLACE: Brief objective] | [REPLACE: How it helps achieve sequence goal] |
| [REPLACE: 03_task] | [REPLACE: Brief objective] | [REPLACE: How it helps achieve sequence goal] |

## Dependencies

### Prerequisites (from other sequences)

- [REPLACE: Sequence X]: [REPLACE: What we need from it]
- [REPLACE: Sequence Y]: [REPLACE: What we need from it]

### Provides (to other sequences)

- [REPLACE: What this sequence produces]: Used by [REPLACE: Sequence Z]
- [REPLACE: What this sequence produces]: Used by [REPLACE: Sequence W]

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [REPLACE: Risk description] | [REPLACE: Low/Med/High] | [REPLACE: Low/Med/High] | [REPLACE: Prevention strategy] |

## Progress Tracking

### Step Milestones

- [ ] **Step 1**: [REPLACE: Key deliverable achieved]
- [ ] **Step 2**: [REPLACE: Key deliverable achieved]
- [ ] **Step 3**: [REPLACE: Key deliverable achieved]

### Metrics to Monitor

- Task completion rate: [REPLACE: X/Y tasks]
- Quality gate pass rate: [REPLACE: X%]
- Active blockers: [REPLACE: X blocking dependencies]

## Pre-Sequence Checklist

Before starting this sequence:

- [ ] Phase goal understood and aligned
- [ ] Dependencies from other sequences available
- [ ] Resources assigned and available
- [ ] Interfaces/specifications ready (if applicable)

## Post-Completion Evaluation

**Date Completed:** [REPLACE: Date]

**Goal Achievement:** [REPLACE: Fully Achieved/Partially Achieved/Not Achieved]

### Deliverables Assessment

| Deliverable | Status | Quality Score | Notes |
|-------------|--------|---------------|-------|
| [REPLACE: Deliverable 1] | [REPLACE: Status emoji] | [REPLACE: 1-5] | [REPLACE: Assessment notes] |
| [REPLACE: Deliverable 2] | [REPLACE: Status emoji] | [REPLACE: 1-5] | [REPLACE: Assessment notes] |
| [REPLACE: Deliverable 3] | [REPLACE: Status emoji] | [REPLACE: 1-5] | [REPLACE: Assessment notes] |

### What Worked Well

- [REPLACE: Success factor]
- [REPLACE: Success factor]

### What Could Be Improved

- [REPLACE: Improvement area]
- [REPLACE: Improvement area]

### Impact on Phase Goal

[REPLACE: Description of how well this sequence's completion supported the overall phase goal]

### Recommendations

- For similar sequences: [REPLACE: Recommendation]
- For next phase: [REPLACE: Recommendation]

## Quality Gates

### Testing and Verification

- [ ] All unit tests pass
- [ ] Integration tests complete
- [ ] Performance benchmarks met
- [ ] Security scan clean

### Code Review

- [ ] Code review conducted
- [ ] Review feedback addressed
- [ ] Standards compliance verified

### Iteration Decision

- [ ] Need another iteration? [REPLACE: Yes/No]
- [ ] If yes, new tasks created: [REPLACE: List task numbers]

---

## Usage Guide

This SEQUENCE_GOAL.md file should be:

1. Created when planning the sequence
2. Placed in the sequence directory (e.g., `001_PLAN/01_requirements/SEQUENCE_GOAL.md`)
3. Referenced by all tasks within the sequence
4. Updated during execution to track progress
5. Evaluated upon sequence completion

The goal serves as:

- Clear target for all sequence tasks
- Success measurement criteria
- Dependency tracking document
- Quality assurance checkpoint

### Example (Filled Out)

# Sequence Goal: 02_user_authentication

**Sequence:** 02_user_authentication | **Phase:** 003_IMPLEMENT | **Status:** Active | **Created:** 2024-01-23

## Sequence Objective

**Primary Goal:** Implement complete user authentication system with registration, login, logout, and password reset functionality.

**Contribution to Phase Goal:** This sequence delivers 25% of the Phase 003 implementation goal by providing the foundational authentication layer required by all other features.

## Success Criteria

The sequence goal is achieved when:

### Required Deliverables

- [ ] **Authentication API**: All 6 auth endpoints implemented and tested
- [ ] **User Model**: Database model with secure password storage
- [ ] **JWT System**: Token generation and validation working

### Quality Standards

- [ ] **Security**: Passes OWASP authentication security checklist
- [ ] **Performance**: Login response time < 200ms at 95th percentile
- [ ] **Test Coverage**: Minimum 90% code coverage for auth module

### Completion Criteria

- [ ] All 8 tasks in sequence completed successfully
- [ ] Security review passed
- [ ] Integration tests with frontend successful
- [ ] API documentation complete

## Task Alignment

Verify tasks support this sequence goal:

| Task | Task Objective | Contribution to Sequence Goal |
|------|----------------|-------------------------------|
| 01_create_user_model | Create User database model | Provides data persistence layer |
| 02_implement_registration | Build registration endpoint | Enables user account creation |
| 03_implement_login | Build login endpoint | Enables user authentication |
| 04_implement_jwt | Setup JWT token system | Provides stateless auth mechanism |
| 05_password_reset | Build password reset flow | Completes auth feature set |

## Dependencies

### Prerequisites (from other sequences)

- 01_database_setup: PostgreSQL connection and migration system
- 02_api_framework: Express.js setup with middleware

### Provides (to other sequences)

- Authentication middleware: Used by all protected routes
- User model: Used by profile, permissions, and audit sequences
