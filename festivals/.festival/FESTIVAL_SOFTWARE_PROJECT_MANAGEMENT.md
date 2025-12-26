---
id: FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT
aliases: []
tags: []
---

# Festival Planning Methodology

## Overview

Festival methodology is a **collaborative, step-oriented planning approach** between humans and AI agents that enables goal achievement through systematic progression. Unlike traditional project management, festivals focus on identifying and completing the logical steps needed to achieve goals, leveraging unprecedented AI-human efficiency that makes traditional time estimates obsolete.

## Core Principles

1. **Goal-Oriented Step Planning**: Think in terms of steps needed to achieve goals, not time estimates or schedules
2. **Human-AI Collaborative Planning**: Humans provide goals and requirements, AI agents identify and structure the logical steps needed
3. **Requirements-Driven Implementation**: Implementation sequences can ONLY be created after requirements are defined - either through planning phases or external documentation
4. **Just-in-Time Sequence Creation**: Implementation work is added one step at a time as requirements become clear, based on logical progression
5. **Hyper-Efficient AI Execution**: AI-human collaboration works at unprecedented speeds, making traditional time estimates meaningless
6. **Step-Based Progression**: Work is organized as logical steps (phases → sequences → tasks) that build toward goal achievement
7. **Context Preservation**: All decisions and rationale captured in CONTEXT.md to maintain continuity across sessions
8. **Quality Gates**: Every implementation sequence includes verification steps to ensure goal progression
9. **Extensible Methodology**: Extensions available for specialized needs like multi-system coordination

## Step-Based vs Time-Based Thinking

**FUNDAMENTAL PRINCIPLE**: Festival Methodology thinks in **STEPS TO GOALS**, not time estimates.

### Why Steps, Not Time?

Traditional project management focuses on:

- Time estimates and schedules
- Duration-based planning
- Timeline management
- Resource allocation over time

Festival Methodology focuses on:

- **Logical steps toward goal achievement**
- **Completion criteria for each step**
- **Dependencies between steps**
- **Parallel step opportunities**

### The Efficiency Reality

AI-human collaboration operates at unprecedented efficiency levels that make traditional time estimates obsolete. Instead of asking "How long will this take?", Festival Methodology asks:

- "What steps are needed to achieve this goal?"
- "What's the logical order for these steps?"
- "What can be done in parallel?"
- "How do we know each step is complete?"
- "What's the next step after this one completes?"

## Collaborative Workflow

**CRITICAL UNDERSTANDING**: Festival Methodology is NOT about AI agents pre-planning entire projects. It's about **human-AI collaboration** where:

### Human Responsibilities

- Provide project goals and requirements
- Define success criteria and constraints
- Make architectural and design decisions
- Review and approve AI-generated sequences
- Guide iteration and adaptation

### AI Agent Responsibilities  

- Identify logical steps needed to achieve goals
- Structure requirements into executable step sequences
- Create detailed task specifications with completion criteria
- Execute implementation steps autonomously at unprecedented speed
- Document decisions and progress toward goal achievement
- Request clarification when requirements are unclear

### The Planning-Implementation Boundary

**Planning Steps (Optional):**

- May be completed before festival creation
- May be first step in festival progression
- May be provided as external documentation
- Results in clear requirements for implementation steps

**Implementation Steps:**

- Can ONLY be created after requirements are defined
- Are added one logical step at a time based on requirements
- Follow goal progression logic, not time schedules
- Emerge from human-provided specifications and goal definitions

## Festival Structure and Phase Flexibility

### Phase Types and Structure

**Planning/Research Phases (Unstructured):**

- Used for requirements gathering, research, and documentation
- Often just contain documents, findings, and specifications
- No need for sequences and tasks unless deep planning requires it
- Examples: 001_RESEARCH, 001_PLAN, 001_REQUIREMENTS

**Implementation Phases (Structured):**

- MUST have sequences and tasks for AI agent execution
- This is where agents work autonomously for long periods
- Add as many implementation phases as needed
- Examples: 002_IMPLEMENT_CORE, 003_IMPLEMENT_FEATURES, 004_IMPLEMENT_UI

**Key Principle**: Don't pre-plan phases. Add them as needed when requirements emerge or new implementation work is identified.

**Three-Level Hierarchy**: Phases → Sequences → Tasks

