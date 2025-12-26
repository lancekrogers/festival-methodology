# Collaborative Planning Guide

This guide explains how Festival Methodology works as a **goal-oriented collaborative system** between humans and AI agents. Understanding this step-based collaboration model is crucial for successful goal achievement through festival execution.

## Core Collaboration Principle

Festival Methodology is NOT traditional project management focused on time and schedules. Instead, it's a **goal-oriented partnership** where humans and AI agents collaborate to identify and execute the logical steps needed to achieve goals:

- **Humans excel at**: Goal definition, requirements, architectural decisions, step validation, success criteria
- **AI agents excel at**: Step identification, logical progression planning, step structure, execution at unprecedented speed

## The Collaboration Model

### Human Responsibilities

**Goal and Step Definition:**

- Define clear, achievable goals and success criteria
- Specify requirements for goal achievement
- Make architectural and technology decisions
- Set step completion standards and quality criteria
- Provide context for goal importance and constraints

**Step Validation:**

- Document what needs to be achieved
- Provide acceptance criteria for goal completion
- Share external planning documents with step information
- Clarify ambiguous goal requirements
- Validate proposed step progressions

**Goal Progress Guidance:**

- Review AI-generated step progression structures
- Approve or refine logical step sequences
- Guide step priorities for optimal goal achievement
- Address blockers that prevent step completion
- Make decisions about goal scope and changes

### AI Agent Responsibilities

**Step Structure and Progression:**

- Convert goal requirements into logical step sequences
- Break down complex goals into manageable step progressions
- Organize steps for parallel execution toward goal achievement
- Apply consistent step completion verification
- Maintain goal progression documentation

**Step-Based Planning:**

- Create specific step specifications with completion criteria
- Define step dependencies and prerequisites
- Identify logical step order and parallel opportunities
- Structure steps for optimal goal progression
- Generate step tracking and progress documentation

**Goal-Oriented Execution:**

- Execute implementation steps autonomously at unprecedented speed
- Document step completion and progress toward goals
- Request clarification when step requirements are unclear
- Update goal progression tracking
- Maintain context about goal achievement across sessions

## Collaboration Workflows

### Workflow 1: External Planning Available

**Human provides complete requirements:**

```
1. Human: "I have complete specifications for user authentication system"
2. Human: Shares planning documents with specific requirements
3. AI: Reviews documents, asks clarifying questions
4. AI: Structures requirements into festival sequences
5. Human: Reviews and approves festival structure
6. AI: Creates detailed tasks and begins execution
7. Ongoing: Iterative feedback and adjustment
```

### Workflow 2: Planning Phase Needed

**Requirements need to be gathered:**

```
1. Human: "I want to build a user management system"
2. AI: "Let's plan this together. What specific functionality do you need?"
3. Collaborative: Requirements discovery through structured conversation
4. Human: Provides detailed requirements and specifications
5. AI: Structures requirements into festival sequences
6. Human: Reviews and refines structure
7. AI: Creates detailed tasks and begins execution
```

### Workflow 3: Iterative Development  

**Requirements evolve during implementation:**

```
1. Human: Provides initial requirements
2. AI: Creates first implementation sequences
3. AI: Executes initial sequences
4. Human: Reviews results, provides additional requirements
5. AI: Creates next sequences based on learnings
6. Repeat: Iterative cycle of implementation and requirement refinement
```

## Collaboration Boundaries

### What Humans Should Decide

**Never delegate these to AI agents:**

- Overall project vision and goals
- Business requirements and priorities  
- User experience decisions
- Architectural patterns and technology choices
- Security and compliance requirements
- Integration strategies
- Quality standards and acceptance criteria

### What AI Agents Should Structure

**Humans should provide requirements, AI should structure:**

- Task breakdown and sequencing
- Implementation order and dependencies
- Parallel execution opportunities
- Quality gate definition and placement
- Progress tracking and documentation
- Detailed task specifications

### Shared Decision Areas

