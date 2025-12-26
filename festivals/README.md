# Festival Methodology - AI Agent Build System

## 🛑 MANDATORY FIRST STEPS - DO NOT SKIP

### Step 1: Verify Methodology Resources Exist

YOU MUST EXECUTE THIS COMMAND:

```bash
ls -la .festival/
```

### Step 2: Read the Implementation Guide

YOU MUST READ NOW: `.festival/README.md`
This contains the methodology overview and will guide you through the process.

### Step 3: Understand Core Methodology

READ THESE CORE DOCUMENTS NOW:

- `.festival/FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md` - Core methodology principles
- `.festival/PROJECT_MANAGEMENT_SYSTEM.md` - How tracking works

## ⚠️ IMPORTANT: Template Reading Strategy

**DO NOT READ TEMPLATES UNTIL YOU NEED THEM**
Templates are in `.festival/templates/` but should ONLY be read when you reach the specific step requiring them. This preserves context window.

## Festival Workflow

### Planning Phase

1. **Understand the user's goal** through discussion
2. **WHEN READY TO PLAN**: Read `.festival/agents/festival_planning_agent.md`
3. The planning agent will guide you through creating the structure

### Structure Creation Phase

**ONLY when ready to create each document, read its template:**

1. **When creating project overview**: Read `.festival/templates/FESTIVAL_OVERVIEW_TEMPLATE.md`
2. **When creating festival goals**: Read `.festival/templates/FESTIVAL_GOAL_TEMPLATE.md`
3. **When defining standards**: Read `.festival/templates/FESTIVAL_RULES_TEMPLATE.md`
4. **When defining interfaces** (Phase 002): Read `.festival/templates/COMMON_INTERFACES_TEMPLATE.md`
5. **When creating phase goals**: Read `.festival/templates/PHASE_GOAL_TEMPLATE.md`
6. **When creating sequence goals**: Read `.festival/templates/SEQUENCE_GOAL_TEMPLATE.md`
7. **When creating tasks**: Read `.festival/templates/TASK_TEMPLATE.md` or `TASK_TEMPLATE_SIMPLE.md`
8. **When setting up tracking**: Read `.festival/templates/FESTIVAL_TODO_TEMPLATE.md`
9. **When capturing decisions**: Read `.festival/templates/CONTEXT_TEMPLATE.md`

### Execution Phase

- **For quality review**: Read `.festival/agents/festival_review_agent.md` ONLY when review is needed
- **For methodology enforcement**: Read `.festival/agents/festival_methodology_manager.md` ONLY during execution
- **For examples**: Read files in `.festival/examples/` ONLY when you need concrete examples

## ⚠️ Context Preservation Rules

1. **Never read all templates at once** - Read each template only when creating that specific document
2. **Don't re-read templates** - Once you understand a template's structure, don't read it again
3. **Use examples sparingly** - Only read examples when stuck or need clarification
4. **Preserve context for execution** - Save your context window for the actual work, not documentation

## Directory Structure

```text
festivals/                          # Your festival workspace
├── planned/                        # Festivals being planned
├── active/                         # Currently executing festivals
├── completed/                      # Finished festivals
├── dungeon/                        # Deprioritized or cancelled work (archived)
├── .festival/                      # Methodology resources (DO NOT READ ALL AT ONCE)
└── README.md                       # This file - agent instructions
```

## Verification Checkpoint

Before proceeding with ANY festival work, confirm:

- [ ] I have executed `ls -la .festival/`
- [ ] I have read `.festival/README.md`
- [ ] I understand I should NOT read templates until needed
- [ ] I will read each template ONLY when creating that specific document

## Creating Your Festival - Step by Step

### Phase 1: Understanding (DO NOT READ TEMPLATES YET)

- Discuss the project goals with the user
- Understand scope and requirements
- Identify complexity level

### Phase 2: Planning (READ PLANNING AGENT ONLY)

