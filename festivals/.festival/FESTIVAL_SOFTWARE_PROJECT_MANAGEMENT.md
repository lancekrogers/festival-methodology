---
id: FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT
aliases: []
tags: []
---

# Festival Planning Methodology

## Overview

Festival methodology is a goal-oriented, flexible planning approach that
prioritizes concrete objectives over process overhead. Unlike traditional agile
methodologies, festivals focus on achieving specific goals rather than adhering
to rigid ceremonies and timeboxes.

## Core Principles

1. **Goal-First Planning**: Every festival starts with a clear, high-level
   objective
2. **Concrete Requirements**: Goals are broken down into tangible, achievable
   requirements
3. **Flexible Timeframes**: No artificial sprints or deadlines - festivals
   complete when the goal is achieved
4. **Single Initial Planner**: One person with deep understanding plans the
   festival initially, acting as a benevolent dictator to ensure coherent vision
   and avoid design-by-committee
5. **Adaptive Execution**: New requirements can be added as discovered during
   development
6. **Sequential Dependencies**: Work is organized in numbered sequences that
   must complete in order
7. **Parallel Work Within Sequences**: Tasks within each sequence can be
   distributed across the team
8. **Mandatory Verification**: Every sequence includes testing, code review, and
   results iteration tasks
9. **Numbered Task Execution**: Tasks and sequences use number prefixes to
   indicate execution order

## Festival Structure

The festival directory structure introduces **phases** as a NEW organizational layer above the existing sequences and tasks structure. Phases group related sequences together logically. The standard four-phase structure is recommended, but phases can be customized, repeated, or reordered based on project needs.

**Three-Level Hierarchy**: Phases → Sequences → Tasks

- **Phases** (NEW): Top-level organization grouping related work (3-digit numbering: 001_, 002_, 003_)
- **Sequences** (EXISTING): Work that must happen in order within a phase (2-digit numbering: 01_, 02_)
- **Tasks** (EXISTING): Individual work items within sequences (2-digit numbering: 01_, 02_)

Here's the recommended structure:

```text
festivals/
├── completed/                  # Optional: Successfully completed festivals
├── canceled/                   # Optional: Abandoned festivals
├── archived/                   # Optional: Deprioritized festivals (backlog)
└── festival_<id>/
    ├── FESTIVAL_OVERVIEW.md    # High-level goal, systems, and features overview
    ├── COMMON_INTERFACES.md    # Master interface definitions (updated throughout phases)
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
    ├── 002_DEFINE_INTERFACES/  # PHASE: Interface Definition and Finalization
    │   ├── docs/                               # Phase-specific documentation
    │   ├── 01_identify_interfaces/             # SEQUENCE: Interface discovery
    │   │   ├── 01_api_inventory.md             # TASK
    │   │   ├── 01_data_flow_analysis.md        # TASK (parallel with above)
    │   │   ├── 02_interface_mapping.md         # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 02_define_contracts/                # SEQUENCE: Contract definition
    │   │   ├── 01_api_specifications.md        # TASK
    │   │   ├── 01_data_schemas.md              # TASK (parallel with above)
    │   │   ├── 01_component_interfaces.md      # TASK (parallel with above)
    │   │   ├── 02_validation_rules.md          # TASK (after 01_ tasks)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   ├── 03_finalize_interfaces/             # SEQUENCE: Interface finalization
    │   │   ├── 01_stakeholder_review.md        # TASK
    │   │   ├── 02_interface_approval.md        # TASK (after 01_ task)
    │   │   ├── 03_testing_and_verify.md        # TASK
    │   │   ├── 04_code_review.md               # TASK
    │   │   ├── 05_review_results_update_tasks_iterate_if_needed.md  # TASK
    │   │   └── results/                        # Sequence results
    │   └── completed/                          # Completed sequences in this phase
    ├── 003_IMPLEMENT/          # PHASE: Implementation
    │   ├── docs/                               # Phase-specific documentation
    │   ├── 01_backend_foundation/              # SEQUENCE: Backend implementation
    │   │   ├── 01_database_setup.md            # TASK (parallel tasks have same number)
    │   │   ├── 01_api_endpoints.md             # TASK (can work simultaneously)
    │   │   ├── 01_auth_middleware.md           # TASK (based on finalized interfaces)
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
    ├── 004_REVIEW_AND_UAT/     # PHASE: User Review and UAT
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
    └── archived/               # Optional: Deprioritized work (backlog)
```

Note: This structure is a guideline, not a rigid requirement. Adapt it to fit your festival's specific needs.

## Planning Process

### 1. Define the Goal

Start with a clear, concrete objective. This should be outcome-focused, not activity-focused. Goals can range from simple (a single Jira ticket) to complex (a company-wide vision). Anyone with a goal and understanding of what needs to be done can be a festival planner.

