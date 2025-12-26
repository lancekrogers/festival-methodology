---
id: TASK_EXAMPLES
aliases:
  - TASK EXAMPLES
  - FESTIVAL TASK EXAMPLES
tags: []
created: '2025-09-06'
modified: '2025-09-06'
---

# Festival Task Examples

## Purpose

This document provides concrete examples of well-written festival tasks that lead to actionable implementation. Use these as models for creating specific, detailed tasks rather than abstract or generic descriptions.

## Good vs Bad Task Examples

### ❌ BAD: Abstract and Generic

```markdown
# Task: 01_user_management.md
## Objective
Implement user management functionality
## Requirements
- [ ] Create user system
- [ ] Add authentication
- [ ] Handle user data
```

### ✅ GOOD: Concrete and Specific

```markdown
# Task: 01_create_user_table_and_model.md
## Objective
Create PostgreSQL user table and Sequelize model with email/password authentication fields

## Requirements
- [ ] Create `users` table with id, email, password_hash, created_at, updated_at
- [ ] Create `models/User.js` with Sequelize model definition
- [ ] Add email validation and password hashing methods
- [ ] Create database migration file `20240101_create_users_table.js`

## Implementation Steps
1. Run: `npx sequelize-cli migration:generate --name create-users-table`
2. Edit migration file with SQL schema
3. Create `models/User.js` with Sequelize model
4. Add bcrypt for password hashing
5. Run: `npx sequelize-cli db:migrate`
6. Test with: `npm test -- --grep "User model"`
```

---

## Database Tasks

### Example 1: PostgreSQL Table Creation

```markdown
# Task: 01_create_posts_table.md

## Objective
Create posts table in PostgreSQL with proper indexes and constraints

## Implementation Steps
```sql
CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
  status post_status DEFAULT 'draft',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TYPE post_status AS ENUM ('draft', 'published', 'archived');
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
```

## Commands to Execute

```bash
psql -d myapp_dev -f migrations/001_create_posts.sql
npm run db:seed -- --file seeds/posts.js
npm test -- tests/models/post.test.js
```

## Deliverables

- [ ] `migrations/001_create_posts.sql` file
- [ ] `seeds/posts.js` with sample data
- [ ] Updated `models/Post.js` Sequelize model

```

### Example 2: MongoDB Collection Setup
```markdown
# Task: 01_setup_users_collection.md

## Objective
Set up MongoDB users collection with validation schema and indexes

## Implementation Steps
```javascript
// In MongoDB shell or setup script
db.createCollection("users", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["email", "passwordHash", "createdAt"],
      properties: {
        email: { bsonType: "string", pattern: "^.+@.+..+$" },
        passwordHash: { bsonType: "string", minLength: 60 },
        profile: {
          bsonType: "object",
          properties: {
            firstName: { bsonType: "string" },
            lastName: { bsonType: "string" },
            avatar: { bsonType: "string" }
          }
        }
      }
    }
  }
});

db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ createdAt: 1 });
```

## Commands to Execute

```bash
mongosh myapp_dev --eval "load('scripts/setup_users_collection.js')"
node scripts/seed_users.js
npm test -- tests/models/user.integration.test.js
```

```

---

## API Implementation Tasks

### Example 3: REST API Endpoint
```markdown
# Task: 01_implement_user_registration_endpoint.md

## Objective
Create POST /api/users endpoint for user registration with validation and password hashing

## Implementation Steps

### 1. Create Route Handler
```javascript
// routes/users.js
const express = require('express');
const bcrypt = require('bcrypt');
const { User } = require('../models');
const { validateRegistration } = require('../middleware/validation');

router.post('/', validateRegistration, async (req, res) => {
  try {
    const { email, password, firstName, lastName } = req.body;
    
    // Check if user exists
    const existingUser = await User.findOne({ where: { email } });
    if (existingUser) {
      return res.status(409).json({ error: 'Email already registered' });
    }
    
    // Hash password
    const saltRounds = 12;
    const passwordHash = await bcrypt.hash(password, saltRounds);
    
    // Create user
    const user = await User.create({
      email,
      passwordHash,
      profile: { firstName, lastName }
    });
    
    res.status(201).json({
      id: user.id,
      email: user.email,
      profile: user.profile,
      createdAt: user.createdAt
    });
  } catch (error) {
    res.status(500).json({ error: 'Registration failed' });
  }
});
```

### 2. Create Validation Middleware

```javascript
// middleware/validation.js
const { body } = require('express-validator');

