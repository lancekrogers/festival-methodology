# Festival Methodology - AI Agent Build System

Festival methodology is a goal-oriented project management system for AI agent development workflows. It uses a three-level hierarchy: **Phases → Sequences → Tasks**.

## Directory Structure

```
festivals/                          # Your festival workspace
├── planned/                        # Festivals being planned
├── active/                         # Currently executing festivals  
├── completed/                      # Finished festivals
├── archived/                       # Deprioritized or cancelled work
├── .festival/                      # Methodology resources (hidden)
└── README.md                       # This file - agent instructions
```

## Agent Instructions

### Step 1: Understand the Methodology

**Read these files to understand Festival Methodology:**

- `.festival/FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md` - Core methodology documentation
- `.festival/PROJECT_MANAGEMENT_SYSTEM.md` - Project tracking system explanation

### Step 2: Plan the Festival

**Read for festival planning:**

- `.festival/README.md` - Implementation guide and agent usage
- Use `.festival/agents/festival_planning_agent.md` for guided planning

### Step 3: Create Festival Structure

**When ready to create documents, read templates:**

- `.festival/templates/FESTIVAL_OVERVIEW_TEMPLATE.md` - Project goals and success criteria
- `.festival/templates/FESTIVAL_RULES_TEMPLATE.md` - Project standards
- `.festival/templates/COMMON_INTERFACES_TEMPLATE.md` - Interface definition system
- `.festival/templates/TASK_TEMPLATE.md` - Individual task structure
- `.festival/templates/FESTIVAL_TODO_TEMPLATE.md` - Project tracking system

**For reference during creation:**

- `.festival/examples/TASK_EXAMPLES.md` - Concrete task examples
- `.festival/examples/FESTIVAL_TODO_EXAMPLE.md` - Project tracking example

### Step 4: Execute and Manage

**During festival execution:**

- Use `.festival/agents/festival_review_agent.md` - Quality validation
- Use `.festival/agents/festival_methodology_manager.md` - Process enforcement
- Track progress using `TODO.md` based on `FESTIVAL_TODO_TEMPLATE.md`

## Festival Organization

### planned/

New festivals being designed. Create festival directories here during planning.

### active/  

Festivals currently being executed. Move from `planned/` when execution begins.

### completed/

Successfully finished festivals. Move from `active/` when all objectives achieved.

### archived/

Cancelled, deprioritized, or paused work. Move here to keep workspace clean.

## Quick Start for Agents

1. **Planning a Festival**: Read `.festival/FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md` for methodology understanding
2. **Creating Festival**: Use templates in `.festival/templates/` to create actual project documents  
3. **Need Examples**: Reference files in `.festival/examples/` for concrete guidance
4. **Quality Assurance**: Use agents in `.festival/agents/` for guidance and validation

## Context Management

**Read methodology files FIRST** to understand principles before reading templates. Only read templates when ready to create specific documents. This saves context and ensures proper understanding of the methodology before implementation.

**Template Usage**: Copy template content as starting structure, then customize with actual project requirements. Do not re-read templates once you understand their structure.

**Agent Workflow**: Use specialized agents (planning, review, manager) for guided assistance rather than reading all documentation every time.

---

**For Agents**: This README.md provides the complete roadmap. Follow the step-by-step instructions and read only the referenced files needed for your current task. The `.festival/` directory contains all methodology resources organized for efficient context usage.

