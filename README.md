# Festival Methodology

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Development Status](https://img.shields.io/badge/status-pre--release-orange)](CHANGELOG.md)

A goal-based methodology that helps you **collaboratively create actionable tasks** for AI agents to execute in long-running autonomous sessions. Festival transforms high-level objectives into structured, executable work that AI can complete independently.

## What Festival Does

Festival bridges the gap between what you want to build and what AI agents can actually execute:

```mermaid
graph LR
    subgraph "Your Input"
        G[Goal: Build Auth System]
    end

    subgraph "Festival Process"
        G --> C[Collaborative Planning]
        C --> S[Structured Tasks]
        S --> V[Validated Plan]
    end

    subgraph "AI Execution"
        V --> E1[Agent 1: Research]
        V --> E2[Agent 2: Design]
        V --> E3[Agent 3: Build]
        E1 --> D[✓ Delivered System]
        E2 --> D
        E3 --> D
    end

    style G fill:#e1f5fe
    style D fill:#c8e6c9
    style C fill:#f5f5f5
```

## Core Benefits

Festival enables:

- **Long-running autonomous builds** - AI agents work for hours or days, not minutes
- **Collaborative task creation** - You and AI work together to break down goals
- **Executable specifications** - Every task includes concrete steps AI can follow
- **Persistent context** - Knowledge accumulates across work sessions
- **Parallel execution** - Multiple agents work simultaneously on different parts

## How It Works: From Goal to Execution

### 1. Start with a Goal

You define what you want to achieve - a complete feature, system, or product.

### 2. Collaborative Planning

Festival helps you and AI agents break the goal into phases, sequences, and tasks.

### 3. Actionable Task Creation

Each task becomes a detailed specification with:

- Clear objectives
- Concrete implementation steps
- Specific deliverables
- Validation criteria

### 4. Autonomous Execution

AI agents execute tasks independently, maintaining context and building toward the goal.

## The Three-Level Structure

```mermaid
graph TD
    G[Goal: E-Commerce Platform]
    G --> P1[Phase 1: Planning]
    G --> P2[Phase 2: Design]
    G --> P3[Phase 3: Implementation]

    P3 --> S1[Sequence 3.1: Backend]
    P3 --> S2[Sequence 3.2: Frontend]

    S1 --> T1[Task: Create User API]
    S1 --> T2[Task: Build Auth Service]
    S1 --> T3[Task: Setup Database]

    style G fill:#e3f2fd
    style P1,P2,P3 fill:#fff3e0
    style S1,S2 fill:#f3e5f5
    style T1,T2,T3 fill:#f5f5f5
```

- **Goal**: The outcome you want to achieve
- **Phases**: Major stages of work (planning, design, implementation, validation)
- **Sequences**: Related tasks that must complete in order
- **Tasks**: Concrete, executable work items with full specifications

## Creating Actionable Tasks

Festival tasks aren't vague descriptions - they're complete specifications AI can execute:

```markdown
# Task: 01_implement_user_authentication.md

## Goal Context

Building authentication for the e-commerce platform to enable user accounts

## Objective

Create JWT-based authentication with email/password login

## Requirements

- [ ] User registration endpoint
- [ ] Login with email/password
- [ ] JWT token generation (15min access, 7day refresh)
- [ ] Password hashing with bcrypt
- [ ] Rate limiting (5 attempts/minute)

## Implementation Steps

1. Install dependencies:
   npm install jsonwebtoken bcrypt express-rate-limit

2. Create database schema:

   - users table (id, email, password_hash, created_at)
   - refresh_tokens table (token, user_id, expires_at)

3. Implement endpoints:

   - POST /api/auth/register
   - POST /api/auth/login
   - POST /api/auth/refresh
   - POST /api/auth/logout

4. Add middleware:
   - Authentication verification
   - Rate limiting
   - Input validation

## Validation

- Test registration with: curl -X POST localhost:3000/api/auth/register ...
- Verify JWT expiration times
- Check rate limiting blocks after 5 attempts
- Ensure passwords are hashed, not plain text

## Deliverables

- [ ] src/routes/auth.js - Authentication endpoints
- [ ] src/middleware/auth.js - JWT verification
- [ ] src/models/User.js - User model with password hashing
- [ ] tests/auth.test.js - Complete test coverage
```

This level of detail enables AI agents to work autonomously without constant clarification.

## Long-Running Autonomous Execution

```mermaid
gantt
    title AI Agents Building Authentication System
    dateFormat HH:mm
    axisFormat %H:%M

    section Research Phase
    Analyze requirements     :done, research1, 00:00, 2h
    Research best practices  :done, research2, after research1, 1h
    Document findings        :done, research3, after research2, 1h

    section Design Phase
    Design API contracts     :active, design1, after research3, 2h
    Create data schemas      :active, design2, after design1, 1h
    Plan architecture        :active, design3, after design2, 2h

    section Build Phase
    Implement backend        :build1, after design3, 8h
    Create frontend          :build2, after design3, 6h
    Write tests              :build3, after design3, 4h

    section Validate
    Integration testing      :validate1, after build1, 2h
    Security audit           :validate2, after validate1, 1h
    Final review             :validate3, after validate2, 1h
```

AI agents work continuously, moving through phases autonomously while you review at checkpoints.

## The Collaborative Process

```mermaid
sequenceDiagram
    participant You
    participant Festival
    participant AI Agent
    participant Filesystem

    You->>Festival: Define goal
    Festival->>You: Suggest structure
    You->>Festival: Refine phases
    Festival->>AI Agent: Generate tasks
    AI Agent->>Filesystem: Create task files
    You->>Filesystem: Review & adjust
    You->>AI Agent: Execute tasks
    AI Agent->>AI Agent: Work autonomously
    AI Agent->>Filesystem: Deliver results
    You->>Filesystem: Validate completion
```

## Getting Started

### 1. Install Festival Structure

```bash
cp -r festivals/ /your/workspace/
cd /your/workspace/festivals/
```

### 2. Define Your Goal

Create `FESTIVAL_OVERVIEW.md` with:

- Clear objective
- Success criteria
- Constraints
- Key features

### 3. Use Planning Agent

The Festival planning agent helps structure your project:

```bash
# Point AI to the planning agent
festivals/.festival/agents/festival_planning_agent.md
```

### 4. Collaborate on Task Creation

Work with AI to create detailed, actionable tasks in the filesystem structure.

### 5. Launch Autonomous Execution

AI agents read tasks and work independently to achieve the goal.

## Festival vs Other Approaches

| Aspect             | Festival                   | Traditional PM | Ad-hoc AI          |
| ------------------ | -------------------------- | -------------- | ------------------ |
| **Focus**          | Goal achievement via tasks | Task tracking  | Quick answers      |
| **Task Detail**    | Complete executable specs  | User stories   | Vague prompts      |
| **Execution Time** | Hours to days              | Sprint cycles  | Minutes            |
| **Context**        | Persists in filesystem     | Meeting notes  | Lost between chats |
| **AI Autonomy**    | Full autonomous execution  | N/A            | Constant prompting |
| **Collaboration**  | Human-AI task creation     | Human teams    | Human directs      |

## Directory Structure

```
festivals/
├── active/                     # Current projects
│   └── auth_system/
│       ├── FESTIVAL_OVERVIEW.md       # Goal & success criteria
│       ├── 001_PLAN/                  # Research & requirements
│       │   ├── 01_requirements/       # Requirement gathering
│       │   └── 02_research/           # Technical research
│       ├── 002_DESIGN/                # System design
│       │   ├── 01_api_design/         # API specifications
│       │   └── 02_data_model/         # Database schema
│       ├── 003_IMPLEMENT/             # Build phase
│       │   ├── 01_backend/            # Backend tasks
│       │   ├── 02_frontend/           # Frontend tasks
│       │   └── 03_testing/            # Test tasks
│       └── 004_VALIDATE/              # Validation phase
├── completed/                  # Finished projects
└── .festival/                  # Methodology resources
    ├── agents/                 # AI agent prompts
    ├── templates/              # Task templates
    └── examples/               # Example tasks
```

## What's Included

- **Planning Agents** - AI prompts for structuring projects
- **Task Templates** - Formats for creating executable tasks
- **Real Examples** - 15+ examples of well-written tasks
- **Methodology Guide** - Complete documentation

## Why Festival Works

1. **Goals drive structure** - Everything traces back to the objective
2. **Tasks are complete** - No ambiguity, full specifications
3. **Context persists** - Information accumulates across sessions
4. **Parallel execution** - Multiple agents work simultaneously
5. **Human oversight** - Review and adjust at natural checkpoints

## Support & Documentation

- **Complete Guide**: `festivals/.festival/FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md`
- **Templates**: `festivals/.festival/templates/`
- **Examples**: `festivals/.festival/examples/`

## Community

- [Issues](../../issues) - Report problems or suggestions
- [Discussions](../../discussions) - Share experiences
- [Contributing](CONTRIBUTING.md) - Help improve the methodology

## License

MIT - Use it, adapt it, make it yours.

---

**The Bottom Line**: Festival Methodology helps you collaboratively create actionable tasks from goals, enabling AI agents to work autonomously for extended periods and deliver complete, working systems.