const validateRegistration = [
  body('email').isEmail().normalizeEmail(),
  body('password').isLength({ min: 8 }).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/),
  body('firstName').trim().isLength({ min: 1, max: 50 }),
  body('lastName').trim().isLength({ min: 1, max: 50 })
];
```

## Testing Commands

```bash
# Test the endpoint
curl -X POST http://localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "StrongPass123",
    "firstName": "John",
    "lastName": "Doe"
  }'

# Run tests
npm test -- tests/routes/users.test.js
```

## Deliverables

- [ ] `routes/users.js` with POST endpoint
- [ ] `middleware/validation.js` with registration validation
- [ ] `tests/routes/users.test.js` with endpoint tests

```

### Example 4: GraphQL Resolver
```markdown
# Task: 01_create_user_mutations.md

## Objective
Implement GraphQL mutations for user registration and login

## Implementation Steps

### 1. Define GraphQL Schema
```graphql
# schema/user.graphql
type User {
  id: ID!
  email: String!
  profile: UserProfile!
  createdAt: String!
}

type UserProfile {
  firstName: String!
  lastName: String!
  avatar: String
}

input RegisterInput {
  email: String!
  password: String!
  firstName: String!
  lastName: String!
}

type Mutation {
  registerUser(input: RegisterInput!): AuthPayload!
  loginUser(email: String!, password: String!): AuthPayload!
}

type AuthPayload {
  token: String!
  user: User!
}
```

### 2. Implement Resolvers

```javascript
// resolvers/user.js
const bcrypt = require('bcrypt');
const jwt = require('jsonwebtoken');
const { User } = require('../models');