**Collaborative decisions:**

- Festival phase organization
- Sequence design and boundaries
- Task granularity and detail level
- Testing strategies and approaches
- Review processes and cycles

## Communication Patterns

### Effective Human Communication

**Good Requirements Communication:**

```
✅ "Build JWT authentication with 15-minute access tokens and 7-day refresh tokens. 
   Users login with email/password. Include rate limiting at 5 attempts per minute.
   Must integrate with existing PostgreSQL user table."
```

**Poor Requirements Communication:**

```
❌ "Add some authentication"
❌ "Make it secure"  
❌ "Do what you think is best"
```

**Good Feedback Communication:**

```
✅ "The user authentication sequence looks good, but break task 3 into separate 
   password hashing and validation tasks"
   
✅ "Add a sequence for password reset functionality - I forgot to mention that 
   requirement"
```

### Effective AI Communication

**Good AI Response to Requirements:**

```
✅ "Based on your JWT requirements, I'll create two sequences:

   01_jwt_core/ - Token generation, validation, refresh mechanism
   02_security_features/ - Rate limiting, password hashing, session management
   
   Each sequence will have specific tasks for the functionality you described.
   Does this structure match your vision?"
```

**Poor AI Response:**

```
❌ "I'll create a user system with what I think you need"
❌ "Let me design the authentication architecture for you"
❌ "I'll build a complete user management system"
```

## Common Collaboration Anti-Patterns

### ❌ AI Over-Reach

**Problem:** AI agent tries to make business or architectural decisions

**Example:**

```
Human: "Add user authentication"
AI: "I'll design a microservices architecture with OAuth2, Redis sessions, 
     and three-tier security model"
```

**Solution:** AI should ask for requirements, not make architectural assumptions

### ❌ Human Under-Specification  

**Problem:** Human provides insufficient requirements and expects AI to fill gaps

**Example:**

```
Human: "Build an e-commerce site"
AI: "What specific functionality do you need?"
Human: "Just the standard stuff, you figure it out"
```

**Solution:** Human should provide specific requirements, AI should ask clarifying questions

### ❌ Premature Implementation

**Problem:** Starting implementation before requirements are clear

**Example:**

```
Human: "I'm thinking about user management"
AI: Creates implementation sequences immediately
```

**Solution:** Define requirements first, then structure implementation

### ❌ Scope Drift

**Problem:** Requirements change without festival structure updates

**Example:**

```
Midway through implementation:
Human: "Oh, we also need admin controls and audit logging"
(But festival structure isn't updated)
```

**Solution:** Update festival structure when requirements change

## Successful Collaboration Examples

### Example 1: E-commerce Platform

**Human Input:**

```
"Build e-commerce platform with:
- Product catalog with categories, search, filtering
- Shopping cart with persistent storage
- User accounts with order history
- Payment processing with Stripe integration
- Admin interface for inventory management"
```

**AI Structure:**

```
Festival: E-commerce Platform
├── 001_DEFINE_INTERFACES
│   ├── 01_api_contracts/
│   ├── 02_database_schema/
│   └── 03_integration_specs/
├── 002_IMPLEMENT_CATALOG  
│   ├── 01_product_models/
│   ├── 02_category_system/
│   ├── 03_search_filtering/
├── 003_IMPLEMENT_CART
│   ├── 01_cart_operations/
│   ├── 02_persistence_layer/
├── 004_IMPLEMENT_USERS
│   ├── 01_account_management/
│   ├── 02_order_history/
├── 005_IMPLEMENT_PAYMENTS
│   ├── 01_stripe_integration/
│   ├── 02_payment_flows/
└── 006_IMPLEMENT_ADMIN
    ├── 01_inventory_interface/
    ├── 02_admin_controls/
```

**Collaborative Refinement:**

```
Human: "Actually, let's do user accounts first - cart needs user sessions"
AI: "Updating sequence order: Users → Catalog → Cart → Payments → Admin"
