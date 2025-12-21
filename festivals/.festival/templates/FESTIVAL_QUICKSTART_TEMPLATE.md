---
id: FESTIVAL_QUICKSTART_TEMPLATE
aliases:
  - FESTIVAL QUICKSTART TEMPLATE
  - QUICKSTART TEMPLATE
tags: []
created: '2025-09-12'
modified: '2025-09-12'
---

# Festival Quickstart: [Project Name]

> **For Experienced Users**: This template provides minimal setup for teams familiar with Festival Methodology. For first-time users, use FESTIVAL_OVERVIEW_TEMPLATE.md instead.

## Project Goal

[ONE clear sentence describing what this festival will accomplish]

## Success Criteria

- [ ] [Primary functional outcome]
- [ ] [Quality/performance requirement]  
- [ ] [Business/user value delivered]

## Phase Structure

**Phase Type:** [Simple | Multiple Implementation | Research First | Custom]

**Phases (add as needed):**

- **001_[NAME]**: [Brief objective - planning/research phases may just be documents]
- **002_[NAME]**: [Brief objective - implementation phases need sequences/tasks]
- **003_[NAME]**: [Brief objective - add more implementation phases as needed]
- **[Additional]**: [Add phases as requirements emerge]

## Key Interfaces & Contracts (If Multi-System)

**Only needed for projects with multiple interacting systems:**

### API Contracts

- [Interface 1]: [Brief description] (optional)
- [Interface 2]: [Brief description] (optional)

### Data Contracts

- [Data structure 1]: [Brief description] (optional)
- [Data structure 2]: [Brief description] (optional)

### Component Interfaces

- [Component 1]: [Brief description]
- [Component 2]: [Brief description]

## Quality Standards

**This festival follows:**

- [ ] [Code standard reference]
- [ ] [Testing requirement]
- [ ] [Security requirement]
- [ ] [Performance requirement]

**Quality Gates:** Standard testing/review/iteration tasks included in all implementation sequences.

## High-Risk Areas

**Pay special attention to:**

- [Risk 1 and mitigation]
- [Risk 2 and mitigation]
- [Integration point that needs careful coordination]

## Dependencies & Constraints

**External Dependencies:**

**Technical Constraints:**

- [Constraint 1]: [How it affects the approach]
- [Constraint 2]: [How it affects the approach]

**Team Constraints:**

- Team size: [number]
- Key skills: [critical skills needed]
- Key dependencies: [blocking steps that must complete first]

## Sequence Planning Notes

**Parallel Work Opportunities:**

- After Phase [X] completion: [List sequences that can run in parallel]
- [Interface Y] enables: [List dependent work that can start]

**Critical Path:**

- [Sequence that blocks other work]
- [Dependency that could delay the project]

## Stakeholders

**Key Decision Makers:**

- [Name/Role]: [What they need to approve]
- [Name/Role]: [What they need to approve]

**Primary Users:**

- [User type]: [Their main need/concern]
- [User type]: [Their main need/concern]

## Quick Setup Checklist

- [ ] Create festival directory structure
- [ ] Copy this file to `FESTIVAL_OVERVIEW.md`
- [ ] Create `COMMON_INTERFACES.md` (CRITICAL - do this first!)
- [ ] Create `FESTIVAL_RULES.md` if project-specific standards needed
- [ ] Set up phase directories and PHASE_GOAL.md files
- [ ] Create initial sequences with proper quality gates
- [ ] Set up progress tracking (TODO.md or YAML)

## Next Steps

1. **Define Interfaces** (if using interface-first approach)
   - Complete COMMON_INTERFACES.md before implementation starts
   - Get team alignment on all contracts

2. **Create Detailed Tasks**
   - Use TASK_TEMPLATE.md for complex tasks
   - Use TASK_TEMPLATE_SIMPLE.md for routine tasks
   - Reference TASK_EXAMPLES.md for patterns

3. **Begin Execution**
   - Update CONTEXT.md with key decisions
   - Track progress in TODO.md
   - Hold regular review checkpoints

---

## Template Usage Notes

**Use this template when:**

- Team is familiar with Festival Methodology
- Project requirements are already clear
- Quick setup is preferred
- You need minimal documentation overhead

**Upgrade to full templates when:**

- Stakeholders need detailed documentation
- Team members are new to Festival Methodology  
- Project has complex compliance requirements
- Multiple teams need coordination

**Customization Tips:**

- Remove sections that don't apply to your project
- Add domain-specific sections as needed
- Adjust quality standards to match project risk
- Include project-specific constraints and considerations
