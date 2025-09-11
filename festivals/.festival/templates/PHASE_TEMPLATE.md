# Phase: [NNN_PHASE_NAME]

**Phase Number:** [NNN] | **Status:** [Planning/Active/Complete] | **Type:** [Planning/Definition/Implementation/Validation]

> 📋 **Important**: Create a `PHASE_GOAL.md` file in this phase directory using the PHASE_GOAL_TEMPLATE.md to define specific goals and evaluation criteria for this phase.

## Objective
[Clear statement of what this phase will accomplish and its role in the overall festival]

## Success Criteria
- [ ] [Major deliverable or milestone 1]
- [ ] [Major deliverable or milestone 2]
- [ ] [Major deliverable or milestone 3]

## Planned Sequences

### [01_sequence_name]
**Purpose:** [What this sequence accomplishes]
**Dependencies:** [None or list other sequences]
**Deliverables:**
- [Key deliverable 1]
- [Key deliverable 2]

### [02_sequence_name]
**Purpose:** [What this sequence accomplishes]
**Dependencies:** [Previous sequences]
**Deliverables:**
- [Key deliverable 1]
- [Key deliverable 2]

### [03_sequence_name]
**Purpose:** [What this sequence accomplishes]
**Dependencies:** [Previous sequences]
**Deliverables:**
- [Key deliverable 1]
- [Key deliverable 2]

## Phase Dependencies

**Prerequisites (from previous phases):**
- [What must be complete before this phase can begin]

**Provides (to subsequent phases):**
- [What this phase produces that later phases need]

## Parallel Work Opportunities
[Identify sequences that can run in parallel]
- Sequences 01 and 02 can run in parallel
- Sequence 03 depends on both 01 and 02

## Risk Assessment
| Risk | Impact | Mitigation Strategy |
|------|--------|-------------------|
| [Risk description] | [High/Medium/Low] | [How to prevent or handle] |

## Quality Gates
Before this phase is considered complete:
- [ ] PHASE_GOAL.md evaluation completed
- [ ] All sequences completed
- [ ] All sequence goals achieved
- [ ] All deliverables verified
- [ ] Stakeholder review conducted
- [ ] Documentation updated
- [ ] Next phase ready to begin

## Notes
[Additional context, constraints, or considerations for this phase]

---

## Usage Guide

Phases represent major milestones in your festival. Standard phases include:

1. **001_PLAN** - Requirements, architecture, initial planning
2. **002_DEFINE_INTERFACES** - Critical phase for defining all contracts
3. **003_IMPLEMENT** - Parallel implementation based on interfaces
4. **004_REVIEW_AND_UAT** - Testing, validation, and acceptance

Customize phases based on your project needs.

### Example (Filled Out):

# Phase: 002_DEFINE_INTERFACES

**Phase Number:** 002 | **Status:** Planning | **Type:** Definition

## Objective
Define all system interfaces, contracts, and specifications to enable parallel development in Phase 003.

## Success Criteria
- [ ] All API endpoints fully specified
- [ ] All data models defined with schemas
- [ ] All function signatures documented
- [ ] All integration points identified
- [ ] Interface documentation reviewed and approved

## Planned Sequences

### 01_api_design
**Purpose:** Design and document all REST API endpoints
**Dependencies:** None (can start immediately)
**Deliverables:**
- OpenAPI specification
- Endpoint documentation
- Error code definitions

### 02_data_models
**Purpose:** Define all data structures and database schemas
**Dependencies:** None (can run parallel with 01)
**Deliverables:**
- Database schema diagrams
- Model definitions
- Relationship mappings

### 03_integration_contracts
**Purpose:** Define external service integrations
**Dependencies:** 01_api_design, 02_data_models
**Deliverables:**
- Third-party API usage docs
- Webhook specifications
- Event contracts

### 04_frontend_contracts
**Purpose:** Define frontend-backend contracts
**Dependencies:** 01_api_design, 02_data_models
**Deliverables:**
- Component prop interfaces
- State management structure
- Route definitions

### 05_review_and_finalize
**Purpose:** Review all interfaces and lock for implementation
**Dependencies:** All previous sequences
**Deliverables:**
- Approved interface documentation
- Implementation ready signal

## Phase Dependencies

**Prerequisites (from previous phases):**
- Requirements fully documented (Phase 001)
- Architecture decisions made (Phase 001)
- Technology stack selected (Phase 001)

**Provides (to subsequent phases):**
- Complete interface specifications for Phase 003
- Test scenarios derived from interfaces
- Integration test plans

## Parallel Work Opportunities
- Sequences 01 and 02 can run completely in parallel
- Sequences 03 and 04 can start once 01 and 02 produce initial drafts
- Frontend and backend teams can begin planning based on draft interfaces

## Risk Assessment
| Risk | Impact | Mitigation Strategy |
|------|--------|-------------------|
| Incomplete interface definition | High | Multiple review cycles, stakeholder signoff |
| Interface changes during implementation | High | Lock interfaces before Phase 003, change control process |
| Missing edge cases | Medium | Comprehensive examples in specifications |

## Quality Gates
Before this phase is considered complete:
- [ ] All sequences completed
- [ ] Technical review conducted
- [ ] Stakeholder approval received
- [ ] No undefined interfaces remain
- [ ] Implementation teams confirm clarity

## Notes
This is the most critical phase for enabling parallel development. Time invested here saves multiples during implementation. All interfaces must be locked before proceeding to Phase 003.