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

Phases represent major steps toward your goal. Common patterns:

**Planning Phases** (often unstructured):
- Requirements gathering, research, documentation
- May just contain documents, no sequences needed

**Implementation Phases** (must be structured):
- Where AI agents execute tasks autonomously
- Must have sequences and tasks
- Add as many as needed (CORE, FEATURES, UI, etc.)

**Validation Phases**:
- User acceptance, reviews, deployment prep

Add phases as requirements emerge, not all upfront.

### Example (Filled Out):

# Phase: 002_IMPLEMENT_CORE

**Phase Number:** 002 | **Status:** Planning | **Type:** Implementation

## Objective
Implement core system functionality including authentication, database layer, and basic API endpoints.

## Success Criteria
- [ ] User authentication system working
- [ ] Database connections established
- [ ] Core API endpoints functional
- [ ] Basic error handling implemented
- [ ] Unit tests passing
- [ ] Code review completed

## Planned Sequences

### 01_authentication_system
**Purpose:** Implement user registration, login, and session management
**Dependencies:** None (can start immediately)
**Deliverables:**
- User model and database
- Login/logout endpoints
- JWT token management

### 02_database_layer
**Purpose:** Set up database connections and models
**Dependencies:** None (can run parallel with 01)
**Deliverables:**
- Database configuration
- Core data models
- Migration scripts

### 03_core_api_endpoints
**Purpose:** Implement basic CRUD operations
**Dependencies:** 01_authentication_system, 02_database_layer
**Deliverables:**
- REST endpoints for core entities
- Input validation
- Response formatting

## Phase Dependencies

**Prerequisites (from previous phases):**
- Requirements fully documented (Phase 001)
- Architecture decisions made (Phase 001)
- Technology stack selected (Phase 001)

**Provides (to subsequent phases):**
- Working core functionality for feature implementation
- Database schema for additional entities
- Authentication system for user features

## Parallel Work Opportunities
- Sequences 01 and 02 can run completely in parallel
- Sequence 03 depends on both 01 and 02

## Risk Assessment
| Risk | Impact | Mitigation Strategy |
|------|--------|-------------------|
| Database performance issues | High | Early performance testing, indexing strategy |
| Security vulnerabilities | High | Security review, penetration testing |
| Integration problems | Medium | Continuous integration testing |

## Quality Gates
Before this phase is considered complete:
- [ ] All sequences completed
- [ ] All tests passing
- [ ] Code review conducted
- [ ] Security review completed
- [ ] Documentation updated

## Notes
This phase implements the foundation that all other features will build upon. Focus on stability and extensibility.