- NOW read: `.festival/agents/festival_planning_agent.md`
- Follow the agent's structured interview process
- Create initial structure in `planned/[festival-name]/`

### Phase 3: Document Creation (READ EACH TEMPLATE AS NEEDED)

For each document you need to create:

1. Identify which document you're creating
2. READ the corresponding template ONLY NOW
3. Create the document
4. Move to next document

### Phase 4: Execution (READ AGENTS AS NEEDED)

- Start work on tasks
- READ review agent ONLY when quality check is needed
- READ methodology manager ONLY if process questions arise

## Festival Organization

### planned/

New festivals being designed. Create festival directories here during planning.

### active/

Festivals currently being executed. Move from `planned/` when execution begins.

### completed/

Successfully finished festivals. Move from `active/` when all objectives achieved.

### dungeon/

Archived festivals - cancelled, deprioritized, or paused work. Move here to keep workspace clean.

## Using the fest CLI (Recommended for AI Agents)

The `fest` CLI provides efficient festival management that **saves tokens and ensures structure consistency**. Always use CLI commands instead of manually creating files.

### Why Use CLI Instead of Manual File Creation

| Approach | Token Cost | Structure Guarantee | Speed |
|----------|------------|---------------------|-------|
| Manual file creation | High (read template + write file) | Error-prone | Slow |
| CLI with `--json` | Low (single command) | Guaranteed correct | Fast |

### Quick Reference

```bash
# Create a new festival
fest create festival --name "my-project" --goal "Build X" --json

# Create phases
fest create phase --name "PLAN" --json
fest create phase --name "IMPLEMENT" --json

# Create sequences (run from inside phase directory)
fest create sequence --name "requirements" --json

# Create tasks (run from inside sequence directory)
fest create task --name "analyze_requirements" --json

# Sync quality gate tasks to all sequences
fest task defaults sync --dry-run --json
fest task defaults sync --approve --json
```

### Agent Workflow Pattern

1. **Use `fest create` commands** with `--json` flag for machine-readable output
2. **Parse JSON response** to confirm success and get created file paths
3. **Use `fest task defaults sync`** after creating sequences to add quality gates
4. **Edit only content sections** of generated files - structure is already correct

### Available Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `fest create festival` | Create festival scaffold | Starting new project |
| `fest create phase` | Add phase to festival | Building structure |
| `fest create sequence` | Add sequence to phase | Grouping tasks |
| `fest create task` | Add task to sequence | Creating work items |
| `fest task defaults sync` | Sync quality gates | After creating sequences |
| `fest renumber` | Fix numbering gaps | After removing elements |
| `fest count` | Token estimation | Planning context usage |

### JSON Output

All commands support `--json` for automation:

```json
{
  "ok": true,
  "action": "create_festival",
  "festival": {
    "name": "my-project",
    "path": "/path/to/festivals/active/my-project"
  },
  "created": [
    "FESTIVAL_OVERVIEW.md",
    "FESTIVAL_GOAL.md",
    "TODO.md"
  ]
}
```

### Quality Gate Configuration

Create `fest.yaml` in your festival root to customize quality gate tasks:

```yaml
version: "1.0"
quality_gates:
  enabled: true
  tasks:
    - id: testing_and_verify
      template: QUALITY_GATE_TESTING
      enabled: true
    - id: code_review
      template: QUALITY_GATE_REVIEW
      enabled: true
    - id: review_results_iterate
      template: QUALITY_GATE_ITERATE
      enabled: true
```

For full CLI documentation, see `.festival/CLI_REFERENCE.md`.

## Remember: Preserve Context

Your context window is precious. Reading all documentation upfront wastes context that should be used for actual work. Follow this README's instructions to read resources just-in-time.

---

**For Agents**: This README provides mandatory first steps and a just-in-time reading strategy. You MUST follow the verification steps at the top before any festival work. Read templates and examples ONLY when you need them, not before.
