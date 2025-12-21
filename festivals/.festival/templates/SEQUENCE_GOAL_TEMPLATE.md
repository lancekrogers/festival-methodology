---
id: sequence-goal
aliases:
  - sg
description: Defines sequence objective, deliverables, and quality standards
---

# Sequence Goal: [NN_sequence_name]

**Sequence:** [NN_sequence_name] | **Phase:** [NNN_PHASE_NAME] | **Status:** [Planning/Active/Complete] | **Created:** [Date]

## Sequence Objective

**Primary Goal:** [One clear sentence stating what this sequence must accomplish]

**Contribution to Phase Goal:** [How achieving this sequence goal directly supports the phase goal]

## Success Criteria

The sequence goal is achieved when:

### Required Deliverables

- [ ] **[Deliverable 1]**: [Specific output or artifact produced]
- [ ] **[Deliverable 2]**: [Specific output or artifact produced]
- [ ] **[Deliverable 3]**: [Specific output or artifact produced]

### Quality Standards

- [ ] **[Standard 1]**: [Quality measure with specific target]
- [ ] **[Standard 2]**: [Quality measure with specific target]
- [ ] **[Standard 3]**: [Quality measure with specific target]

### Completion Criteria

- [ ] All tasks in sequence completed successfully
- [ ] Quality verification tasks passed
- [ ] Code review completed and issues addressed
- [ ] Documentation updated

## Task Alignment

Verify tasks support this sequence goal:

| Task | Task Objective | Contribution to Sequence Goal |
|------|---------------|------------------------------|
| [01_task] | [Brief objective] | [How it helps achieve sequence goal] |
| [02_task] | [Brief objective] | [How it helps achieve sequence goal] |
| [03_task] | [Brief objective] | [How it helps achieve sequence goal] |

## Dependencies

### Prerequisites (from other sequences)

- [Sequence X]: [What we need from it]
- [Sequence Y]: [What we need from it]

### Provides (to other sequences)

- [What this sequence produces]: Used by [Sequence Z]
- [What this sequence produces]: Used by [Sequence W]

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [Risk description] | Low/Med/High | Low/Med/High | [Prevention strategy] |

## Progress Tracking

### Step Milestones

- [ ] **Step 1**: [Milestone 1 - key deliverable achieved]
- [ ] **Step 2**: [Milestone 2 - key deliverable achieved]
- [ ] **Step 3**: [Milestone 3 - key deliverable achieved]

### Metrics to Monitor

- Task completion rate: [X/Y tasks]
- Quality gate pass rate: [X%]
- Active blockers: [X blocking dependencies]

## Pre-Sequence Checklist

Before starting this sequence:

- [ ] Phase goal understood and aligned
- [ ] Dependencies from other sequences available
- [ ] Resources assigned and available
- [ ] Interfaces/specifications ready (if applicable)

## Post-Completion Evaluation

**Date Completed:** [Date]

**Goal Achievement:** [Fully Achieved/Partially Achieved/Not Achieved]

### Deliverables Assessment

| Deliverable | Status | Quality Score | Notes |
|-------------|--------|--------------|-------|
| [Deliverable 1] | ✅/⚠️/❌ | [1-5] | [Assessment notes] |
| [Deliverable 2] | ✅/⚠️/❌ | [1-5] | [Assessment notes] |
| [Deliverable 3] | ✅/⚠️/❌ | [1-5] | [Assessment notes] |

### What Worked Well

- [Success factor 1]
- [Success factor 2]

### What Could Be Improved

- [Improvement area 1]
- [Improvement area 2]

### Impact on Phase Goal

[Description of how well this sequence's completion supported the overall phase goal]

### Recommendations

- For similar sequences: [Recommendation]
- For next phase: [Recommendation]

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

- [ ] Need another iteration? [Yes/No]
- [ ] If yes, new tasks created: [List task numbers]

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
|------|---------------|------------------------------|
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