- **Phases** (NEW): Top-level organization grouping related work (3-digit numbering: 001_, 002_, 003_)
- **Sequences** (EXISTING): Work that must happen in order within a phase (2-digit numbering: 01_, 02_)
- **Tasks** (EXISTING): Individual work items within sequences (2-digit numbering: 01_, 02_)

### Sequence Creation Guidelines

**WHEN TO CREATE SEQUENCES:**

✅ **Create sequences when:**

- Human provides specific requirements or specifications
- Planning phase has been completed with clear deliverables
- External planning documents define what needs to be built
- Human explicitly asks for implementation of specific functionality

❌ **DO NOT create sequences when:**

- No requirements have been provided
- Planning phase hasn't been completed
- Guessing what might need to be implemented
- Making assumptions about user needs

### Sequence Design Guidelines

**Good Sequences** contain 3-6 related tasks that:

- Build on each other logically
- Share common setup or dependencies  
- Form a cohesive unit of work (e.g., "user authentication", "API endpoints")
- Can be assigned to one person/agent for focused work
- Are derived from specific requirements or specifications

**Avoid These Sequence Anti-Patterns:**

- Single task per sequence (make it a standalone task instead)
- Unrelated tasks grouped arbitrarily
- Sequences with >8 tasks (break into multiple sequences)
- Mixing different work types (frontend + backend + DevOps in same sequence)
- **Creating sequences without requirements** (the biggest anti-pattern)

**Example Good Sequence:**

```
01_user_authentication/
├── 01_create_user_model.md
├── 02_add_password_hashing.md
├── 03_implement_login_endpoint.md
├── 04_add_jwt_tokens.md
├── 05_testing_and_verify.md       ← Standard quality gate
├── 06_code_review.md              ← Standard quality gate
└── 07_review_results_iterate.md   ← Standard quality gate
```

Here's the recommended structure:

```text
festivals/
├── completed/                  # Optional: Successfully completed festivals
├── canceled/                   # Optional: Abandoned festivals
├── dungeon/                    # Optional: Archived/deprioritized festivals (backlog)
└── festival_<id>/
    ├── FESTIVAL_OVERVIEW.md    # High-level goal, systems, and features overview
    ├── FESTIVAL_RULES.md       # Rules and principles to follow throughout the festival
    ├── 001_PLAN/               # PHASE: Requirements and Planning
    │   ├── docs/              # Phase-specific documentation
    │   ├── 01_requirements_gathering/    # SEQUENCE: Requirements work
    │   │   ├── 01_stakeholder_interviews.md    # TASK
    │   │   ├── 01_user_research.md             # TASK (parallel with above)
    │   │   ├── 02_requirements_analysis.md     # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 02_architecture_planning/          # SEQUENCE: Architecture work
    │   │   ├── 01_system_design.md             # TASK
    │   │   ├── 01_technology_selection.md      # TASK (parallel with above)
    │   │   ├── 02_feasibility_study.md         # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   └── completed/                          # Completed sequences in this phase
    ├── 002_IMPLEMENT/          # PHASE: Implementation
    │   ├── docs/                               # Phase-specific documentation
    │   ├── 01_backend_foundation/              # SEQUENCE: Backend implementation
    │   │   ├── 01_database_setup.md            # TASK (parallel tasks have same number)
    │   │   ├── 01_api_endpoints.md             # TASK (can work simultaneously)
    │   │   ├── 01_auth_middleware.md           # TASK
    │   │   ├── 02_integration_layer.md         # TASK (must complete after 01_* tasks)
    │   │   ├── 03_automated_testing.md         # TASK (automated testing and verification)
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 02_frontend_implementation/         # SEQUENCE: Frontend implementation
    │   │   ├── 01_component_library.md         # TASK
    │   │   ├── 01_user_interface.md            # TASK (parallel with above)
    │   │   ├── 02_state_management.md          # TASK (after 01_ tasks)
    │   │   ├── 03_automated_testing.md         # TASK (automated testing)
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 03_integration_testing/             # SEQUENCE: Integration testing
    │   │   ├── 01_end_to_end_tests.md          # TASK (automated E2E testing)
    │   │   ├── 02_performance_testing.md       # TASK (after 01_ task)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   └── completed/                          # Completed sequences in this phase
    ├── 003_REVIEW_AND_UAT/     # PHASE: User Review and UAT
    │   ├── docs/                               # Phase-specific documentation
    │   ├── 01_user_acceptance_testing/         # SEQUENCE: User acceptance testing
    │   │   ├── 01_uat_planning.md              # TASK
    │   │   ├── 01_test_scenarios.md            # TASK (parallel with above)
    │   │   ├── 02_user_testing_execution.md    # TASK (after 01_ tasks)
    │   │   ├── 03_feedback_collection.md       # TASK (after 02_ task)
    │   │   ├── 04_testing_and_verify.md        # TASK
    │   │   ├── 05_code_review.md               # TASK
    │   │   ├── 06_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 02_stakeholder_review/              # SEQUENCE: Stakeholder validation
    │   │   ├── 01_business_requirements_validation.md  # TASK
    │   │   ├── 01_stakeholder_demos.md         # TASK (parallel with above)
    │   │   ├── 02_sign_off_process.md          # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 03_deployment_readiness/            # SEQUENCE: Deployment preparation
    │   │   ├── 01_deployment_validation.md     # TASK
    │   │   ├── 01_documentation_review.md      # TASK (parallel with above)
    │   │   ├── 02_training_material_validation.md  # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   └── completed/                          # Completed sequences in this phase
    ├── completed/              # Optional: Completed phases moved here
    ├── canceled/               # Optional: Abandoned phases or sequences
    └── dungeon/                # Optional: Archived/deprioritized work (backlog)
```

