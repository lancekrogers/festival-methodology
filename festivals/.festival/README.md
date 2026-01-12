# Festival Methodology Resources Guide

This directory contains all the resources needed to implement Festival Methodology in your projects. This guide helps you navigate and use these resources effectively.

## 📚 Step-Based Reading Strategy

**CRITICAL**: To preserve context window and focus on goal progression, follow these rules:

### When to Read What

| Resource              | Read When                                 |
| --------------------- | ----------------------------------------- |
| This README           | Immediately - provides navigation         |
| Core methodology docs | During initial understanding              |
| Templates             | ONLY when creating that specific document |
| Examples              | ONLY when stuck or need clarification     |
| Agents                | ONLY when using that specific agent       |

### Never Do This

❌ Reading all templates upfront "to understand them"
❌ Loading all agents at once

### Always Do This

✅ Read templates one at a time as needed
✅ Read examples only when stuck
✅ Keep templates closed after use
✅ Focus context on actual work, not documentation

## Quick Navigation

- **[Phase Adaptability](#phase-adaptability)** - How to customize phases for your project needs
- **[Templates](#templates)** - Document templates for creating festivals
- **[Agents](#ai-agents)** - Specialized AI agents for festival workflow
- **[Examples](#examples)** - Concrete examples and patterns
- **[Core Documentation](#core-documentation)** - Methodology principles and theory

## Phase Adaptability

**CRITICAL**: Festival phases are guidelines, not rigid requirements. Adapt the structure to match your actual work needs.

## Requirements-Driven Implementation

**MOST CRITICAL**: Implementation sequences can ONLY be created after requirements are defined. This is the core principle of Festival Methodology.

### When Implementation Sequences Can Be Created

✅ **Create implementation sequences when:**

- Human provides specific requirements or specifications
- Planning phase has been completed with deliverables
- External planning documents define what to build
- Human explicitly requests implementation of specific functionality

❌ **NEVER create implementation sequences when:**

- No requirements have been provided
- Planning phase hasn't been completed
- Guessing what might need to be implemented
- Making assumptions about functionality

### The Human-AI Collaboration Model

**Human provides:**

- Project goals and vision
- Requirements and specifications
- Architectural decisions
- Feedback and iteration guidance

**AI agent creates:**

- Structured sequences from requirements
- Detailed task specifications
- Implementation plans
- Progress tracking and documentation

### Standard 3-Phase Pattern (Simple Starting Point)

```
001_PLAN → 002_IMPLEMENT → 003_REVIEW_AND_UAT
```

**Use this when:** Starting with basic requirements gathering and implementation.

### Phase Types

The methodology supports different phase types with appropriate structure:

| Phase Type | Structure | Use Case |
|------------|-----------|----------|
| Planning | Freeform (topics) | Requirements gathering, architecture decisions |
| Research | Freeform (topics) | Investigation, exploration, prototyping |
| Design | Freeform (topics) | UI/UX exploration, system design |
| Implementation | Numbered (sequences/tasks) | Building features, writing code |
| Review | Numbered (sequences/tasks) | Testing, validation, sign-off |

**Freeform Phases (Planning, Research, Design):**
- Use any directory structure that makes sense
- Topic-based directories (requirements/, architecture/, decisions/)
- No numbered sequences or tasks required
- Progress measured by documents and decisions made

**Numbered Phases (Implementation, Review):**
- Use the standard hierarchy: `NN_sequence_name/` with `NN_task_name.md`
- Sequential execution order
- Progress measured by tasks completed

### Phase Flexibility Principles

**Planning Phases (Freeform):**

- Use topic-based directories instead of numbered sequences
- Perfect for backward thinking from goals to requirements
- Contains documents, decisions, and research findings

**Implementation Phases (Structured):**

- MUST have sequences and tasks for AI execution
- Add as many as needed (IMPLEMENT_CORE, IMPLEMENT_FEATURES, IMPLEMENT_UI, etc.)
- This is where agents work autonomously for long periods

### Common Adaptations

**Already Planned Projects:**

```
001_IMPLEMENT → 002_REVIEW_AND_UAT
```

Skip planning if requirements are already provided.

**Multiple Implementation Phases:**

```
001_PLAN → 002_IMPLEMENT_CORE → 003_IMPLEMENT_FEATURES → 004_IMPLEMENT_POLISH → 005_FINAL_REVIEW
```

Add implementation phases as needed for complex builds.

**Research-Heavy Projects:**

```
001_RESEARCH → 002_PROTOTYPE → 003_PLAN → 004_IMPLEMENT → 005_VALIDATE
```

Planning phases (001-003) may just contain documents and findings.
Implementation phase (004) has structured sequences and tasks.

**Bug Fix or Enhancement:**

```
001_ANALYZE → 002_IMPLEMENT → 003_TEST_AND_VERIFY
```

Minimal phases for focused work.

### Phase Design Guidelines

**Good Phase:**

- Represents a distinct step toward goal achievement
- Has clear purpose (requirements gathering OR implementation)
- Planning phases: Can be unstructured documents
- Implementation phases: Must have sequences and tasks
- Added when needed, not pre-planned

**Bad Phase:**

- Created just to follow a pattern
- Planning phase with unnecessary sequences/tasks
- Single sequence worth of work
- Time-based rather than goal-based

### Sequence vs Task Decision

**Create a Sequence When:**

- You have 3+ related tasks
- Tasks must be done in order
- Tasks share common setup/teardown
- Work forms a logical unit

**Make it a Single Task When:**

- Work is atomic and self-contained
- No clear subtasks
- Can be completed in one session
- Doesn't benefit from breakdown

### Standard Quality Gates

**EVERY implementation sequence should end with these tasks:**

```
XX_implementation_tasks.md
XX_testing_and_verify.md
XX_code_review.md
XX_review_results_iterate.md
```

**Example Implementation Sequence:**

```
01_backend_api/
├── 01_create_user_endpoints.md
├── 02_add_authentication.md
├── 03_implement_validation.md
├── 04_testing_and_verify.md     ← Standard quality gate
├── 05_code_review.md            ← Standard quality gate
└── 06_review_results_iterate.md ← Standard quality gate
```

These quality gates ensure:

- Functionality works as specified
- Code meets project standards
- Issues are identified and resolved
- Knowledge is transferred through review

## Goal Files - New

Goal files provide clear objectives and evaluation criteria at every level of the festival hierarchy. They ensure each phase and sequence has a specific goal to work towards and can be evaluated upon completion.

### Goal File Hierarchy

```
festival/
├── FESTIVAL_GOAL.md          # Overall festival goals and success criteria
├── 001_PLAN/
│   ├── PHASE_GOAL.md         # Phase-specific goals
│   ├── 01_requirements/
│   │   └── SEQUENCE_GOAL.md  # Sequence-specific goals
│   └── 02_architecture/
│       └── SEQUENCE_GOAL.md
└── [continues for all phases and sequences]
```

### Goal Templates

1. **FESTIVAL_GOAL_TEMPLATE.md**
   - Comprehensive festival-level goals
   - Success criteria across all dimensions
   - KPIs and stakeholder metrics
   - Post-festival evaluation framework

2. **PHASE_GOAL_TEMPLATE.md**
   - Phase-specific objectives
   - Contribution to festival goal
   - Phase evaluation criteria
   - Lessons learned capture

3. **SEQUENCE_GOAL_TEMPLATE.md**
   - Sequence-level objectives
   - Task alignment verification
   - Progress tracking metrics
   - Post-completion assessment

### Using Goal Files

**When Planning Goal Progression:**

1. Create FESTIVAL_GOAL.md from template (overall goal achievement criteria)
2. Create PHASE_GOAL.md for each phase (step toward festival goal)
3. Create SEQUENCE_GOAL.md for each sequence (step toward phase goal)
4. Ensure alignment: Sequence goals → Phase goals → Festival goal achievement

**During Execution:**

- Track progress against goal metrics
- Update completion status
- Identify risks to goal achievement

**At Completion:**

- Evaluate goal achievement
- Document lessons learned
- Capture recommendations
- Get stakeholder sign-off

## Templates

Templates provide standardized structures for festival documentation. Each template includes inline examples and clear instructions.

### Essential Templates

1. **FESTIVAL_OVERVIEW_TEMPLATE.md**
   - Define project goals and success criteria
   - Create stakeholder matrix
   - Document problem statement
   - _Use this first when starting a new festival_

2. **Interface Planning Extension** (When Needed)
   - For multi-system projects requiring coordination
   - Define interfaces before implementation
   - Enables parallel development
   - _See extensions/interface-planning/ for templates_

3. **FESTIVAL_RULES_TEMPLATE.md**
   - Project-specific standards and guidelines
   - Quality gates and compliance requirements
   - Team agreements and conventions
   - _Customize for your project's needs_

4. **TASK_TEMPLATE.md**
   - Comprehensive task structure (full version)
   - Detailed implementation steps
   - Testing and verification sections
   - _Use for complex or critical tasks_

5. **FESTIVAL_TODO_TEMPLATE.md** (Markdown)
   - Human-readable progress tracking
   - Checkbox-based task management
   - Visual project status
   - _Use for manual tracking and documentation_

6. **FESTIVAL_TODO_TEMPLATE.yaml** (YAML)
   - Machine-readable progress tracking
   - Structured data for automation
   - CI/CD integration ready
   - _Use for automated tooling and reporting_

### When to Use Each Format

**Use Markdown (.md) when:**

- Working directly with AI agents
- Manual progress tracking
- Creating documentation
- Sharing with stakeholders

**Use YAML (.yaml) when:**

- Integrating with CI/CD pipelines
- Building automation tools
- Generating reports programmatically
- Need structured data parsing

## AI Agents

Specialized agents help maintain methodology consistency and guide festival execution.

### Available Agents

1. **festival_planning_agent.md**
   - Conducts structured project interviews
   - Creates complete festival structures
   - Ensures proper three-level hierarchy
   - _Trigger: Starting a new project or festival_

2. **festival_review_agent.md**
   - Validates festival structure compliance
   - Reviews quality gates
   - Ensures methodology adherence
   - _Trigger: Before moving phases or major milestones_

3. **festival_methodology_manager.md**
   - Enforces methodology during execution
   - Prevents process drift
   - Provides ongoing governance
   - _Trigger: During active development_

### Using Agents with Claude Code

```
You: Please use the festival planning agent to help me create a festival for [project description]

Claude: [Loads festival_planning_agent.md and conducts structured interview]
```

### Agent Collaboration Pattern

```
Planning → Review → Execution Management
    ↑         ↓           ↓
    └─────────────────────┘
         (Iterate)
```

## fest CLI Tool

Use the `fest` CLI for efficient festival management. It saves tokens and ensures correct structure.

```bash
# Learn the methodology
fest understand

# Create structure
fest create festival --name "my-project" --json
fest create phase --name "IMPLEMENT" --json

# Add quality gates
fest task defaults sync --approve --json
```

Run `fest understand` for methodology guidance, or `fest --help` for command details.

## Examples

Learn from concrete implementations and proven patterns.

### Available Examples

1. **TASK_EXAMPLES.md**
   - 15+ real task examples
   - Covers different domains (database, API, frontend, DevOps)
   - Shows good vs bad patterns
   - Reference for writing effective tasks

2. **FESTIVAL_TODO_EXAMPLE.md**
   - Complete festival tracking example
   - Shows all states and transitions
   - Demonstrates progress reporting
   - Template for your TODO.md files

### Common Patterns

**Pattern 1: Interface-First Development**

```
Phase 001: Define requirements
Phase 002: Define ALL interfaces ← Critical
Phase 003: Parallel implementation
Phase 004: Integration and testing
```

**Pattern 2: Quality Gates**

```
Every sequence ends with:
- XX_testing_and_verify
- XX_code_review
- XX_review_results_iterate
```

**Pattern 3: Parallel Task Execution**

```
Tasks with same number can run in parallel:
- 01_frontend_setup.md
- 01_backend_setup.md
- 01_database_setup.md
```

## Core Documentation

Understanding the methodology principles and theory.

### Essential Reading Order

1. **FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md**
   - Core methodology principles
   - Three-level hierarchy explanation
   - Interface-first development rationale
   - _Read this first to understand the "why"_

2. **PROJECT_MANAGEMENT_SYSTEM.md**
   - Markdown/YAML tracking system
   - Progress calculation methods
   - Automation opportunities
   - _Read this to understand tracking mechanics_

## Creating Your First Festival

### Quick Start Process

1. **Understand the Goal**

   ```
   Read: FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md (core principles)
   Step: Learn step-based goal achievement approach
   ```

2. **Plan the Festival**

   ```
   Use: festival_planning_agent.md
   Creates: Initial festival structure based on goal requirements
   Step: Define logical progression toward goal
   ```

3. **Create Core Documents**

   ```
   Templates needed:
   - FESTIVAL_OVERVIEW_TEMPLATE.md → FESTIVAL_OVERVIEW.md
   - FESTIVAL_RULES_TEMPLATE.md → FESTIVAL_RULES.md
   - (Optional) Interface templates if multi-system project
   ```

4. **Structure Phases**

   ```
   Common phases (add as needed):
   - 001_PLAN (if requirements gathering needed)
   - 002_IMPLEMENT (structured sequences/tasks)
   - 003_IMPLEMENT_FEATURES (additional implementation)
   - 004_REVIEW_AND_UAT (validation)
   ```

5. **Create Tasks**

   ```
   Use: TASK_TEMPLATE.md
   Reference: TASK_EXAMPLES.md
   ```

6. **Track Progress**

   ```
   Create: TODO.md from FESTIVAL_TODO_TEMPLATE.md
   Update: As tasks complete
   ```

## Template Customization Guide

### Adapting Templates to Your Needs

All templates are starting points. Customize them by:

1. **Removing irrelevant sections**
   - Not every project needs every section
   - Keep what adds value

2. **Adding project-specific sections**
   - Add sections for your domain
   - Include compliance requirements
   - Add team-specific needs

3. **Adjusting complexity**
   - Simple projects: Use minimal sections
   - Complex projects: Use comprehensive templates
   - Critical tasks: Include all verification steps

### Template Metadata (Frontmatter)

Templates include YAML frontmatter for tooling:

```yaml
---
id: TEMPLATE_NAME
aliases: [alternative, names]
tags: []
created: "YYYY-MM-DD"
modified: "YYYY-MM-DD"
---
```

This metadata:

- Enables search and indexing
- Supports knowledge management tools
- Provides version tracking
- Can be safely ignored if not needed

## Best Practices

### Do's

- ✅ Start with planning agent for new festivals
- ✅ Define interfaces before implementation
- ✅ Include quality gates in every sequence
- ✅ Update TODO.md as you progress
- ✅ Customize templates to fit your project

### Don'ts

- ❌ Skip Phase 002 (Interface Definition)
- ❌ Start coding before planning is complete
- ❌ Ignore quality verification tasks
- ❌ Use templates without customization
- ❌ Mix parallel and sequential tasks incorrectly

## Troubleshooting

### Common Issues and Solutions

**Issue: Festival structure too complex**

- Solution: Start with fewer sequences per phase
- Expand as you understand the project better

**Issue: Tasks too abstract**

- Solution: Reference TASK_EXAMPLES.md
- Make tasks concrete with specific deliverables

**Issue: Losing methodology compliance**

- Solution: Engage festival_methodology_manager.md
- Regular reviews with festival_review_agent.md

**Issue: Multi-system coordination**

- Solution: Use Interface Planning Extension for multi-system projects
- See extensions/interface-planning/ for templates and guidance

## Integration with Development Workflow

### With Version Control

```bash
your-project/
├── .git/
├── src/                    # Your code
├── festivals/              # Festival planning
│   ├── active/            # Current festival
│   └── .festival/         # This directory
└── README.md
```

### With CI/CD

- Parse YAML TODO files for progress metrics
- Generate dashboards from festival status
- Automate phase transitions based on completion
- Validate task completion criteria

### With Project Management Tools

- Export TODO.md to JIRA/Linear/etc.
- Generate Gantt charts from task dependencies
- Calculate velocity from completion rates
- Create burndown charts from progress data

## Summary

This directory contains everything needed to implement Festival Methodology successfully:

1. **Templates** - Start with these, customize as needed
2. **Agents** - Use for guidance and quality control
3. **Examples** - Learn from concrete implementations
4. **Documentation** - Understand the principles

Remember: Festival Methodology is a framework, not a prescription. Adapt it to your needs while maintaining the core principles of interface-first development and three-level hierarchy.

For questions or contributions, see the main [CONTRIBUTING.md](../../CONTRIBUTING.md) file.
