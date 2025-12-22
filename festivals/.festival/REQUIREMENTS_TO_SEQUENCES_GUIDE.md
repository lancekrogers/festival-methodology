# Requirements to Sequences Guide

This guide shows how to convert human-provided requirements into logical step progressions that achieve goals. This is a core AI agent skill in Festival Methodology - structuring human requirements into step-based goal achievement progressions.

## Prerequisites

**Before using this guide, ensure you have:**

- Specific requirements or specifications from human
- Clear deliverables or acceptance criteria
- Understanding of what needs to be built (not just what might be needed)
- External planning documents OR completed planning phase

**If you don't have these, STOP. Ask the human for requirements instead of guessing.**

## The Step-Based Conversion Process

**CRITICAL MINDSET**: Think in terms of logical steps that build toward goal achievement, not time estimates or arbitrary task lists.

### Step 1: Analyze Requirements for Goal Achievement

**Good Requirements Look Like:**

```
"Build a user authentication system with:
- Email/password login
- JWT token management (15min access, 7day refresh) 
- Password reset via email
- Role-based access control (user/admin)
- Rate limiting (5 attempts/minute)
- Integration with existing user database"
```

**Bad Requirements Look Like:**

```
"We need some kind of user system"
"Add authentication stuff"  
"Make it secure"
```

### Step 2: Identify Logical Step Progressions

From the good example above, identify steps that build toward the goal:

**Step Progression 1: Core Authentication Foundation**

- Login/logout functionality (enables user sessions)
- Password hashing and validation (secures user credentials)
- Session management (maintains user state)

**Step Progression 2: Token Management System**

- JWT generation and validation (enables stateless auth)
- Token refresh mechanism (maintains long-term sessions)
- Token expiration handling (ensures security)

**Step Progression 3: Security Enhancement Features**

- Rate limiting implementation (prevents abuse)
- Password reset workflow (enables account recovery)
- Role-based permissions (controls access)

**Step Progression 4: System Integration**

- Database schema updates (stores auth data)
- Existing system integration (connects to current architecture)
- API endpoint creation (enables client interaction)

### Step 3: Convert Step Progressions to Sequences

Each logical step progression becomes a sequence with 3-6 implementation steps + completion verification:

```
01_core_authentication/
├── 01_user_model_updates.md
├── 02_password_hashing.md
├── 03_login_logout_endpoints.md
├── 04_session_management.md
├── 05_testing_and_verify.md
├── 06_code_review.md
└── 07_review_results_iterate.md

02_token_management/
├── 01_jwt_generation.md
├── 02_token_validation.md
├── 03_refresh_mechanism.md
├── 04_expiration_handling.md
├── 05_testing_and_verify.md
├── 06_code_review.md
└── 07_review_results_iterate.md

03_security_features/
├── 01_rate_limiting.md
├── 02_password_reset_workflow.md
├── 03_role_based_permissions.md
├── 04_testing_and_verify.md
├── 05_code_review.md
└── 06_review_results_iterate.md
```

## Common Requirements Patterns

### Feature Requirements

```
Human: "Add shopping cart functionality with add/remove items, 
quantity updates, persistent storage, and checkout integration"

Sequences:
01_cart_operations/
├── 01_add_remove_items.md
├── 02_quantity_management.md
├── 03_cart_persistence.md
├── 04_testing_and_verify.md

02_checkout_integration/
├── 01_cart_checkout_flow.md
├── 02_integration_testing.md
├── 03_testing_and_verify.md
```

### API Requirements

```
Human: "Create REST API for blog posts with CRUD operations,
search functionality, category filtering, and pagination"

Sequences:
01_blog_api_core/
├── 01_post_model_schema.md
├── 02_crud_endpoints.md
├── 03_validation_logic.md
├── 04_testing_and_verify.md

02_blog_api_features/
├── 01_search_implementation.md
├── 02_category_filtering.md
├── 03_pagination_logic.md
├── 04_testing_and_verify.md
```

### Database Requirements

```
Human: "Update database schema for multi-tenant support with
tenant isolation, data migration, and admin controls"

Sequences:
01_schema_updates/
├── 01_tenant_model.md
├── 02_existing_table_updates.md
├── 03_foreign_key_constraints.md
├── 04_testing_and_verify.md

02_data_migration/
├── 01_migration_scripts.md
├── 02_data_validation.md
├── 03_rollback_procedures.md
├── 04_testing_and_verify.md

03_tenant_isolation/
├── 01_query_filtering.md
├── 02_admin_controls.md
├── 03_access_validation.md
├── 04_testing_and_verify.md
```

## Quality Principles

