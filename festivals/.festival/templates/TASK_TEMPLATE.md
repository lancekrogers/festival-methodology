---
id: TASK_TEMPLATE
aliases:
  - TASK TEMPLATE
  - TASK-TEMPLATE
tags: []
created: '2025-08-26'
modified: '2025-08-26'
---

# Task: [N_Task_Name]

> **Task Number**: N (where N indicates execution order) **Parallel Execution**:
> [Yes/No - if yes, list other tasks with same number] **Dependencies**: [List
> any tasks that must complete before this one]

## Objective

[ONE CLEAR SENTENCE about what will be accomplished with specific deliverables]

**Example**: "Create PostgreSQL user table and Sequelize model with email/password authentication fields"
**Not**: "Implement user management functionality"

## Rules Compliance

Before starting this task, review FESTIVAL_RULES.md, particularly:

- [Relevant section 1]
- [Relevant section 2]
- [Relevant section 3]

## Context

[Why this task is needed, dependencies, background information]

## Requirements

**Use specific, testable requirements with exact names and implementations:**
- [ ] [Create specific file with exact name: `models/User.js`]
- [ ] [Implement specific function: `validateEmail()` with regex pattern]
- [ ] [Add specific database field: `password_hash VARCHAR(255)`]

**Example**:
- [ ] Create `users` table with id, email, password_hash, created_at, updated_at fields
- [ ] Create `models/User.js` with Sequelize model definition and email validation
- [ ] Add bcrypt password hashing with salt rounds = 12
- [ ] Create migration file `migrations/001_create_users.js`

## Deliverables

**List exact files, functions, and outputs that will be created:**
- [Specific file path: `src/models/User.js`]
- [Specific function: `User.authenticate(email, password)`]
- [Specific test file: `tests/models/User.test.js`]
- [Configuration change: Add bcrypt to package.json dependencies]

**Example**:
- `models/User.js` - Sequelize model with email validation and password hashing
- `migrations/001_create_users.js` - Database migration file
- `tests/models/User.test.js` - Unit tests covering authentication and validation
- Updated `package.json` with bcrypt dependency

## Definition of Done

- [ ] [Completion criteria 1]
- [ ] [Quality criteria 1]
- [ ] [Acceptance criteria 1]

## Pre-Task Checklist

- [ ] Read FESTIVAL_RULES.md completely
- [ ] Understand task requirements
- [ ] Review existing code/content and patterns
- [ ] Identify dependencies based on task numbering
- [ ] Verify all lower-numbered tasks in sequence are complete
- [ ] Plan approach

## Implementation Steps

**Provide numbered, actionable steps with exact commands and code:**

### 1. [First Step Title]
```bash
# Exact commands to run
npm install bcrypt sequelize
```

### 2. [Second Step Title]
```javascript
// Actual code to implement
const bcrypt = require('bcrypt');
const { DataTypes } = require('sequelize');
```

### 3. [Third Step Title]
```sql
-- SQL to execute
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL
);
```

## Technical Notes

[Any technical considerations, constraints, or important information]

## Testing Commands

**Provide specific commands to test the implementation:**
```bash
# Run migrations
npx sequelize-cli db:migrate

# Run tests
npm test -- tests/models/User.test.js

# Test manually
node -e "const User = require('./models/User'); console.log('User model loaded');"
```

## Resources

**Link to specific documentation and examples:**
- [Sequelize Model Documentation](https://sequelize.org/docs/v6/core-concepts/model-basics/)
- [bcrypt npm package](https://www.npmjs.com/package/bcrypt)
- [Related example: See TASK_EXAMPLES.md - Example 1: PostgreSQL Table Creation]
- [Code reference: `examples/user-auth/models/User.js`]

## Estimated Effort

[Rough estimate in ideal hours/days]

## Completion Checklist

- [ ] All requirements met
- [ ] Tests pass (if applicable)
- [ ] Documentation updated
- [ ] Quality standards met
- [ ] Security considered (if applicable)
- [ ] Performance assessed (if applicable)
- [ ] Rules compliance verified
- [ ] Self-review completed

## Good vs Bad Examples

### ❌ BAD - Abstract and Vague
```markdown
Objective: Set up user authentication
Requirements:
- [ ] Create user system
- [ ] Add login functionality
- [ ] Handle passwords securely
```

### ✅ GOOD - Specific and Actionable
```markdown
Objective: Create User model with bcrypt authentication and email validation
Requirements:
- [ ] Create `models/User.js` with Sequelize model
- [ ] Implement `User.authenticate(email, password)` method
- [ ] Add email validation with regex: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
- [ ] Hash passwords with bcrypt salt rounds = 12
```

## Notes

**Reference TASK_EXAMPLES.md for 15+ concrete examples of well-written tasks across different domains (database, API, frontend, DevOps, testing).**

[Any additional information, assumptions, or considerations]

## For Verification Tasks (testing_and_verify, code_review)

### Testing Results Location

- Place all testing results in: `./results/testing_results_[timestamp].md`

### Code Review Results Location

- Place all review documents in: `./results/code_review_[timestamp].md`

### Iteration Tasks

- When creating `review_results_update_tasks_iterate_if_needed.md`:
  - Review all documents in `./results/` directory
  - Create new numbered tasks if iteration is needed
  - Document decision to proceed or iterate
