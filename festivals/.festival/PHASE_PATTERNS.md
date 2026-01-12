# Festival Goal Achievement Patterns

This guide provides common step progression patterns for different types of goals. **CRITICAL**: Implementation steps can only be created after goal requirements are defined. Use these patterns as starting points and adapt them based on your specific goal achievement needs.

## Goal-First Step Progression Principle

**FUNDAMENTAL UNDERSTANDING**: All step progressions in these patterns assume that goal requirements have been defined through:

- Completed planning steps with clear goal deliverables
- External planning documents provided by human defining goal achievement criteria
- Specific requirements and specifications for goal achievement from human
- Clear goal definitions with step completion criteria

**Never create implementation step progressions when:**

- Goal requirements are unclear or undefined
- Goal planning steps haven't been completed
- You're guessing what steps might achieve the goal

**CRITICAL**: Festival Methodology thinks in STEPS toward goal achievement, not time estimates. Focus on logical progression and dependencies, leveraging hyper-efficient AI-human collaboration that makes traditional time planning obsolete.

## Standard 3-Step Goal Achievement Pattern

**Use when:** Most development scenarios - clean goal progression from planning to completion.

```
001_PLAN → 002_IMPLEMENT → 003_REVIEW_AND_UAT
```

**Goal progression:** Complete goal achievement from conception to delivery.

**Step sequences per phase:**

- 001_PLAN: 2-3 step sequences (goal requirements, architecture, planning)
- 002_IMPLEMENT: 3-6 step sequences (backend steps, frontend steps, integration steps)
- 003_REVIEW_AND_UAT: 2-3 step sequences (validation steps, goal verification, completion steps)

**For Multi-System Projects:** Consider the [Interface Planning Extension](../extensions/interface-planning/) to add interface definition phases when system coordination is needed.

## Streamlined Patterns

### Already Planned Project

**Use when:** Human has provided requirements and architecture from external planning.

```
001_IMPLEMENT → 002_REVIEW_AND_UAT
```

**Requirements Source:** External planning documents provided by human.
**Focus:** Structure existing requirements into implementation sequences and validate completion.

### Simple Enhancement

**Use when:** Adding features to existing systems.

```
001_ANALYZE → 002_IMPLEMENT → 003_VERIFY
```

**Minimal approach:** Just enough process for quality work.

### Bug Fix or Maintenance

**Use when:** Focused problem-solving work.

```
001_INVESTIGATE → 002_FIX → 003_VALIDATE
```

**Ultra-minimal:** Get to the solution efficiently.

## Iterative Development Patterns

### Multi-Stage Implementation

**Use when:** Large systems need staged rollout.

```
001_PLAN → 
002_IMPLEMENT_CORE → 003_IMPLEMENT_FEATURES → 004_IMPLEMENT_POLISH → 
005_REVIEW_AND_UAT
```

**Benefits:** Allows incremental delivery and validation at each stage.
**Approach:** Build foundation first, then layer features and polish.

### MVP + Enhancement Pattern

**Use when:** Need to get something working quickly, then improve.

```
001_PLAN_MVP → 002_IMPLEMENT_MVP → 003_VALIDATE_MVP →
004_PLAN_ENHANCEMENT → 005_IMPLEMENT_ENHANCEMENT → 006_FINAL_REVIEW
```

**Benefits:** Early validation, iterative improvement.
**Risk:** Technical debt if MVP shortcuts aren't addressed.

### Feature Wave Pattern

**Use when:** Multiple related features need coordinated development.

```
001_PLAN →
002_IMPLEMENT_WAVE1 → 003_IMPLEMENT_WAVE2 → 004_IMPLEMENT_WAVE3 →
005_INTEGRATION → 006_REVIEW_AND_UAT
```

**Benefits:** Parallel team work, staged feature delivery.
**Coordination:** Requires clear feature boundaries and integration planning.

## Research and Exploration Patterns

### Research-Heavy Project

**Use when:** Significant unknowns need exploration before building.

```
001_RESEARCH → 002_PROTOTYPE → 003_EVALUATE →
004_PLAN → 005_IMPLEMENT → 006_VALIDATE
```

**Extended discovery:** Additional research steps upfront to reduce later risk.
**Decision points:** Each phase includes go/no-go evaluation.

### Technology Evaluation

