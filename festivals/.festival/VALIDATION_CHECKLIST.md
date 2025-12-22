# Festival Methodology Validation Checklist

Use this checklist to ensure your festival follows methodology principles and is ready for execution.

## Festival Structure Validation

### ✅ Three-Level Hierarchy

- [ ] Festival has numbered phases (001, 002, 003, etc.)
- [ ] Each phase contains numbered sequences (01, 02, 03, etc.)
- [ ] Each sequence contains numbered tasks (01, 02, 03, etc.)
- [ ] Numbering indicates execution order and dependencies

### ✅ Goal Files Present

- [ ] `FESTIVAL_GOAL.md` exists at festival root
- [ ] Each phase has `PHASE_GOAL.md` in its directory
- [ ] Each sequence has `SEQUENCE_GOAL.md` in its directory
- [ ] Goals align: Sequence goals → Phase goals → Festival goal

### ✅ Standard Phases Present

- [ ] **001_PLAN** - Requirements and architecture defined
- [ ] **002_DEFINE_INTERFACES** - All interfaces specified
- [ ] **003_IMPLEMENT** - Implementation tasks created
- [ ] **004_REVIEW_AND_UAT** - Validation tasks defined

### ✅ Core Documentation

- [ ] `FESTIVAL_OVERVIEW.md` exists with clear goals
- [ ] `FESTIVAL_GOAL.md` tracks measurable objectives
- [ ] `FESTIVAL_RULES.md` defines project standards
- [ ] `COMMON_INTERFACES.md` specifies all contracts
- [ ] `TODO.md` tracks festival progress

## Phase-Specific Validation

### Phase 001_PLAN Checklist

- [ ] Requirements fully documented
- [ ] Stakeholders identified
- [ ] Success criteria defined
- [ ] Architecture decisions made
- [ ] Technology stack selected
- [ ] Initial risk assessment complete

### Phase 002_DEFINE_INTERFACES Checklist (CRITICAL)

- [ ] All API endpoints specified
- [ ] Data models fully defined
- [ ] Function signatures documented
- [ ] Integration points identified
- [ ] Error handling defined
- [ ] Example requests/responses provided
- [ ] **Interfaces locked before Phase 003**

### Phase 003_IMPLEMENT Checklist

- [ ] Tasks reference interface definitions
- [ ] Parallel work opportunities identified
- [ ] Dependencies clearly marked
- [ ] Each task has clear deliverables
- [ ] Testing approach defined

### Phase 004_REVIEW_AND_UAT Checklist

- [ ] Acceptance criteria defined
- [ ] Test scenarios created
- [ ] Review process documented
- [ ] Stakeholder sign-off process clear
- [ ] Deployment plan exists

## Sequence Validation

### ✅ Every Sequence Should Have

- [ ] Clear objective statement
- [ ] Numbered tasks in execution order
- [ ] Quality verification tasks:
  - [ ] `XX_testing_and_verify` task
  - [ ] `XX_code_review` task
  - [ ] `XX_review_results_iterate` task

### ✅ Task Dependencies

- [ ] Tasks with same number can run in parallel
- [ ] Sequential tasks have increasing numbers
- [ ] Dependencies explicitly stated
- [ ] No circular dependencies

## Task Quality Validation

### ✅ Good Task Characteristics

- [ ] **Specific objective** (not vague or abstract)
- [ ] **Concrete deliverables** (files, functions, outputs)
- [ ] **Testable requirements** (checklist items)
- [ ] **Implementation steps** (actionable instructions)
- [ ] **Definition of done** (clear completion criteria)

### ❌ Task Red Flags

- [ ] Objective longer than 2 sentences
- [ ] Requirements too abstract ("implement user management")
- [ ] No specific deliverables listed
- [ ] Missing implementation steps
- [ ] No testing/verification approach

## Interface Definition Validation

### ✅ API Endpoints Should Include

- [ ] HTTP method (GET, POST, etc.)
- [ ] Full path with parameters
- [ ] Request body schema
- [ ] Response body schema
- [ ] Error responses
- [ ] Authentication requirements
- [ ] Example requests/responses

### ✅ Data Models Should Include

- [ ] Field names and types
- [ ] Required vs optional fields
- [ ] Validation rules
- [ ] Relationships to other models
- [ ] Indexes and constraints
- [ ] Example instances

## Goal Validation

### ✅ Goal Alignment

- [ ] Each sequence goal directly supports its phase goal
- [ ] Each phase goal directly supports the festival goal
- [ ] No orphaned goals (goals not contributing to higher level)
- [ ] No missing goals (all levels have defined goals)

### ✅ Goal Measurability

