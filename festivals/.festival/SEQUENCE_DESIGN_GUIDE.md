# Sequence Design Guide

This guide helps you create effective sequences that group related tasks into logical units of work.

**CRITICAL**: Sequences are primarily for IMPLEMENTATION phases where AI agents need structured work. Planning phases often just contain documents and don't need sequences unless deep planning is required.

## When to Create Sequences

**FUNDAMENTAL PRINCIPLE**: Sequences are created FROM requirements, not TO discover requirements.

### ✅ Create Sequences When

- Human has provided specific requirements or specifications
- Planning phase has been completed with clear deliverables
- External planning documents define what needs to be built
- Human explicitly requests implementation of specific functionality
- You have concrete user stories or technical specifications to implement

### ❌ NEVER Create Sequences When

- No requirements have been provided
- Planning phase hasn't been completed or provided deliverables
- You are guessing what might need to be implemented
- Making assumptions about user needs
- Requirements are vague or undefined

### Requirements-Driven Workflow

```
Requirements/Specifications → Structure into Sequences → Create Tasks → Execute
```

**Not:**

```
❌ Guess what's needed → Create sequences → Hope they're right
```

## What Makes a Good Sequence

### Size Guidelines

- **3-6 tasks** is the ideal range
- **2 tasks** minimum (plus quality gates)
- **8 tasks** maximum before breaking into multiple sequences

### Content Guidelines

**Good Sequence Criteria:**

- Tasks build on each other logically
- Tasks share common setup, dependencies, or domain knowledge
- Work forms a cohesive unit (e.g., "user authentication", "payment processing")
- Can be assigned to one person/agent for focused work
- Has clear entry criteria (what must be done before starting)
- Has clear completion criteria (what defines "done")

## Sequence Anti-Patterns to Avoid

### ❌ Single Task Per Sequence

```
BAD:
01_user_model/
└── 01_create_user_model.md

02_password_hashing/
└── 01_add_password_hashing.md

03_login_endpoint/
└── 01_implement_login.md
```

**Fix:** Combine related tasks into logical sequences

```
GOOD:
01_user_authentication/
├── 01_create_user_model.md
├── 02_add_password_hashing.md
├── 03_implement_login_endpoint.md
├── 04_testing_and_verify.md
├── 05_code_review.md
└── 06_review_results_iterate.md
```

### ❌ Arbitrarily Grouped Tasks

```
BAD:
01_mixed_work/
├── 01_setup_database.md
├── 02_create_react_component.md
├── 03_configure_nginx.md
└── 04_write_api_docs.md
```

**Fix:** Group by logical domain/system

```
GOOD:
01_database_setup/
├── 01_setup_postgresql.md
├── 02_create_migrations.md
├── 03_testing_and_verify.md

02_frontend_auth/
├── 01_create_login_component.md
├── 02_add_form_validation.md
├── 03_testing_and_verify.md
```

### ❌ Overly Large Sequences

```
BAD:
01_complete_user_system/
├── 01_database_schema.md
├── 02_user_model.md
├── 03_authentication.md
├── 04_authorization.md
├── 05_profile_management.md
├── 06_password_reset.md
├── 07_email_verification.md
├── 08_user_settings.md
├── 09_admin_controls.md
├── 10_audit_logging.md
├── 11_testing_and_verify.md
```

**Fix:** Break into focused sequences

```
GOOD:
01_user_foundation/
├── 01_database_schema.md
├── 02_user_model.md
├── 03_testing_and_verify.md

02_authentication/
├── 01_login_system.md
├── 02_password_hashing.md
├── 03_jwt_tokens.md
├── 04_testing_and_verify.md

03_user_management/
├── 01_profile_editing.md
├── 02_password_reset.md
├── 03_email_verification.md
├── 04_testing_and_verify.md
```

## Standard Quality Gates

**EVERY implementation sequence MUST end with these three tasks:**

```
XX_testing_and_verify.md      ← Verify functionality works as specified
XX_code_review.md             ← Review code quality and standards
XX_review_results_iterate.md  ← Address findings and iterate if needed
```

### Quality Gate Templates

**Testing and Verify Task:**

```markdown
# Task: 05_testing_and_verify.md

## Objective
Verify that all sequence deliverables work as specified

## Requirements
- [ ] Unit tests pass for all new code
- [ ] Integration tests cover main workflows
- [ ] Manual testing confirms user stories
- [ ] Performance meets requirements
- [ ] Security checks pass
```

**Code Review Task:**

```markdown
# Task: 06_code_review.md

## Objective
Review code quality, standards compliance, and architecture

## Requirements
- [ ] Code follows project style guidelines
- [ ] Architecture aligns with COMMON_INTERFACES.md
- [ ] Documentation is complete and accurate
- [ ] No security vulnerabilities identified
- [ ] Performance considerations addressed
```

**Review Results and Iteration Task:**