**Use when:** Choosing between technology options.

```
001_REQUIREMENTS → 002_EVALUATE_OPTIONS → 003_PROTOTYPE_FINALIST →
004_DECIDE → 005_IMPLEMENT → 006_REVIEW
```

**Comparative approach:** Parallel evaluation of alternatives.
**Evidence-based:** Decisions backed by working prototypes.

### Innovation/Experiment

**Use when:** Building something genuinely new.

```
001_EXPLORE → 002_HYPOTHESIS → 003_EXPERIMENT →
004_LEARN → 005_ITERATE → 006_PRODUCTIONIZE
```

**Learning-focused:** Expect multiple iterations and pivots.
**Validation:** Each phase validates assumptions.

## Research Phase Structure and Document Types

Research phases (001_RESEARCH, 001_INVESTIGATE, etc.) use a **freeform subdirectory structure** instead of numbered sequences and tasks. This allows flexible organization of research documents.

### Research Document Types

The Festival Methodology provides four specialized research document templates:

| Document Type | Template | Use When |
|--------------|----------|----------|
| **Investigation** | `RESEARCH_INVESTIGATION_TEMPLATE.md` | Exploring unknowns, gathering initial information, understanding problem space |
| **Comparison** | `RESEARCH_COMPARISON_TEMPLATE.md` | Evaluating options, comparing alternatives, making technology decisions |
| **Specification** | `RESEARCH_SPECIFICATION_TEMPLATE.md` | Defining requirements, documenting design decisions, creating implementation contracts |
| **Analysis** | `RESEARCH_ANALYSIS_TEMPLATE.md` | Deep-dive on specific topics, performance analysis, root cause analysis |

### Research Phase Directory Structure

```
001_RESEARCH/
├── PHASE_GOAL.md                           # Research phase objectives
├── topic_1/
│   ├── investigation_current_system.md     # Investigation document
│   ├── comparison_database_options.md      # Comparison document
│   └── analysis_performance.md             # Analysis document
├── topic_2/
│   └── specification_api_design.md         # Specification document
└── results/
    └── research_summary.md                 # Synthesized findings
```

### When to Use Each Document Type

**Investigation** - Start here when facing unknowns:

- "How does the current system work?"
- "What are the existing patterns in the codebase?"
- "What technologies are available?"

**Comparison** - Use when choosing between options:

- "Which database should we use?"
- "Should we build or buy?"
- "Which framework best fits our needs?"

**Specification** - Use to define what will be built:

- "What are the requirements for this feature?"
- "How should this API be designed?"
- "What are the acceptance criteria?"

**Analysis** - Use for deep technical understanding:

- "Why is the system slow?"
- "What is causing these errors?"
- "How secure is the current implementation?"

### Research Phase CLI Commands

```bash
# Create a research phase
fest create phase --name "RESEARCH" --type research

# Create research documents
fest research create --type investigation --title "API Authentication Options"
fest research create --type comparison --title "Database Selection"

# Generate research summary
fest research summary --phase 001_RESEARCH

# Link research to implementation phases
fest research link api-auth.md --phase 002_IMPLEMENT
```

### Research Phase Patterns

#### Discovery-First Pattern

```
001_RESEARCH/
├── investigation_problem_space.md
├── investigation_existing_solutions.md
├── comparison_approaches.md
└── specification_selected_approach.md
```

**Use when:** Starting a project with significant unknowns.

#### Technology Selection Pattern

```
001_EVALUATE_OPTIONS/
├── investigation_requirements.md
├── comparison_option_a_vs_b.md
├── comparison_option_b_vs_c.md
└── analysis_proof_of_concept.md
```

**Use when:** Making critical technology decisions.

#### Root Cause Pattern

```
001_INVESTIGATE/
├── investigation_symptoms.md
├── analysis_system_behavior.md
├── analysis_code_review.md
└── specification_fix_approach.md
```

**Use when:** Diagnosing and fixing complex issues.

### Research Document Quality Gates

Research documents should meet these criteria before being considered complete:

- [ ] All research questions answered or documented as unresolved
- [ ] Evidence provided for all findings
- [ ] Recommendations are actionable and specific
- [ ] Document reviewed by stakeholder
- [ ] Findings integrated into festival planning

## Specialized Domain Patterns