const userMutations = {
  registerUser: async (_, { input }) => {
    const { email, password, firstName, lastName } = input;
    
    const existingUser = await User.findOne({ where: { email } });
    if (existingUser) throw new Error('Email already registered');
    
    const passwordHash = await bcrypt.hash(password, 12);
    const user = await User.create({
      email,
      passwordHash,
      profile: { firstName, lastName }
    });
    
    const token = jwt.sign({ userId: user.id }, process.env.JWT_SECRET, { expiresIn: '7d' });
    
    return { token, user };
  }
};
```

## Testing Commands

```bash
# Test GraphQL mutation
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { registerUser(input: { email: \"test@example.com\", password: \"StrongPass123\", firstName: \"John\", lastName: \"Doe\" }) { token user { id email } } }"
  }'

npm test -- tests/resolvers/user.test.js
```

```

---

## Frontend Component Tasks

### Example 5: React Component with Hooks
```markdown
# Task: 01_create_login_form_component.md

## Objective
Create LoginForm React component with form validation and authentication integration

## Implementation Steps

### 1. Create Component File
```jsx
// components/LoginForm.jsx
import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { validateEmail } from '../utils/validation';

const LoginForm = ({ onSuccess, onError }) => {
  const [formData, setFormData] = useState({ email: '', password: '' });
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErrors({});

    // Validation
    const newErrors = {};
    if (!validateEmail(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }
    if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setLoading(false);
      return;
    }

    try {
      const result = await login(formData.email, formData.password);
      onSuccess(result);
    } catch (error) {
      setErrors({ submit: error.message });
      onError(error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="login-form">
      <div className="form-group">
        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          value={formData.email}
          onChange={(e) => setFormData({ ...formData, email: e.target.value })}
          className={errors.email ? 'error' : ''}
          required
        />
        {errors.email && <span className="error-message">{errors.email}</span>}
      </div>

      <div className="form-group">
        <label htmlFor="password">Password</label>
        <input
          id="password"
          type="password"
          value={formData.password}
          onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          className={errors.password ? 'error' : ''}
          required
        />
        {errors.password && <span className="error-message">{errors.password}</span>}
      </div>

      {errors.submit && <div className="error-message">{errors.submit}</div>}

      <button type="submit" disabled={loading}>
        {loading ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
};

export default LoginForm;
```

### 2. Create Styling

```css
/* components/LoginForm.css */
.login-form {
  max-width: 400px;
  margin: 2rem auto;
  padding: 2rem;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}

.form-group input.error {
  border-color: #dc3545;
}

.error-message {
  color: #dc3545;
  font-size: 0.875rem;
  margin-top: 0.25rem;
}
```

## Testing Commands

```bash
# Run component tests
npm test -- LoginForm.test.jsx

# Run Storybook
npm run storybook

# Test in browser
npm start
```

## Deliverables

- [ ] `components/LoginForm.jsx` component
- [ ] `components/LoginForm.css` styling
- [ ] `components/__tests__/LoginForm.test.jsx` test file
- [ ] `components/LoginForm.stories.jsx` Storybook story

```

---

## DevOps and Configuration Tasks

### Example 6: Docker Configuration
```markdown
# Task: 01_create_dockerfile_for_node_app.md

## Objective
Create production-ready Dockerfile for Node.js application with multi-stage build

## Implementation Steps

### 1. Create Dockerfile
```dockerfile
# Dockerfile
FROM node:18-alpine as build

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY . .
RUN npm run build

# Production stage
FROM node:18-alpine as production

RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001

WORKDIR /app

# Copy built application
COPY --from=build --chown=nodejs:nodejs /app/dist ./dist
COPY --from=build --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=build --chown=nodejs:nodejs /app/package.json ./package.json

USER nodejs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

CMD ["npm", "start"]
```

### 2. Create .dockerignore

```
node_modules
npm-debug.log
.git
.gitignore
README.md
.env
.nyc_output
coverage
.cache
dist
```

### 3. Create docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      target: production
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgres://user:pass@db:5432/myapp
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

## Commands to Execute

```bash
# Build image
docker build -t myapp:latest .

# Run with compose
docker-compose up -d

# Check logs
docker-compose logs -f app

# Run tests in container
docker-compose exec app npm test
```

## Deliverables

- [ ] `Dockerfile` with multi-stage build
- [ ] `.dockerignore` file
- [ ] `docker-compose.yml` for development
- [ ] `docker-compose.prod.yml` for production

```

### Example 7: CI/CD Pipeline
```markdown
# Task: 01_setup_github_actions_ci.md

## Objective
Set up GitHub Actions CI pipeline with testing, linting, and deployment

## Implementation Steps

### 1. Create GitHub Actions Workflow
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run linting
      run: npm run lint

    - name: Run type check
      run: npm run type-check

    - name: Run tests
      run: npm test
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/test_db

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Setup Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Login to DockerHub
      uses: docker/login-action@v3
      with:
        username: ${{ secrets.DOCKERHUB_USERNAME }}
        password: ${{ secrets.DOCKERHUB_TOKEN }}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: |
          myapp/myapp:latest
          myapp/myapp:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Deploy to production
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: ${{ secrets.HOST }}
        username: ${{ secrets.USERNAME }}
        key: ${{ secrets.SSH_KEY }}
        script: |
          docker pull myapp/myapp:latest
          docker-compose -f /opt/myapp/docker-compose.prod.yml up -d
```

## Commands to Test Locally

```bash
# Install act (GitHub Actions local runner)
brew install act  # or curl https://raw.githubusercontent.com/nektos/act/master/install.sh | bash

# Test workflow locally
act -j test

# Test specific job
act push -j build
```

## Deliverables

- [ ] `.github/workflows/ci.yml` pipeline file
- [ ] Updated `package.json` with lint and type-check scripts
- [ ] `codecov.yml` configuration file

```

---

## Testing Tasks

### Example 8: Unit Tests with Jest
```markdown
# Task: 01_create_user_service_tests.md

## Objective
Create comprehensive unit tests for UserService class using Jest and mocking

## Implementation Steps

### 1. Create Test File
```javascript
// tests/services/UserService.test.js
const UserService = require('../../services/UserService');
const { User } = require('../../models');
const bcrypt = require('bcrypt');

jest.mock('../../models');
jest.mock('bcrypt');

describe('UserService', () => {
  let userService;
  
  beforeEach(() => {
    userService = new UserService();
    jest.clearAllMocks();
  });

  describe('createUser', () => {
    it('should create user with hashed password', async () => {
      // Arrange
      const userData = {
        email: 'test@example.com',
        password: 'plainPassword',
        firstName: 'John',
        lastName: 'Doe'
      };
      
      const hashedPassword = 'hashedPassword123';
      const createdUser = {
        id: 1,
        email: userData.email,
        passwordHash: hashedPassword,
        profile: { firstName: 'John', lastName: 'Doe' }
      };

      User.findOne.mockResolvedValue(null);
      bcrypt.hash.mockResolvedValue(hashedPassword);
      User.create.mockResolvedValue(createdUser);

      // Act
      const result = await userService.createUser(userData);

      // Assert
      expect(User.findOne).toHaveBeenCalledWith({ where: { email: userData.email } });
      expect(bcrypt.hash).toHaveBeenCalledWith(userData.password, 12);
      expect(User.create).toHaveBeenCalledWith({
        email: userData.email,
        passwordHash: hashedPassword,
        profile: { firstName: 'John', lastName: 'Doe' }
      });
      expect(result).toEqual(createdUser);
    });

    it('should throw error if email already exists', async () => {
      // Arrange
      const userData = { email: 'existing@example.com', password: 'pass' };
      User.findOne.mockResolvedValue({ id: 1, email: userData.email });

      // Act & Assert
      await expect(userService.createUser(userData))
        .rejects.toThrow('Email already registered');
      
      expect(User.create).not.toHaveBeenCalled();
    });
  });

  describe('authenticateUser', () => {
    it('should return user when credentials are valid', async () => {
      // Arrange
      const email = 'test@example.com';
      const password = 'correctPassword';
      const user = { 
        id: 1, 
        email, 
        passwordHash: 'hashedPassword',
        profile: { firstName: 'John' }
      };

      User.findOne.mockResolvedValue(user);
      bcrypt.compare.mockResolvedValue(true);

      // Act
      const result = await userService.authenticateUser(email, password);

      // Assert
      expect(User.findOne).toHaveBeenCalledWith({ 
        where: { email },
        include: ['profile']
      });
      expect(bcrypt.compare).toHaveBeenCalledWith(password, user.passwordHash);
      expect(result).toEqual(user);
    });
  });
});
```

### 2. Create Integration Test

```javascript
// tests/integration/user.integration.test.js
const request = require('supertest');
const app = require('../../app');
const { User } = require('../../models');

describe('User Integration Tests', () => {
  beforeEach(async () => {
    await User.destroy({ where: {}, truncate: true });
  });

  describe('POST /api/users', () => {
    it('should register a new user', async () => {
      const userData = {
        email: 'newuser@example.com',
        password: 'StrongPass123',
        firstName: 'Jane',
        lastName: 'Smith'
      };

      const response = await request(app)
        .post('/api/users')
        .send(userData)
        .expect(201);

      expect(response.body).toMatchObject({
        email: userData.email,
        profile: {
          firstName: userData.firstName,
          lastName: userData.lastName
        }
      });
      expect(response.body).toHaveProperty('id');
      expect(response.body).not.toHaveProperty('passwordHash');
    });
  });
});
```

## Commands to Execute

```bash
# Run unit tests
npm test -- tests/services/UserService.test.js

# Run integration tests
npm run test:integration

# Run with coverage
npm run test:coverage

# Watch mode
npm test -- --watch
```

## Deliverables

- [ ] `tests/services/UserService.test.js` unit test file
- [ ] `tests/integration/user.integration.test.js` integration tests
- [ ] Updated `jest.config.js` configuration
- [ ] Mock data factories in `tests/factories/`

```

---

## Key Principles for Good Tasks

### 1. **Be Specific and Concrete**
- Use exact file names, function names, variable names
- Include actual code snippets and commands
- Specify exact directory structures and file locations

### 2. **Include Implementation Steps**
- Break down the work into numbered, actionable steps
- Provide the actual code or configuration to implement
- Include command-line instructions where needed

### 3. **Provide Testing Instructions**
- Include specific commands to test the implementation
- Show expected outputs or behaviors
- Include both automated tests and manual testing steps

### 4. **Specify Deliverables Clearly**
- List exact files that should be created or modified
- Include file paths and naming conventions
- Specify what should be committed to version control

### 5. **Include Real Examples**
- Use concrete data, names, and values instead of placeholders
- Show realistic API responses, database records, UI states
- Provide copy-pasteable code that works out of the box

### 6. **Consider Dependencies and Context**
- Explain how this task connects to other parts of the system
- List prerequisites and what should be completed first
- Include error handling and edge cases

---

## Template for Creating New Task Examples

When adding new examples to this document, use this structure:

```markdown
### Example X: [Task Type Description]
```markdown
# Task: 01_specific_task_name.md

## Objective
[One clear sentence about what will be accomplished]

## Implementation Steps
[Numbered list of specific actions with code/commands]

### 1. [Step Title]
```[language]
[Actual code to implement]
```

## Commands to Execute

```bash
[Exact commands to run]
```

## Deliverables

- [ ] [Specific file or output that will be created]

```
```

This structure ensures every task example provides concrete, actionable guidance rather than abstract descriptions.
