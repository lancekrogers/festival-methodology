# Festival Methodology Agent Registry

Specialized AI agents for Festival Methodology workflow management.

## Agent Overview

Festival Methodology includes three specialized agents that work together to ensure successful project execution:

**Agent Workflow:**

```
Planning Agent → Creates Festival Structure
     ↓
Review Agent → Validates Structure
     ↓
Manager Agent → Enforces Methodology
     ↓
Feedback Loop → Back to Planning
```

## Agent Catalog

### 🔵 festival_planning_agent

**Purpose:** Create comprehensive festival plans through structured interviews

**Trigger Conditions:**

- Starting a new project
- Converting existing project to festival methodology
- Major scope expansion requiring restructuring
- Creating a new major feature as sub-festival

**Key Capabilities:**

- Conducts systematic requirements discovery
- Creates three-level hierarchy (Phases → Sequences → Tasks)
- Generates all festival documentation
- Ensures interface-first architecture
- Plans quality gates and verification

**Sample Prompts:**

```
"Use the festival planning agent to help me create a festival for a user authentication system"

"I have a PRD for an e-commerce platform. Use the planning agent to create a festival structure"

"Help me plan a festival for migrating our monolith to microservices"
```

**Output:**

- Complete festival directory structure
- FESTIVAL_OVERVIEW.md with goals and success criteria
- Initial phase and sequence planning
- COMMON_INTERFACES.md structure
- FESTIVAL_RULES.md with project standards

---

### 🟢 festival_review_agent

**Purpose:** Validate festival structure and ensure methodology compliance

**Trigger Conditions:**

- Before transitioning between phases
- After major festival structure changes
- Weekly/sprint review checkpoints
- Before stakeholder presentations
- Post-completion assessment

**Key Capabilities:**

- Validates three-level hierarchy compliance
- Checks interface definition completeness
- Ensures quality gates are present
- Reviews task specificity and testability
- Identifies methodology violations
- Suggests structural improvements

**Sample Prompts:**

```
"Use the review agent to validate my festival structure before we start Phase 003"

"Review our current festival for methodology compliance"

"Check if our interface definitions are complete enough to begin implementation"
```

**Output:**

- Compliance report with specific issues
- Recommendations for improvement
- Risk assessment for methodology violations
- Checklist of items to address
- Quality score and readiness assessment

---

### 🔴 festival_methodology_manager

**Purpose:** Enforce methodology principles during active execution

**Trigger Conditions:**

- Daily during active development
- When methodology drift detected
- Team confusion about process
- Integration issues arising
- Quality problems emerging

**Key Capabilities:**

- Real-time methodology enforcement
- Process drift prevention
- Team guidance and correction
- Integration issue prevention
- Quality standard maintenance
- Parallel work coordination

**Sample Prompts:**

```
"Use the methodology manager to help us get back on track - we're starting to skip quality gates"

"Our team is confused about whether to start implementation. Engage the methodology manager"

"We're having integration issues. Use the manager agent to identify process problems"
```

**Output:**

- Specific corrective actions
- Process realignment plan
- Team guidance documentation
- Updated task priorities
- Integration conflict resolution

## Agent Collaboration Patterns

### Pattern 1: New Festival Creation

```
1. Planning Agent → Creates initial structure
2. Review Agent → Validates structure
3. Planning Agent → Refines based on feedback
4. Manager Agent → Guides execution
```

### Pattern 2: Methodology Rescue

```
1. Manager Agent → Identifies drift/problems
2. Review Agent → Comprehensive assessment
3. Planning Agent → Restructures problem areas
4. Manager Agent → Enforces corrections
```

### Pattern 3: Phase Transition

```
1. Review Agent → Phase completion check
2. Planning Agent → Next phase detail planning
3. Manager Agent → Transition coordination
```

## Agent Integration Guide

### With Claude Code

**Loading an Agent:**

```
You: Please load the festival planning agent to help me create a new project

Claude: [Reads festival_planning_agent.md and begins structured interview]
```

**Sequential Agent Use:**

```
You: First use the planning agent to create structure, then the review agent to validate

Claude: [Executes agents in sequence with handoff between them]
```

**Parallel Insights:**

```
You: Get perspectives from both review and manager agents on our current state

Claude: [Runs both agents and synthesizes insights]
```

### With Other AI Assistants

Agents are markdown files with clear instructions. Any AI assistant can:

1. Read the agent file
2. Adopt the agent's persona and expertise
3. Follow the agent's structured approach
4. Generate specified outputs