- [ ] Festival goal has quantifiable KPIs
- [ ] Phase goals have specific success criteria
- [ ] Sequence goals have concrete deliverables
- [ ] All goals have evaluation frameworks

## Quality Gate Validation

### ✅ At Sequence Level

- [ ] SEQUENCE_GOAL.md evaluated
- [ ] Testing task validates all deliverables
- [ ] Review task checks code quality
- [ ] Iteration task addresses findings

### ✅ At Phase Level

- [ ] PHASE_GOAL.md evaluation completed
- [ ] All sequences complete before phase transition
- [ ] All sequence goals achieved
- [ ] Stakeholder review conducted
- [ ] Documentation updated
- [ ] Next phase prerequisites met

### ✅ At Festival Level

- [ ] FESTIVAL_GOAL.md evaluation completed
- [ ] All success criteria met
- [ ] All phase goals achieved
- [ ] All phases complete
- [ ] Final deliverables verified
- [ ] Stakeholder acceptance received

## Common Anti-Patterns to Avoid

### ❌ Methodology Violations

- [ ] Starting implementation before interfaces defined
- [ ] Skipping quality verification tasks
- [ ] Vague or abstract task definitions
- [ ] Missing dependencies between tasks
- [ ] No clear success criteria

### ❌ Process Smell

- [ ] Phases with only one sequence
- [ ] Sequences with only one task
- [ ] Tasks without concrete deliverables
- [ ] Interfaces defined during implementation
- [ ] Quality gates added as afterthought

## Progress Tracking Validation

### ✅ TODO.md Should Show

- [ ] Current phase and status
- [ ] Sequence completion percentages
- [ ] Task-level checkboxes
- [ ] Clear visual progress indicators
- [ ] Blockers and issues highlighted

### ✅ Progress Metrics

- [ ] Can calculate % complete at each level
- [ ] Can identify critical path
- [ ] Can see parallel work opportunities
- [ ] Can track velocity/burn rate

## Ready-to-Execute Checklist

### Before Starting Phase 001

- [ ] Festival overview documented
- [ ] Team understands methodology
- [ ] Tools and environment ready
- [ ] Stakeholders aligned

### Before Starting Phase 002

- [ ] Requirements fully understood
- [ ] Architecture decisions final
- [ ] All integration points identified
- [ ] Interface templates ready

### Before Starting Phase 003

- [ ] **All interfaces locked** (CRITICAL)
- [ ] Development environment ready
- [ ] Team assignments clear
- [ ] Parallel work plan created

### Before Starting Phase 004

- [ ] All implementation complete
- [ ] Test environment ready
- [ ] Review criteria defined
- [ ] Stakeholders available

## Continuous Validation

### Daily Checks

- [ ] Progress updated in TODO.md
- [ ] Blockers identified and communicated
- [ ] Methodology compliance maintained
- [ ] Quality gates not skipped

### Weekly Checks

- [ ] Festival structure still appropriate
- [ ] Interfaces still locked (Phase 003+)
- [ ] Progress tracking accurate
- [ ] Risk assessment updated

### Phase Transition Checks

- [ ] Current phase complete
- [ ] Next phase ready
- [ ] Stakeholder review done
- [ ] Lessons learned captured

## Validation Scoring

Count your checkmarks:

- **90-100% checked**: Festival is well-structured and ready
- **75-89% checked**: Address gaps before proceeding
- **60-74% checked**: Significant improvements needed
- **Below 60%**: Restructure using planning agent

## Quick Validation Commands

For automated validation (if tooling available):

```bash
# Check structure
find . -name "*.md" | grep -E "^[0-9]{2,3}_" | wc -l

# Check interfaces
grep -r "endpoint\|schema\|interface" COMMON_INTERFACES.md | wc -l

# Check quality gates
find . -name "*verify*.md" -o -name "*review*.md" | wc -l

# Check progress
grep -c "\[x\]" TODO.md
```

## When Validation Fails

If validation reveals issues:

1. **Minor Issues** (1-3 items):
   - Fix immediately
   - Update affected documentation
   - Continue execution

2. **Moderate Issues** (4-8 items):
   - Pause current work
   - Use review agent for assessment
   - Create correction tasks
   - Resume after fixes

3. **Major Issues** (9+ items):
   - Stop execution
   - Use planning agent to restructure
   - Use review agent to validate fixes
   - Restart from appropriate phase

## Summary

This checklist ensures your festival:

- Follows three-level hierarchy
- Defines interfaces before implementation
- Includes quality verification
- Maintains clear documentation
- Enables parallel development
- Tracks progress effectively

Regular validation prevents methodology drift and ensures systematic progress toward your goals.