Note: This structure is a guideline, not a rigid requirement. Adapt it to fit your festival's specific needs.

## Planning Process

### 1. Define the Goal

Start with a clear, concrete objective. This should be outcome-focused, not activity-focused. Goals can range from simple (a single Jira ticket) to complex (a company-wide vision). Anyone with a goal and understanding of what needs to be done can be a festival planner.

### 2. Break Down into Systems and Features

- **Systems**: Major components or areas of work
- **Features**: Specific functionality within systems

### 3. Plan Implementation Approach

**Define how implementation work will be organized and executed.**

For most projects:

1. **Define Implementation Approach**: Identify the components and systems that need to be built
2. **Plan Implementation Structure**: Organize implementation work into logical sequences:
   - System components and their responsibilities
   - Expected behavior and constraints
   - Error handling approaches
   - Integration patterns
3. **Review and Iterate**: Have stakeholders and technical leads review the implementation plan

**For Multi-System Projects**: Consider the [Interface Planning Extension](extensions/interface-planning/) to add formal interface definition phases when system coordination is critical.

**Benefits of Clear Implementation Planning:**

- Enables organized development across teams and agents
- Reduces confusion and rework through clear structure
- Provides clear understanding of system components
- Allows systematic festival execution with minimal iterations

### 4. Create FESTIVAL_OVERVIEW.md

Document:

- The high-level goal
- Systems breakdown (if applicable)
- Features breakdown (if applicable)
- Success criteria

### 5. Define FESTIVAL_RULES.md

Establish the principles and quality standards that all workers must follow throughout the festival. This ensures consistent quality and reminds workers of best practices at each step. Rules should cover:

- Engineering excellence principles
- Quality standards and gates
- Development process requirements
- Decision-making guidelines

### 6. Organize Work into Flexible Phases

**Phases are a NEW organizational layer above the existing sequences and tasks structure.** They group related sequences together logically. The 3-phase structure handles most development scenarios, but phases can be customized, repeated, or reordered based on project needs.

**Understanding the Three-Level Hierarchy:**

- **Phases** (NEW CONCEPT): High-level organization grouping related sequences (use 3-digit numbering: 001_, 002_, 003_)
- **Sequences** (EXISTING CONCEPT): Work that must happen in order within a phase (use 2-digit numbering: 01_, 02_)
- **Tasks** (EXISTING CONCEPT): Individual work items within sequences (use 2-digit numbering: 01_, 02_)

#### Common Phase Patterns (Not Rigid)

#### Planning Phases (When Needed)

**Examples**: 001_PLAN, 001_RESEARCH, 001_REQUIREMENTS
**Structure**: Often just documents and findings - no sequences/tasks required unless deep planning
**Purpose**: Gather requirements, research unknowns, document decisions

#### Implementation Phases (Core Work)