```markdown
# Task: 07_review_results_iterate.md

## Objective
Address review findings and iterate until acceptance criteria met

## Requirements
- [ ] All code review findings resolved
- [ ] Failed tests fixed or requirements clarified
- [ ] Performance issues addressed
- [ ] Security concerns resolved
- [ ] Stakeholder acceptance obtained
```

## Common Sequence Patterns

### Database Sequence Pattern

```
01_database_setup/
├── 01_schema_design.md
├── 02_create_migrations.md
├── 03_seed_data.md
├── 04_testing_and_verify.md
├── 05_code_review.md
└── 06_review_results_iterate.md
```

### API Development Pattern

```
01_user_api/
├── 01_endpoint_design.md
├── 02_request_validation.md
├── 03_business_logic.md
├── 04_response_formatting.md
├── 05_error_handling.md
├── 06_testing_and_verify.md
├── 07_code_review.md
└── 08_review_results_iterate.md
```

### Frontend Component Pattern

```
01_login_component/
├── 01_component_structure.md
├── 02_form_validation.md
├── 03_state_management.md
├── 04_styling.md
├── 05_testing_and_verify.md
├── 06_code_review.md
└── 07_review_results_iterate.md
```

### DevOps/Infrastructure Pattern

```
01_deployment_setup/
├── 01_environment_config.md
├── 02_ci_cd_pipeline.md
├── 03_monitoring_setup.md
├── 04_testing_and_verify.md
├── 05_code_review.md
└── 06_review_results_iterate.md
```

## Sequence Planning Checklist

Before creating a sequence, verify:

- [ ] **Logical Cohesion**: Do all tasks relate to the same feature/system?
- [ ] **Appropriate Size**: 3-6 implementation tasks + 3 quality gates?
- [ ] **Clear Dependencies**: What must be done before this sequence?
- [ ] **Completion Criteria**: How will you know this sequence is done?
- [ ] **Quality Gates**: Are testing/review/iteration tasks included?
- [ ] **Parallel Opportunities**: Can tasks within sequence run simultaneously?

## Sequence vs Single Task Decision

### Create a Sequence When

- You have multiple related subtasks
- Tasks share common setup or knowledge domain
- Work benefits from being done by same person/agent
- Tasks build on each other
- Quality gates apply to the group of work

### Create a Single Task When

- Work is atomic and self-contained
- No natural subtasks emerge
- Can be completed in one focused session
- Doesn't benefit from breakdown
- Quality verification is simple

## Integration with Festival Structure

### Within Phases

### Sequences by Phase Type

**Planning/Research Phases (Often Unstructured):**

- May just contain documents and findings
- Add sequences only if deep planning requires structure
- Example: Just README.md with requirements, no sequences needed

**Implementation Phases (Must Be Structured):**

- ALWAYS have sequences and tasks for AI execution
- Examples for 002_IMPLEMENT_CORE:
  - 01_backend_foundation/
  - 02_database_layer/
  - 03_api_endpoints/
- Examples for 003_IMPLEMENT_FEATURES:
  - 01_user_management/
  - 02_payment_processing/
  - 03_notification_system/

**Validation Phases:**

- 01_user_acceptance_testing/
- 02_performance_validation/
- 03_deployment_preparation/

### Cross-Sequence Dependencies

Use clear numbering to indicate dependencies:

- `01_foundation/` must complete before `02_features/`
- Tasks with same numbers can run in parallel
- Document dependencies in sequence README files

## Tools and Automation

### Progress Tracking

Update your FESTIVAL_TODO.md as sequences complete:

```markdown
## Phase 003_IMPLEMENT [██████░░░░] 60%

### 01_backend_core [✅] Complete
- [x] 01_database_models.md
- [x] 02_api_endpoints.md
- [x] 03_authentication.md
- [x] 04_testing_and_verify.md
- [x] 05_code_review.md
- [x] 06_review_results_iterate.md

### 02_frontend_components [🚧] In Progress
- [x] 01_login_component.md
- [x] 02_dashboard_layout.md
- [ ] 03_user_profile.md
- [ ] 04_testing_and_verify.md
- [ ] 05_code_review.md
- [ ] 06_review_results_iterate.md
```

### Quality Metrics

Track sequence success:

- Average tasks per sequence (target: 3-6)
- Quality gate completion rate (target: 100%)
- Rework rate after review (target: <20%)
- Sequence completion time consistency

## Troubleshooting

### "My sequences feel arbitrary"

- Focus on user stories or system components
- Group tasks that share the same knowledge domain
- Ensure tasks build on each other logically

### "Too many single-task sequences"

- Look for related work that can be combined
- Consider if work is actually a single larger task
- Group by technical domain (database, API, UI)

### "Sequences are too large"

- Break by feature boundaries
- Separate setup/core/advanced functionality
- Use dependency relationships to split

### "Quality gates feel repetitive"

- Customize for each sequence's domain
- Include sequence-specific validation
- Focus on the unique risks of that work

## Summary

Good sequence design is about creating logical, manageable units of work that include proper quality controls. Avoid the temptation to create single-task sequences, and always include the standard quality gates. Remember: sequences should represent focused work that one person/agent can own end-to-end.