## Agent Customization

### Adapting Agents to Your Context

Each agent can be customized for your specific needs:

**Domain Specialization:**

```markdown
Add to planning agent:
"Additional expertise in [fintech/healthcare/gaming] requirements"
```

**Team Structure:**

```markdown
Add to manager agent:
"Coordinate with [QA team/DevOps/Product] using [specific process]"
```

**Tool Integration:**

```markdown
Add to review agent:
"Generate reports compatible with [JIRA/Linear/Asana]"
```

### Creating Custom Agents

Template for new agents:

```markdown
---
name: festival-[purpose]-agent
description: [One paragraph description with examples]
color: [blue/green/red/yellow]
---

You are a specialized AI assistant expert in [specific domain].

Your core expertise includes:
- [Expertise area 1]
- [Expertise area 2]
- [Expertise area 3]

Your approach:
1. [Step 1]
2. [Step 2]
3. [Step 3]

Your outputs:
- [Output type 1]
- [Output type 2]
```

## Agent Effectiveness Metrics

### Planning Agent Success Indicators

- ✅ Clear three-level hierarchy created
- ✅ All interfaces defined before implementation
- ✅ Quality gates at every sequence
- ✅ Concrete, testable tasks
- ✅ Parallel work opportunities identified

### Review Agent Success Indicators

- ✅ Methodology violations caught early
- ✅ Structure improvements suggested
- ✅ Quality issues identified
- ✅ Clear remediation steps provided
- ✅ Risk assessment accurate

### Manager Agent Success Indicators

- ✅ Process drift prevented
- ✅ Team stays on methodology
- ✅ Integration issues avoided
- ✅ Quality maintained throughout
- ✅ Parallel work coordinated smoothly

## Common Agent Scenarios

### Scenario 1: "Our project is off track"

**Use:** Manager Agent first (diagnosis), then Review Agent (assessment), then Planning Agent (restructure)

### Scenario 2: "Starting a greenfield project"

**Use:** Planning Agent (full structure), then Review Agent (validation)

### Scenario 3: "Ready to start coding"

**Use:** Review Agent (interface completeness check), then Manager Agent (execution guidance)

### Scenario 4: "Inheriting an existing project"

**Use:** Review Agent (current state), Planning Agent (retrofit to festival), Manager Agent (transition)

### Scenario 5: "Scope changed significantly"

**Use:** Planning Agent (restructure), Review Agent (validate changes), Manager Agent (coordinate transition)

## Agent Limitations

### What Agents DON'T Do

- ❌ Write actual code
- ❌ Make business decisions
- ❌ Replace human judgment
- ❌ Handle political/team dynamics
- ❌ Estimate timelines (use steps, not time)

### When NOT to Use Agents

- Simple, single-task work
- Well-understood, routine tasks
- Emergency hotfixes
- Prototyping/experimentation
- Personal learning projects

## Agent Evolution

Agents improve through:

1. **User Feedback** - Report what worked/didn't work
2. **Pattern Recognition** - Common issues become preventive checks
3. **Domain Expansion** - New project types add specializations
4. **Community Contributions** - Shared improvements benefit all

## Quick Reference Card

| Need | Agent | Key Prompt |
|------|-------|------------|
| Create new festival | Planning | "Create a festival for..." |
| Validate structure | Review | "Review my festival..." |
| Fix process issues | Manager | "Help us get back on track..." |
| Check readiness | Review | "Are we ready for Phase X?" |
| Prevent drift | Manager | "Monitor our execution..." |
| Restructure | Planning | "Restructure our festival..." |

## Integration Examples

### Example 1: CI/CD Integration

```yaml
# .github/workflows/festival-review.yml
- name: Festival Review Check
  run: |
    # Use review agent criteria as automated checks
    check_interfaces_defined
    check_quality_gates_present
    check_task_specificity
```

### Example 2: Daily Standup Integration

```
Team: "What should we focus on today?"
Run: Manager agent for daily priorities
Output: Specific tasks maintaining methodology
```

### Example 3: Sprint Planning

```
Team: "Planning next sprint"
Run: Planning agent for sequence detail
Run: Review agent for readiness check
Output: Sprint backlog aligned with festival
```

## Summary

The three Festival Methodology agents work as a team:

- **Planning Agent** - Creates and structures
- **Review Agent** - Validates and assesses  
- **Manager Agent** - Guides and enforces

Use them individually or in combination to maintain methodology excellence throughout your festival execution.