**Examples**: 002_IMPLEMENT_CORE, 003_IMPLEMENT_FEATURES, 004_IMPLEMENT_UI
**Structure**: MUST have sequences and tasks for AI execution
**Purpose**: Where agents work autonomously on structured implementation tasks
**Key**: Add as many implementation phases as your goal requires

#### Validation Phases

**Examples**: 005_REVIEW_AND_UAT, 006_VALIDATE, 007_ACCEPTANCE
**Purpose**: Human validation, user acceptance, completion verification

#### Extensions for Specialized Needs

For projects requiring system coordination, use the [Interface Planning Extension](extensions/interface-planning/) which adds interface definition phases. See the [Extensions Guide](extensions/) for other specialized workflow patterns.

#### Example Phase Progressions

**Simple Project**: `001_IMPLEMENT → 002_REVIEW`
(Requirements already provided)

**Standard Pattern**: `001_PLAN → 002_IMPLEMENT → 003_REVIEW_AND_UAT`
(Basic requirements gathering and implementation)

**Multiple Implementations**: `001_PLAN → 002_IMPLEMENT_CORE → 003_IMPLEMENT_FEATURES → 004_IMPLEMENT_UI → 005_REVIEW`
(Complex project with staged implementation)

**Research First**: `001_RESEARCH → 002_PROTOTYPE → 003_IMPLEMENT → 004_VALIDATE`
(Phases 001-002 are unstructured exploration, 003 is structured implementation)

**Multi-System Projects**: Use the [Interface Planning Extension](extensions/interface-planning/) when coordination is critical.

**Custom Phases**: Add specialized phases like `005_SECURITY_AUDIT/`, `006_PERFORMANCE_OPTIMIZATION/`, or `007_MIGRATION/` as needed

**Within Each Phase:**

- **Phases** use 3-digit numbering (001_, 002_, 003_) to support hundreds of phases
- **Sequences** within phases use 2-digit numbering (01_, 02_, etc.) for proper ordering
- **Tasks** within sequences use 2-digit numbering (01_task.md, 02_task.md, etc.)
- Tasks with the same number can be executed in parallel (e.g., 01_task_a.md, 01_task_b.md, 01_task_c.md)
- Each task gets its own markdown file with clear requirements
- Every sequence must include these verification tasks:
  - `XX_testing_and_verify.md` - Testing and verification (where XX follows implementation tasks)
  - `XX+1_code_review.md` - Code review
  - `XX+2_review_results_update_tasks_iterate_if_needed.md` - Review results and iterate if needed
- Create a `results/` subdirectory in each sequence for testing results and code review documents
- Include a `docs/` directory in each phase for phase-specific documentation

**Numbering System Benefits:**

- **3-digit phases**: Supports up to 999 phases for large, long-running festivals
- **2-digit sequences/tasks**: Maintains readability while supporting up to 99 items per level
- **Proper sorting**: Ensures correct alphabetical and numerical ordering in directory trees
- **Visual consistency**: Clear hierarchy distinction between organizational levels

## Phase Flexibility Benefits

The flexible phase approach provides significant advantages:

**Adapt to Your Needs**: Phases match your actual work, not a rigid template

- Planning phases: Unstructured when simple requirements gathering
- Implementation phases: Structured for AI agent execution
- Multiple implementation phases: Add as many as needed for complex projects
- Skip unnecessary phases: No planning phase if requirements provided

**Reduced Overhead**: Only add structure where it provides value
**Clear Purpose**: Each phase type has distinct characteristics
**Better AI Execution**: Implementation phases optimized for autonomous work
**Natural Progression**: Phases emerge as requirements become clear

## Flexibility and Scaling

Festivals scale from simple to complex:

- **Simple Festival**: Just FESTIVAL_OVERVIEW.md, COMMON_INTERFACES.md, and implementation tasks in 3_Implement/
- **Medium Festival**: All three phases with multiple sequences in each phase
- **Complex Festival**: Full systems/ and features/ directories with extensive documentation

The optional directories serve specific purposes:

- **specs/**: Store requirements documents, analysis notes, research findings, and planning artifacts
- **docs/**: House documentation directly related to the festival's goal
- **completed/**: Move successfully finished festivals or sequences here to keep the active workspace clean
- **canceled/**: Store abandoned festivals or sequences that were planned but won't be executed
- **dungeon/**: Archived work - like a backlog for deprioritized festivals that may be valuable later but aren't needed for the current goal

These directories can be included at any complexity level as needed and are only created when necessary.

## Key Advantages

1. **No Process Overhead**: No daily standups, sprint planning, or retrospectives unless actually needed
2. **Clear Dependencies**: Sequential directories make dependencies obvious
3. **Flexible Scope**: Add requirements as you discover them
4. **Goal Achievement**: Success is measured by goal completion, not velocity metrics
5. **Deep Understanding**: Requires and encourages thorough problem understanding upfront

## When to Use Festival Methodology

Festival methodology works best when:

- You have a clear goal to achieve (from a simple bug fix to a major product launch)
- You want to minimize process overhead
- The work has natural dependencies and multiple system interfaces
- You need flexibility in scope and timeline
- Team members can work independently on well-defined tasks with clear interface contracts
- You're working solo or collaboratively - festivals adapt to both modes
- You can define system interfaces upfront to enable parallel development

## Example Festival

```text
festivals/
└── festival_user_onboarding/
    ├── FESTIVAL_OVERVIEW.md
    ├── COMMON_INTERFACES.md
    ├── FESTIVAL_RULES.md
    ├── 001_PLAN/                         # PHASE: Requirements and Planning
    │   ├── docs/                         # Phase-specific documentation
    │   ├── 01_requirements_analysis/     # SEQUENCE: Requirements work
    │   │   ├── 01_user_journey_mapping.md        # TASK
    │   │   ├── 01_stakeholder_interviews.md      # TASK (parallel)
    │   │   ├── 02_requirements_specification.md  # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 02_system_design/             # SEQUENCE: System design work
    │   │   ├── 01_architecture_planning.md       # TASK
    │   │   ├── 01_technology_evaluation.md       # TASK (parallel)
    │   │   ├── 02_security_considerations.md     # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   └── completed/                    # Completed sequences in this phase
    ├── 002_DEFINE_INTERFACES/            # PHASE: Interface Definition
    │   ├── docs/                         # Phase-specific documentation
    │   ├── 01_api_contracts/             # SEQUENCE: API contract definition
    │   │   ├── 01_user_registration_api.md       # TASK
    │   │   ├── 01_email_verification_api.md      # TASK (parallel)
    │   │   ├── 01_kyc_integration_api.md         # TASK (parallel)
    │   │   ├── 02_error_response_standards.md    # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 02_data_schemas/              # SEQUENCE: Data schema definition
    │   │   ├── 01_user_model_schema.md           # TASK
    │   │   ├── 01_verification_schema.md         # TASK (parallel)
    │   │   ├── 02_database_relationships.md      # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 03_frontend_implementation/   # SEQUENCE: Frontend implementation
    │   │   ├── 01_registration_components.md     # TASK
    │   │   ├── 01_verification_ui_components.md  # TASK (parallel)
    │   │   ├── 02_state_management_setup.md      # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   └── completed/                    # Completed sequences in this phase
    ├── 003_IMPLEMENT/                    # PHASE: Implementation
    │   ├── docs/                         # Phase-specific documentation
    │   ├── 01_backend_foundation/        # SEQUENCE: Backend implementation
    │   │   ├── 01_user_model_updates.md          # TASK (parallel)
    │   │   ├── 01_api_endpoints.md               # TASK (based on interfaces)
    │   │   ├── 01_database_migrations.md         # TASK (parallel)
    │   │   ├── 02_automated_testing.md           # TASK (after 01_ tasks)
    │   │   ├── 03_code_review.md                 # TASK
    │   │   ├── 04_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 02_frontend_implementation/   # SEQUENCE: Frontend implementation
    │   │   ├── 01_registration_flow.md           # TASK (parallel)
    │   │   ├── 01_verification_ui.md             # TASK (based on interfaces)
    │   │   ├── 02_error_handling.md              # TASK (after 01_ tasks)
    │   │   ├── 03_automated_testing.md           # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 03_integration_testing/       # SEQUENCE: Integration testing
    │   │   ├── 01_e2e_tests.md                   # TASK (automated E2E)
    │   │   ├── 01_performance_validation.md      # TASK (parallel)
    │   │   ├── 02_testing_and_verify.md          # TASK (after 01_ tasks)
    │   │   ├── 03_code_review.md                 # TASK
    │   │   ├── 04_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   └── completed/                    # Completed sequences in this phase
    ├── 004_REVIEW_AND_UAT/               # PHASE: User Review and UAT
    │   ├── docs/                         # Phase-specific documentation
    │   ├── 01_user_acceptance_testing/   # SEQUENCE: User acceptance testing
    │   │   ├── 01_uat_planning.md                # TASK
    │   │   ├── 01_test_scenarios.md              # TASK (parallel)
    │   │   ├── 02_user_testing_execution.md      # TASK (after 01_ tasks)
    │   │   ├── 03_feedback_analysis.md           # TASK (after 02_ task)
    │   │   ├── 04_testing_and_verify.md          # TASK
    │   │   ├── 05_code_review.md                 # TASK
    │   │   ├── 06_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   ├── 02_stakeholder_review/        # SEQUENCE: Stakeholder validation
    │   │   ├── 01_business_validation.md         # TASK
    │   │   ├── 01_stakeholder_demos.md           # TASK (parallel)
    │   │   ├── 02_requirements_sign_off.md       # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md          # TASK
    │   │   ├── 04_code_review.md                 # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                          # Sequence results
    │   └── completed/                    # Completed sequences in this phase
    ├── completed/                        # Completed phases
    ├── canceled/                         # Abandoned work
    └── dungeon/                          # Archived/deprioritized work
````

## Festival Rules

Festival Rules are a critical component that ensures quality and consistency
throughout the festival. Every festival should include a FESTIVAL_RULES.md file
that workers reference before and during task execution.

### Purpose of Festival Rules

- **Maintain Quality Standards**: Establish clear quality gates and acceptance
  criteria
- **Ensure Consistency**: All workers follow the same principles and practices
- **Reduce Rework**: Prevent common mistakes by providing clear guidelines
  upfront
- **Promote Excellence**: Embed staff-level engineering principles into every
  task

### Common Rules for Software Festivals

1. **Engineering Excellence**
   - Prefer refactoring existing code over rewriting from scratch
   - Follow established patterns and conventions in the codebase
   - Apply SOLID principles and avoid over-engineering (YAGNI)
   - Keep functions under 50 lines, files under 500 lines

2. **Quality Standards**
   - Write tests for all new functionality
   - Maintain or improve code coverage
   - Run linters and type checkers before marking tasks complete
   - Document architectural decisions and breaking changes

3. **Development Process**
   - Create small, focused pull requests (one logical change)
   - Update documentation alongside code changes
   - Consider security implications in all changes
   - Maintain backward compatibility unless explicitly approved

### Task Integration

Each task should include a "Rules Compliance" section that:

- References relevant rules from FESTIVAL_RULES.md
- Includes a pre-task checklist
- Provides a completion checklist for verification

## Creating Actionable Tasks

### The Problem with Abstract Tasks

AI agents often create generic, high-level task descriptions that don't lead to concrete implementation. This defeats the purpose of the festival methodology.

### ❌ Bad Task Examples (Abstract and Vague)

```markdown
# Task: 01_user_management.md
## Objective
Implement user management functionality
## Requirements
- [ ] Create user system
- [ ] Add authentication
- [ ] Handle user data
## Deliverables
- User management feature
- Authentication system
```

**Problems**:

- No specific file names or code examples
- Vague requirements that don't specify implementation details
- Generic deliverables that could mean anything

### ✅ Good Task Examples (Concrete and Specific)

```markdown
# Task: 01_create_user_table_and_model.md
## Objective
Create PostgreSQL user table and Sequelize model with email/password authentication fields

## Requirements
- [ ] Create `users` table with id, email, password_hash, created_at, updated_at
- [ ] Create `models/User.js` with Sequelize model definition
- [ ] Add email validation method with regex: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
- [ ] Add bcrypt password hashing with salt rounds = 12

## Implementation Steps
1. Run: `npx sequelize-cli migration:generate --name create-users-table`
2. Edit migration file with SQL schema
3. Create `models/User.js` with Sequelize model
4. Add bcrypt dependency: `npm install bcrypt`
5. Run: `npx sequelize-cli db:migrate`

## Testing Commands
```bash
npm test -- tests/models/User.test.js
node -e "const User = require('./models/User'); console.log('User model loaded');"
```

## Deliverables

- [ ] `migrations/001_create_users_table.js` migration file
- [ ] `models/User.js` Sequelize model with authentication methods
- [ ] `tests/models/User.test.js` unit tests
- [ ] Updated `package.json` with bcrypt dependency

```

**Why This Works**:
- Specific file names and directory paths
- Exact code snippets and SQL schemas
- Concrete commands to execute
- Testable deliverables with clear file paths

### Guidelines for Writing Actionable Tasks

#### 1. Use Specific Names and Paths
- ✅ `Create models/User.js with Sequelize model`
- ❌ `Create user model`

#### 2. Include Implementation Steps with Code
- ✅ Provide exact SQL, JavaScript, commands
- ❌ Say "implement database schema"

#### 3. Specify Testing Commands
- ✅ `npm test -- tests/models/User.test.js`
- ❌ "Test the functionality"

#### 4. List Exact Deliverables
- ✅ `src/components/LoginForm.jsx`, `tests/LoginForm.test.js`
- ❌ "Login component and tests"

### Task Complexity Levels

#### Level 1: Single File Creation
**Good for**: Creating individual files, small components, utility functions
```markdown
Objective: Create EmailValidator utility with regex validation
Requirements:
- [ ] Create `utils/EmailValidator.js` with isValid() method
- [ ] Use regex pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
- [ ] Export as CommonJS module
Commands: node -e "const validator = require('./utils/EmailValidator'); console.log(validator.isValid('test@example.com'));"
```

#### Level 2: Multi-File Implementation

**Good for**: API endpoints, React components with styling, database operations

```markdown
Objective: Implement user registration API endpoint with validation
Requirements:
- [ ] Create routes/users.js with POST /users endpoint
- [ ] Create middleware/validation.js with registration validation
- [ ] Add bcrypt password hashing
- [ ] Create tests/routes/users.test.js
Commands: curl -X POST localhost:3000/api/users -d '{"email":"test@example.com","password":"SecurePass123"}'
```

#### Level 3: Feature Implementation

**Good for**: Complete features spanning multiple files and systems

```markdown
Objective: Build user authentication flow with database, API, and frontend
Requirements:
- [ ] Database: Create users table with authentication fields
- [ ] Backend: Implement registration/login endpoints with JWT
- [ ] Frontend: Create LoginForm and RegistrationForm components
- [ ] Testing: Unit tests for all components and integration tests
```

### Common Mistakes to Avoid

#### 1. Using Placeholders Instead of Real Examples

- ❌ `interface [ComponentName]Props`
- ✅ `interface LoginFormProps`

#### 2. Abstract Requirements

- ❌ "Handle user authentication"
- ✅ "Implement JWT authentication with 7-day expiry using jsonwebtoken library"

#### 3. Missing Implementation Details

- ❌ "Create database schema"
- ✅ "Create users table with: id SERIAL PRIMARY KEY, email VARCHAR(255) UNIQUE NOT NULL, password_hash VARCHAR(255) NOT NULL"

#### 4. Vague Testing Instructions  

- ❌ "Test the feature"
- ✅ "Run: curl -X POST localhost:3000/api/users -H 'Content-Type: application/json' -d '{\"email\":\"<test@example.com>\"}'"

### Reference Resources

- **TASK_EXAMPLES.md**: 15+ concrete examples across database, API, frontend, DevOps, and testing domains
- **COMMON_INTERFACES_TEMPLATE.md**: Real interface examples instead of placeholders
- **TASK_TEMPLATE.md**: Enhanced template with good vs bad examples

### Implementation-Ready Principle

Every task should be "implementation-ready" - meaning a developer (human or AI) can start coding immediately without needing additional clarification or research.

**Test**: Can someone copy-paste the code examples and commands from your task and get working results?

## Best Practices

1. **Keep Goals Concrete**: Vague goals lead to scope creep
2. **Write Implementation-Ready Tasks**: Include exact code, commands, and file names
3. **Use Real Examples**: Avoid placeholders - use concrete data and realistic scenarios
4. **Sequence Thoughtfully**: Consider dependencies when creating sequences
5. **Stay Flexible**: Add new sequences or tasks as needed
6. **Complete Before Proceeding**: Finish each sequence before starting the next
7. **Organize Finished Work**: Move completed sequences to completed/, canceled
   work to canceled/, and archived/deprioritized work to dungeon/
8. **Follow Festival Rules**: Reference and adhere to FESTIVAL_RULES.md
   throughout execution
9. **Test Everything**: Include specific testing commands and expected outputs