### Data/Analytics Project

**Use when:** Building data pipelines or analytics systems.

```
001_DATA_DISCOVERY → 002_PIPELINE_DESIGN → 003_EXTRACT_TRANSFORM →
004_LOAD_VALIDATE → 005_ANALYZE_REPORT → 006_PRODUCTIONIZE
```

**Data-centric:** Each phase focused on data lifecycle.
**Quality gates:** Validation at every stage.

### Infrastructure/DevOps

**Use when:** Building or migrating infrastructure.

```
001_ASSESS_CURRENT → 002_DESIGN_TARGET → 003_PREPARE_MIGRATION →
004_MIGRATE_SERVICES → 005_VALIDATE_PERFORMANCE → 006_CLEANUP_OLD
```

**Risk management:** Heavy focus on rollback capabilities.
**Incremental:** Migrate services in stages.

### Mobile App Development

**Use when:** Building native or cross-platform mobile apps.

```
001_PLAN → 
002_IMPLEMENT_CORE → 003_IMPLEMENT_FEATURES → 004_TEST_DEVICES →
005_APP_STORE_PREP → 006_RELEASE
```

**Platform-specific:** Considerations for app store processes.
**Testing emphasis:** Device and OS variation testing.

### API Development

**Use when:** Building services for other developers.

```
001_PLAN → 002_IMPLEMENT_CORE →
003_IMPLEMENT_FEATURES → 004_DOCUMENTATION → 005_DEVELOPER_TESTING
```

**Developer-focused:** API design and documentation drive implementation.
**Developer experience:** Documentation and testing are first-class concerns.

**Note:** For complex API systems, consider the [Interface Planning Extension](../extensions/interface-planning/) to formalize contract definition.

## Hybrid and Custom Patterns

### Legacy Migration

**Use when:** Moving from old systems to new ones.

```
001_ASSESS_LEGACY → 002_PLAN_MIGRATION → 003_BUILD_BRIDGE →
004_MIGRATE_DATA → 005_MIGRATE_FEATURES → 006_DEPRECATE_OLD
```

**Risk mitigation:** Parallel systems during transition.
**Data integrity:** Special focus on data migration validation.

### Compliance/Security Project

**Use when:** Meeting regulatory or security requirements.

```
001_AUDIT_CURRENT → 002_GAP_ANALYSIS → 003_REMEDIATION_PLAN →
004_IMPLEMENT_CONTROLS → 005_VALIDATION_TESTING → 006_CERTIFICATION
```

**Audit trail:** Documentation at every step.
**External validation:** Third-party testing and certification.

### Integration Project

**Use when:** Connecting multiple existing systems.

```
001_SYSTEM_ANALYSIS → 002_INTEGRATION_DESIGN → 003_ADAPTER_DEVELOPMENT →
004_DATA_MAPPING → 005_TESTING_INTEGRATION → 006_PRODUCTION_DEPLOYMENT
```

**System boundaries:** Clear understanding of each system's capabilities.
**Error handling:** Robust handling of system failures.

## Phase Naming Conventions

### Technical Naming (Recommended)

- Use descriptive names that indicate the type of work
- `001_RESEARCH`, `002_PROTOTYPE`, `003_IMPLEMENT_CORE`
- Clear to technical teams

### Business Naming

- Use business terminology for stakeholder communication  
- `001_DISCOVERY`, `002_DESIGN`, `003_BUILD`, `004_LAUNCH`
- Easier for non-technical stakeholders

### Domain-Specific Naming

- Adapt to your industry or problem domain
- Healthcare: `001_REQUIREMENTS_ANALYSIS`, `002_CLINICAL_VALIDATION`
- Finance: `001_RISK_ASSESSMENT`, `002_COMPLIANCE_REVIEW`

## Phase Selection Guidelines

### Ask These Questions

**Planning Status:**

- Have requirements been gathered elsewhere?
- Is the architecture already defined?
- Do you have existing documentation?

**Project Complexity:**

- Is this greenfield or brownfield development?
- How many systems are involved?
- What's the technical risk level?

**Team Context:**

- How experienced is the team with this technology?
- Are team members co-located or distributed?
- What's the team size and skill composition?

**Timeline Constraints:**

- Are there fixed deadlines?
- Is this time-boxed or quality-driven?
- What's the risk tolerance for delays?

