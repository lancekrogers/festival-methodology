---
id: FESTIVAL_QUICKSTART_TEMPLATE
aliases:
  - FESTIVAL QUICKSTART TEMPLATE
  - QUICKSTART TEMPLATE
tags: []
created: '2025-09-12'
modified: '2025-09-12'
---

<!--
TEMPLATE USAGE:
- All [REPLACE: ...] markers MUST be replaced with actual content
- Do NOT leave any [REPLACE: ...] markers in the final document
- Remove this comment block when filling the template
-->

# Festival Quickstart: [REPLACE: Project Name]

> **For Experienced Users**: This template provides minimal setup for teams familiar with Festival Methodology. For first-time users, use FESTIVAL_OVERVIEW_TEMPLATE.md instead.

## Project Goal

[REPLACE: ONE clear sentence describing what this festival will accomplish]

## Success Criteria

- [ ] [REPLACE: Primary functional outcome]
- [ ] [REPLACE: Quality/performance requirement]
- [ ] [REPLACE: Business/user value delivered]

## Phase Structure

**Phase Type:** [REPLACE: Simple | Multiple Implementation | Research First | Custom]

**Phases (add as needed):**

- **001_[REPLACE: PHASE_NAME]**: [REPLACE: Brief objective - planning/research phases may just be documents]
- **002_[REPLACE: PHASE_NAME]**: [REPLACE: Brief objective - implementation phases need sequences/tasks]
- **003_[REPLACE: PHASE_NAME]**: [REPLACE: Brief objective - add more implementation phases as needed]
- **[REPLACE: Additional phases]**: [REPLACE: Add phases as requirements emerge]

## Key Interfaces & Contracts (If Multi-System)

**Only needed for projects with multiple interacting systems:**

### API Contracts

- [REPLACE: Interface 1]: [REPLACE: Brief description] (optional)
- [REPLACE: Interface 2]: [REPLACE: Brief description] (optional)

### Data Contracts

- [REPLACE: Data structure 1]: [REPLACE: Brief description] (optional)
- [REPLACE: Data structure 2]: [REPLACE: Brief description] (optional)

### Component Interfaces

- [REPLACE: Component 1]: [REPLACE: Brief description]
- [REPLACE: Component 2]: [REPLACE: Brief description]

## Quality Standards

**This festival follows:**

- [ ] [REPLACE: Code standard reference]
- [ ] [REPLACE: Testing requirement]
- [ ] [REPLACE: Security requirement]
- [ ] [REPLACE: Performance requirement]

**Quality Gates:** Standard testing/review/iteration tasks included in all implementation sequences.

## High-Risk Areas

**Pay special attention to:**

- [REPLACE: Risk 1 and mitigation]
- [REPLACE: Risk 2 and mitigation]
- [REPLACE: Integration point that needs careful coordination]

## Dependencies & Constraints

**External Dependencies:**

- [REPLACE: External dependency description]

**Technical Constraints:**

- [REPLACE: Constraint 1]: [REPLACE: How it affects the approach]
- [REPLACE: Constraint 2]: [REPLACE: How it affects the approach]

**Team Constraints:**

- Team size: [REPLACE: number]
- Key skills: [REPLACE: critical skills needed]
- Key dependencies: [REPLACE: blocking steps that must complete first]

## Sequence Planning Notes

**Parallel Work Opportunities:**

- After Phase [REPLACE: X] completion: [REPLACE: List sequences that can run in parallel]
- [REPLACE: Interface Y] enables: [REPLACE: List dependent work that can start]

**Critical Path:**

- [REPLACE: Sequence that blocks other work]
- [REPLACE: Dependency that could delay the project]

## Stakeholders

**Key Decision Makers:**

- [REPLACE: Name/Role]: [REPLACE: What they need to approve]
- [REPLACE: Name/Role]: [REPLACE: What they need to approve]

**Primary Users:**

- [REPLACE: User type]: [REPLACE: Their main need/concern]
- [REPLACE: User type]: [REPLACE: Their main need/concern]

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