### Each Task Should

- Address a specific part of the requirements
- Have clear acceptance criteria
- Produce testable deliverables
- Include specific commands or steps
- Define success/failure conditions

### Each Sequence Should

- Implement a cohesive part of the requirements
- Have 3-6 implementation tasks (plus quality gates)
- Enable other sequences to work in parallel
- Have clear dependencies and completion criteria

## Common Mistakes to Avoid

### ❌ Creating Sequences Without Requirements

```
BAD: Human says "We might need user management"
AI creates: 01_user_management/ with assumed functionality
```

### ❌ Overly Generic Sequences

```
BAD: Requirements specify "JWT with 15min/7day tokens"
AI creates: 01_authentication.md (too vague)

GOOD: 
01_jwt_implementation/
├── 01_access_tokens_15min.md
├── 02_refresh_tokens_7day.md
├── 03_token_validation.md
```

### ❌ Mixing Unrelated Requirements

```
BAD: Human specifies user auth + shopping cart
AI creates: 01_user_features/ (mixes unrelated functionality)

GOOD: Separate sequences for auth and cart
```

### ❌ Skipping Quality Gates

```
BAD: Only implementation tasks, no testing/review

GOOD: Every sequence ends with:
- XX_testing_and_verify.md
- XX_code_review.md  
- XX_review_results_iterate.md
```

## Validation Checklist

Before creating sequences, verify:

- [ ] **Requirements Source**: Clear requirements from human or completed planning
- [ ] **Specificity**: Each task addresses specific functionality, not assumptions
- [ ] **Logical Grouping**: Related tasks grouped into cohesive sequences
- [ ] **Quality Gates**: Every sequence has testing/review/iteration tasks
- [ ] **Testability**: Each task produces verifiable deliverables
- [ ] **Dependencies**: Clear understanding of what must complete first
- [ ] **Completeness**: All requirements addressed, nothing assumed

## When Requirements Are Unclear

If requirements are vague or incomplete:

**Don't guess - Ask:**

```
❌ "I'll create basic user management and you can tell me what to change"

✅ "These requirements need clarification:
   - What authentication methods do you want?
   - What user data fields are needed?
   - What permissions/roles should exist?
   - How should password reset work?
   
   Can you provide more specific requirements for user management?"
```

## Working with External Planning Documents

When human provides planning documents:

1. **Read thoroughly** - Understand the full specification
2. **Identify deliverables** - What concrete things need to be built?
3. **Group by implementation** - What can be built together?
4. **Ask for clarification** - What's ambiguous or missing?
5. **Confirm structure** - Present sequence plan for approval

## Integration with Festival Structure

### Phase-Level Planning

```
Human provides: "Build e-commerce platform with user accounts, 
product catalog, shopping cart, and payment processing"

Phase Structure:
001_DEFINE_INTERFACES → 002_IMPLEMENT_USERS → 003_IMPLEMENT_CATALOG → 
004_IMPLEMENT_CART → 005_IMPLEMENT_PAYMENTS → 006_INTEGRATION_TESTING
```

### Sequence Dependencies

```
Dependencies from requirements:
- User authentication must complete before cart (user sessions)
- Product catalog must complete before cart (product data)
- Cart must complete before payments (cart data)

Parallel Opportunities:
- User authentication and product catalog can develop in parallel
- Testing sequences can run in parallel with next implementation
```

## Collaboration Patterns

### Iterative Refinement

```
1. Human provides initial requirements
2. AI structures into sequences  
3. Human reviews and refines
4. AI adjusts sequences
5. Begin implementation
6. Add more sequences as requirements evolve
```

### Just-in-Time Sequencing

```
Don't create all sequences upfront:
1. Create first sequence from clear requirements
2. Get human approval and begin work
3. Create next sequence when ready
4. Adapt based on learnings from previous sequences
```

## Success Metrics

**Good Requirements-to-Sequences Conversion:**

- Human recognizes their requirements in the sequences
- Each task is actionable and specific
- Dependencies are clear and logical
- Quality gates ensure completeness
- Parallel work opportunities identified
- No assumptions about unstated requirements

**Poor Conversion Signs:**

- Human says "That's not what I meant"
- Tasks are vague or assumptive
- Sequences don't match stated requirements
- Quality gates missing or generic
- Dependencies unclear or circular

## Summary

Converting requirements to sequences is about **faithful translation**, not creative interpretation. Your job is to structure what the human has specified, not to fill in what they haven't. When in doubt, ask for clarification rather than making assumptions.

The goal is sequences that make the human think: "Yes, that's exactly what I need built" not "I guess that might work."