**Integration Needs:**

- How many external systems are involved?
- Are there data migration requirements?
- What's the deployment complexity?

## Customization Guidelines

### Adapting Existing Patterns

1. **Start with closest match** to your situation
2. **Add phases** for significant work not covered
3. **Remove phases** that don't apply to your project
4. **Rename phases** to match your domain language
5. **Reorder phases** if dependencies require it

### Creating New Patterns

1. **Map your actual work** - what really needs to happen?
2. **Identify dependencies** - what must happen in sequence?
3. **Find parallel opportunities** - what can happen simultaneously?
4. **Add quality gates** - where do you need validation?
5. **Test with your team** - does the structure make sense?

### Red Flags - When Your Phases Need Work

- Single sequence per phase (phases too granular)
- More than 8 sequences per phase (phases too broad)  
- No clear dependencies between phases (arbitrary grouping)
- Phases that could be reordered without impact (not sequential)
- Team confusion about what phase they're in (unclear boundaries)

## Examples of Good vs Bad Phase Design

### ❌ Bad: Arbitrary Grouping

```
001_SETUP
├── 01_create_repo/
├── 02_user_research/
└── 03_database_schema/

002_CODING  
├── 01_frontend/
├── 02_backend/
└── 03_testing/
```

**Problems:** Unrelated work grouped together, no logical progression.

### ✅ Good: Logical Progression

```
001_PLAN
├── 01_user_research/
├── 02_requirements_analysis/
└── 03_system_architecture/

002_IMPLEMENT
├── 01_backend_services/
├── 02_frontend_components/
└── 03_integration_layer/

003_REVIEW_AND_UAT
├── 01_user_acceptance_testing/
├── 02_stakeholder_review/
└── 03_deployment_readiness/
```

**Benefits:** Clear progression, organized implementation, logical groupings with validation.

## Integration with Festival Goals

Each phase should have:

- **Clear objectives** defined in PHASE_GOAL.md
- **Success criteria** that align with festival goals
- **Entry criteria** (what must be done to start)
- **Exit criteria** (what defines completion)
- **Dependencies** on other phases
- **Deliverables** that feed into subsequent phases

## Freeform Phase Types

Some phases use **freeform subdirectory structure** instead of numbered sequences and tasks.

### Planning Phases

Planning phases use freeform structure because thought naturally flows **backward** from goals to requirements:
- "What do we need to build?" → Goal
- "What does that require?" → Dependencies
- "What decisions need to be made?" → Decision points

This backward thinking is **correct for planning** but doesn't fit sequential task structure.

**Planning Phase Structure:**
```
001_PLANNING/
├── PHASE_GOAL.md           # Planning phase objective
├── requirements/           # Topic directory
│   ├── functional.md       # Exploration document
│   └── non-functional.md
├── architecture/           # Topic directory
│   ├── overview.md
│   └── decisions.md
├── decisions/              # Decision records
│   ├── database.md
│   └── auth-strategy.md
├── START_HERE.md           # Entry point for agents
└── PLANNING_SUMMARY.md     # Synthesis document
```

**Common Planning Topic Directories:**
- `requirements/` - Feature requirements, user stories
- `architecture/` - System design, component diagrams
- `decisions/` - Architecture Decision Records (ADRs)
- `specs/` - Technical specifications
- `research/` - Investigation findings

### Research Phases

Research phases also use freeform structure for exploratory work:
- Investigation of approaches
- Prototype development
- Comparative analysis

### Design Phases

Design phases explore solution space before committing to implementation:
- UI/UX exploration
- System design
- API design

### Validation

Freeform phases are validated differently:
- ✓ PHASE_GOAL.md required
- ✓ At least one topic directory recommended
- ✗ No numbered sequence requirement
- ✗ No numbered task requirement

## Summary

Phase patterns provide structure while maintaining flexibility. Choose patterns that match your project's actual needs, not what seems "standard." The best phase structure is one that:

1. **Matches your work reality** - reflects what actually needs to happen
2. **Enables organized execution** - through clear implementation structure
3. **Includes quality gates** - ensures work meets standards  
4. **Makes sense to your team** - uses familiar language and concepts
5. **Adapts as you learn** - can be modified based on new information

Remember: The methodology serves your project, not the other way around.
