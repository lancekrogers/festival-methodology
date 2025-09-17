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

**Selected Pattern:** [Standard 4-phase | Already Planned | Iterative | Custom]

**Phase Overview:**
- **[001_PHASE_NAME]**: [Brief objective]
- **[002_PHASE_NAME]**: [Brief objective]
- **[003_PHASE_NAME]**: [Brief objective]
- **[004_PHASE_NAME]**: [Brief objective] (if applicable)

## Key Interfaces & Contracts

**Critical interfaces that enable parallel work:**

### API Contracts
- [Interface 1]: [Brief description]
- [Interface 2]: [Brief description]

### Data Contracts  
- [Data structure 1]: [Brief description]
- [Data structure 2]: [Brief description]

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
- [Dependency 1]: [Impact/timeline]
- [Dependency 2]: [Impact/timeline]

**Technical Constraints:**
- [Constraint 1]: [How it affects the approach]
- [Constraint 2]: [How it affects the approach]

**Team Constraints:**
- Team size: [number]
- Key skills: [critical skills needed]
- Timeline: [key milestones]

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
- Time is a constraint for setup
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