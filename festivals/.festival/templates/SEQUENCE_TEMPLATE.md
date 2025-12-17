---
id: sequence
aliases:
  - seq
description: Plan a sequence of related tasks within a phase
---

# Sequence: [NN_sequence_name]

**Sequence Number:** [NN] | **Phase:** [Phase Name] | **Parallel Sequences:** [List or None]

> 📋 **Important**: Create a `SEQUENCE_GOAL.md` file in this sequence directory using the SEQUENCE_GOAL_TEMPLATE.md to define specific goals and evaluation criteria for this sequence.

## Objective

[Clear description of what this sequence will accomplish and its role in the phase]

## Context

[Why this sequence is needed, its dependencies on other sequences, and what it enables]

## Success Criteria

- [ ] [Specific measurable outcome 1]
- [ ] [Specific measurable outcome 2]
- [ ] [Specific measurable outcome 3]

## Planned Tasks

### Core Tasks

1. **[01_task_name]** - [Brief description]
   - Dependencies: None
   - Parallel Group: 1

2. **[02_task_name]** - [Brief description]
   - Dependencies: Task 01
   - Parallel Group: None

3. **[03_task_name]** - [Brief description]
   - Dependencies: Task 02
   - Parallel Group: None

### Quality Verification Tasks

4. **[04_testing_and_verify]** - Validate all sequence deliverables
   - Dependencies: All core tasks
   - Parallel Group: None

5. **[05_code_review]** - Review implementation quality
   - Dependencies: Task 04
   - Parallel Group: None

6. **[06_review_results_iterate]** - Address findings and iterate if needed
   - Dependencies: Task 05
   - Parallel Group: None

## Interfaces Produced

[List any interfaces, contracts, or specifications this sequence will define]

- Interface/Contract 1
- Interface/Contract 2

## Dependencies

**Requires from other sequences:**

- [Sequence/Task]: [What is needed]

**Provides to other sequences:**

- [What this sequence produces that others need]

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| [Risk description] | Low/Med/High | Low/Med/High | [Mitigation strategy] |

## Estimated Complexity

**Effort:** [Low/Medium/High]
**Technical Difficulty:** [Low/Medium/High]
**Coordination Required:** [Low/Medium/High]

## Notes

[Additional context, assumptions, constraints, or considerations]

---

## Usage Guide

This template helps plan a sequence of related tasks within a phase. Sequences should:

- Group related functionality together
- Define clear boundaries and interfaces
- Include quality verification tasks
- Enable parallel work where possible

### Example (Filled Out)

# Sequence: 02_user_authentication

**Sequence Number:** 02 | **Phase:** 003_IMPLEMENT | **Parallel Sequences:** [01_database_setup, 03_frontend_auth]

## Objective

Implement complete user authentication system with registration, login, logout, and password reset functionality.

## Context

This sequence implements the authentication interfaces defined in Phase 002. It depends on the database setup from sequence 01 and will provide authentication services to the frontend in sequence 03.

## Success Criteria

- [ ] Users can register with email/password
- [ ] Users can login and receive JWT tokens
- [ ] Password reset flow works end-to-end
- [ ] All endpoints pass security tests

## Planned Tasks

### Core Tasks

1. **01_create_auth_models** - Create User model and migration
   - Dependencies: None
   - Parallel Group: 1

2. **01_setup_jwt_config** - Configure JWT token generation
   - Dependencies: None
   - Parallel Group: 1

3. **02_implement_registration** - Build registration endpoint
   - Dependencies: Task 01 (models)
   - Parallel Group: None

4. **03_implement_login** - Build login endpoint
   - Dependencies: Tasks 01, 02
   - Parallel Group: None

5. **04_implement_password_reset** - Build password reset flow
   - Dependencies: Tasks 01, 02, 03
   - Parallel Group: None

### Quality Verification Tasks

6. **05_testing_and_verify** - Test all auth endpoints
   - Dependencies: All core tasks
   - Parallel Group: None

7. **06_security_review** - Security audit of auth implementation
   - Dependencies: Task 05
   - Parallel Group: None

8. **07_review_results_iterate** - Fix any issues found
   - Dependencies: Task 06
   - Parallel Group: None

## Interfaces Produced

- POST /api/auth/register
- POST /api/auth/login
- POST /api/auth/logout
- POST /api/auth/reset-password
- User model schema

## Dependencies

**Requires from other sequences:**

- 01_database_setup: PostgreSQL connection and migrations system

**Provides to other sequences:**

- 03_frontend_auth: Authentication API endpoints
- 04_user_profile: User model and auth middleware

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Password storage vulnerability | Low | High | Use bcrypt with proper salt rounds |
| JWT token compromise | Medium | High | Implement refresh tokens and expiry |
| Email service failure | Medium | Medium | Add retry logic and queue system |

## Estimated Complexity

**Effort:** Medium
**Technical Difficulty:** Medium
**Coordination Required:** High (interfaces with frontend and database teams)

## Notes

Using bcrypt for password hashing, JWT for stateless auth. Email service needs to be configured separately in DevOps sequence.