### 2. Break Down into Systems and Features

- **Systems**: Major components or areas of work
- **Features**: Specific functionality within systems

### 3. Define Common Interfaces

**This is the most critical step for enabling parallel development and reducing iterations.**

Before any implementation begins:

1. **Identify All System Interfaces**: APIs, database schemas, service contracts, component props, function signatures, event structures, etc.
2. **Document Interface Specifications**: Create detailed interface definitions with:
   - Input/output parameters and types
   - Expected behavior and constraints
   - Error handling specifications
   - Example usage scenarios
3. **Review and Iterate**: Have stakeholders and technical leads review all interfaces
4. **Finalize Interfaces**: Mark interfaces as finalized before beginning implementation sequences

**Benefits of Interface-First Approach:**

- Enables true parallel development across teams and agents
- Reduces integration issues and rework
- Provides clear contracts for all system boundaries
- Allows accurate festival planning with minimal iterations

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

**Phases are a NEW organizational layer above the existing sequences and tasks structure.** They group related sequences together logically. The standard four-phase structure is recommended, but phases can be customized, repeated, or reordered based on project needs.

**Understanding the Three-Level Hierarchy:**

- **Phases** (NEW CONCEPT): High-level organization grouping related sequences (use 3-digit numbering: 001_, 002_, 003_)
- **Sequences** (EXISTING CONCEPT): Work that must happen in order within a phase (use 2-digit numbering: 01_, 02_)
- **Tasks** (EXISTING CONCEPT): Individual work items within sequences (use 2-digit numbering: 01_, 02_)

#### Standard Four-Phase Structure

#### Phase 001: Plan (001_PLAN/)

**Purpose**: Requirements gathering, analysis, and architectural planning
**Contains sequences for**: Requirements gathering, stakeholder interviews, user research, system architecture design, technology selection, feasibility studies, risk assessment

#### Phase 002: Define Interfaces (002_DEFINE_INTERFACES/)

**Purpose**: Interface definition and finalization - **most critical for parallel development**
**Contains sequences for**: Interface identification, API contracts definition, data schema creation, component interface specification, stakeholder review, interface finalization

#### Phase 003: Implement (003_IMPLEMENT/)

**Purpose**: All implementation work based on finalized interfaces
**Contains sequences for**: Backend development, frontend implementation, automated testing, integration testing, code review, quality assurance

#### Phase 004: Review & UAT (004_REVIEW_AND_UAT/)

**Purpose**: User-focused validation and business acceptance - **human validation separate from automated testing**
**Contains sequences for**: User acceptance testing, stakeholder review, business sign-off, documentation validation, training material validation, deployment readiness

#### Flexible Phase Organization

Phases can be customized for different project types:

**Iterative Development**: `001_PLAN → 002_DEFINE_INTERFACES → 003_IMPLEMENT → 004_REVIEW_AND_UAT → 005_PLAN → 006_IMPLEMENT → 007_REVIEW_AND_UAT`

**Implementation Focus**: `001_PLAN → 002_DEFINE_INTERFACES → 003_IMPLEMENT → 004_REVIEW_AND_UAT → 005_IMPLEMENT → 006_REVIEW_AND_UAT`

**Research Heavy**: `001_PLAN → 002_RESEARCH → 003_DEFINE_INTERFACES → 004_PROTOTYPE → 005_REVIEW_AND_UAT → 006_IMPLEMENT`

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

## Three-Phase Benefits

The three-phase structure provides significant advantages:

**Clear Separation of Concerns**: Each phase has a distinct purpose and deliverable

- Phase 1 (Plan): Complete understanding of requirements and architecture
- Phase 2 (Define Interfaces): Finalized contracts enabling parallel development
- Phase 3 (Implement): Systematic implementation based on solid foundations

**Reduced Risk**: Issues are caught early in planning and interface definition phases
**Parallel Development**: Once interfaces are locked, teams can work simultaneously
**Better Communication**: Clear phase boundaries help stakeholders understand progress
**Quality Built-In**: Each phase includes verification and iteration cycles

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
- **archived/**: Like a backlog - deprioritized work that may be valuable later but isn't needed for the current goal

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
    │   ├── 03_component_interfaces/      # SEQUENCE: Component interface definition
    │   │   ├── 01_registration_components.md     # TASK
    │   │   ├── 01_verification_ui_props.md       # TASK (parallel)
    │   │   ├── 02_state_management_contracts.md  # TASK (after 01_ tasks)
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
    └── archived/                         # Deprioritized work
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
   work to canceled/, and deprioritized work to archived/
8. **Follow Festival Rules**: Reference and adhere to FESTIVAL_RULES.md
   throughout execution
9. **Test Everything**: Include specific testing commands and expected outputs
