---
id: task-simple
aliases:
  - simple-task
  - ts
description: Streamlined task template for simple, well-understood tasks
---

# Task: [Task_Name]

**Task Number:** [N] | **Parallel Group:** [N or None] | **Dependencies:** [Task numbers or None] | **Autonomy:** [high/medium/low]

## Objective

[ONE clear sentence describing what will be accomplished]

## Requirements

- [ ] [Specific, testable requirement 1]
- [ ] [Specific, testable requirement 2]
- [ ] [Specific, testable requirement 3]

## Implementation Steps

### 1. [First Step]

```bash
# Command or code snippet
```

### 2. [Second Step]

```javascript
// Code implementation
```

### 3. [Third Step]

Description of action to take

## Definition of Done

- [ ] All requirements implemented
- [ ] Tests pass (if applicable)
- [ ] Code reviewed (if applicable)
- [ ] Documentation updated (if applicable)

## Notes

[Any additional context, assumptions, or considerations]

---

## Quick Usage Guide

This simplified template is ideal for:

- Well-understood tasks
- Routine implementations
- Tasks with clear patterns
- When time is critical

For complex tasks requiring detailed planning, use `TASK_TEMPLATE.md` instead.

### Example (Filled Out)

# Task: Create User Model

**Task Number:** 01 | **Parallel Group:** 1 | **Dependencies:** None

## Objective

Create a PostgreSQL user table and Sequelize model with email/password authentication.

## Requirements

- [ ] Create `users` table with id, email, password_hash, created_at, updated_at
- [ ] Create `models/User.js` with Sequelize model definition
- [ ] Add email validation and bcrypt password hashing

## Implementation Steps

### 1. Install Dependencies

```bash
npm install sequelize bcrypt
```

### 2. Create Migration

```bash
npx sequelize-cli migration:generate --name create-users-table
```

### 3. Implement Model

```javascript
// models/User.js
const bcrypt = require('bcrypt');
module.exports = (sequelize, DataTypes) => {
  const User = sequelize.define('User', {
    email: {
      type: DataTypes.STRING,
      unique: true,
      validate: { isEmail: true }
    },
    password_hash: DataTypes.STRING
  });
  
  User.prototype.validatePassword = function(password) {
    return bcrypt.compareSync(password, this.password_hash);
  };
  
  return User;
};
```

## Definition of Done

- [ ] Migration runs successfully
- [ ] Model loads without errors
- [ ] Password hashing works
- [ ] Email validation works

## Notes

Using bcrypt with 10 salt rounds for security. Email uniqueness enforced at database level.
