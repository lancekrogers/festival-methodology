# Festival Methodology Resources Guide

This directory contains all the resources needed to implement Festival Methodology in your projects. This guide helps you navigate and use these resources effectively.

## 📚 Just-In-Time Reading Strategy

**CRITICAL**: To preserve context window, follow these rules:

### When to Read What

| Resource | Read When |
|----------|-----------|
| This README | Immediately - provides navigation |
| Core methodology docs | During initial understanding |
| Templates | ONLY when creating that specific document |
| Examples | ONLY when stuck or need clarification |
| Agents | ONLY when using that specific agent |

### Never Do This
❌ Reading all templates upfront "to understand them"
❌ Re-reading templates you've already used
❌ Reading examples before trying yourself
❌ Loading all agents at once

### Always Do This
✅ Read templates one at a time as needed
✅ Read examples only when stuck
✅ Keep templates closed after use
✅ Focus context on actual work, not documentation

## Quick Navigation

- **[Templates](#templates)** - Document templates for creating festivals
- **[Agents](#ai-agents)** - Specialized AI agents for festival workflow
- **[Examples](#examples)** - Concrete examples and patterns
- **[Core Documentation](#core-documentation)** - Methodology principles and theory

## Goal Files - New!

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

**At Planning Time:**
1. Create FESTIVAL_GOAL.md from template
2. Create PHASE_GOAL.md for each phase
3. Create SEQUENCE_GOAL.md for each sequence
4. Ensure alignment: Sequence goals → Phase goals → Festival goal

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
   - *Use this first when starting a new festival*

2. **COMMON_INTERFACES_TEMPLATE.md** 
   - Define all system interfaces before implementation
   - Protocol-agnostic interface definitions
   - Enables parallel development
   - *Critical for Phase 002_DEFINE_INTERFACES*

3. **FESTIVAL_RULES_TEMPLATE.md**
   - Project-specific standards and guidelines
   - Quality gates and compliance requirements
   - Team agreements and conventions
   - *Customize for your project's needs*

4. **TASK_TEMPLATE.md**
   - Comprehensive task structure (full version)
   - Detailed implementation steps
   - Testing and verification sections
   - *Use for complex or critical tasks*

5. **FESTIVAL_TODO_TEMPLATE.md** (Markdown)
   - Human-readable progress tracking
   - Checkbox-based task management
   - Visual project status
   - *Use for manual tracking and documentation*

6. **FESTIVAL_TODO_TEMPLATE.yaml** (YAML)
   - Machine-readable progress tracking
   - Structured data for automation
   - CI/CD integration ready
   - *Use for automated tooling and reporting*

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
   - *Trigger: Starting a new project or festival*

2. **festival_review_agent.md**
   - Validates festival structure compliance
   - Reviews quality gates
   - Ensures methodology adherence
   - *Trigger: Before moving phases or major milestones*

3. **festival_methodology_manager.md**
   - Enforces methodology during execution
   - Prevents process drift
   - Provides ongoing governance
   - *Trigger: During active development*

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
   - *Read this first to understand the "why"*

2. **PROJECT_MANAGEMENT_SYSTEM.md**
   - Markdown/YAML tracking system
   - Progress calculation methods
   - Automation opportunities
   - *Read this to understand tracking mechanics*

## Creating Your First Festival

### Quick Start Process

1. **Understand the Goal**
   ```
   Read: FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md (core principles)
   Time: 10 minutes
   ```

2. **Plan the Festival**
   ```
   Use: festival_planning_agent.md
   Creates: Initial festival structure
   Time: 20-30 minutes
   ```

3. **Create Core Documents**
   ```
   Templates needed:
   - FESTIVAL_OVERVIEW_TEMPLATE.md → FESTIVAL_OVERVIEW.md
   - FESTIVAL_RULES_TEMPLATE.md → FESTIVAL_RULES.md
   - COMMON_INTERFACES_TEMPLATE.md → COMMON_INTERFACES.md
   ```

4. **Structure Phases**
   ```
   Standard phases:
   - 001_PLAN
   - 002_DEFINE_INTERFACES
   - 003_IMPLEMENT
   - 004_REVIEW_AND_UAT
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
created: 'YYYY-MM-DD'
modified: 'YYYY-MM-DD'
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

**Issue: Unclear interfaces**
- Solution: Spend more time on Phase 002
- Use COMMON_INTERFACES_TEMPLATE.md thoroughly

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