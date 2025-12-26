# Festival Methodology Examples Index

Learn from concrete implementations and proven patterns.

## Available Examples

### Task Examples

**File:** `TASK_EXAMPLES.md`

Contains 15+ real-world task examples across different domains:

1. **Database Tasks**
   - Creating tables and migrations
   - Setting up connections
   - Data seeding and fixtures

2. **API Tasks**
   - Endpoint implementation
   - Authentication setup
   - Rate limiting configuration

3. **Frontend Tasks**
   - Component creation
   - State management setup
   - Routing configuration

4. **DevOps Tasks**
   - Docker configuration
   - CI/CD pipeline setup
   - Environment configuration

5. **Testing Tasks**
   - Unit test implementation
   - Integration test setup
   - E2E test configuration

### Festival TODO Example

**File:** `FESTIVAL_TODO_EXAMPLE.md`

Complete example of festival progress tracking showing:

- All progress states (Not Started, In Progress, Complete, Blocked)
- Percentage calculations
- Visual progress indicators
- Dependency tracking
- Issue documentation

## Common Festival Patterns

### Pattern 1: Web Application Festival

```
001_PLAN
├── 01_requirements_analysis
├── 02_architecture_design
└── 03_technology_selection

002_DEFINE_INTERFACES
├── 01_api_design
├── 02_data_models
├── 03_frontend_contracts
└── 04_integration_points

003_IMPLEMENT
├── 01_database_setup
├── 02_backend_api
├── 03_frontend_app
├── 04_authentication
└── 05_integration

004_REVIEW_AND_UAT
├── 01_testing
├── 02_performance_optimization
├── 03_security_audit
└── 04_deployment
```

### Pattern 2: Microservices Festival

```
001_PLAN
├── 01_domain_modeling
├── 02_service_boundaries
└── 03_communication_patterns

002_DEFINE_INTERFACES
├── 01_service_contracts
├── 02_event_schemas
├── 03_api_gateway_design
└── 04_shared_libraries

003_IMPLEMENT
├── 01_shared_infrastructure
├── 02_service_a
├── 03_service_b
├── 04_service_c
└── 05_api_gateway

004_REVIEW_AND_UAT
├── 01_service_testing
├── 02_integration_testing
├── 03_load_testing
└── 04_orchestration
```

### Pattern 3: Data Pipeline Festival

```
001_PLAN
├── 01_data_requirements
├── 02_pipeline_architecture
└── 03_quality_requirements

002_DEFINE_INTERFACES
├── 01_data_schemas
├── 02_transformation_rules
├── 03_api_contracts
└── 04_monitoring_metrics

003_IMPLEMENT
├── 01_ingestion_layer
├── 02_transformation_layer
├── 03_storage_layer
├── 04_serving_layer
└── 05_monitoring

004_REVIEW_AND_UAT
├── 01_data_validation
├── 02_performance_testing
├── 03_quality_assurance
└── 04_documentation
```

### Pattern 4: Mobile App Festival

```
001_PLAN
├── 01_user_research
├── 02_platform_requirements
└── 03_backend_requirements

002_DEFINE_INTERFACES
├── 01_api_design
├── 02_data_models
├── 03_ui_components
└── 04_navigation_flow

003_IMPLEMENT
├── 01_backend_setup
├── 02_app_foundation
├── 03_core_features
├── 04_offline_support
└── 05_push_notifications

004_REVIEW_AND_UAT
├── 01_device_testing
├── 02_performance_optimization
├── 03_app_store_prep
└── 04_beta_testing
```

## Task Numbering Patterns

### Sequential Tasks

```
01_setup_database.md
02_create_models.md
03_add_migrations.md
04_seed_data.md
```

### Parallel Tasks (same number)

```
01_frontend_setup.md
01_backend_setup.md
01_database_setup.md
02_integrate_all.md
```

### Mixed Pattern

```
01_planning.md
02_design_api.md
02_design_ui.md
03_implement_api.md
03_implement_ui.md
04_integration.md
```

## Quality Gate Examples

### Standard Quality Sequence

Every sequence should end with:

```
XX_testing_and_verify.md
XX_code_review.md
XX_review_results_iterate.md
```

### Extended Quality Sequence

For critical features:

```
XX_unit_testing.md
XX_integration_testing.md
XX_security_review.md
XX_performance_testing.md
XX_code_review.md
XX_stakeholder_review.md
XX_review_results_iterate.md
```

## Interface Definition Examples

### REST API Interface

```markdown
## POST /api/users/register

### Request
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "John Doe"
}
```

### Response

```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Errors

- 400: Invalid email format
- 409: Email already exists
- 422: Password too weak

```

### GraphQL Interface
```graphql
type User {
  id: ID!
  email: String!
  name: String!
  posts: [Post!]!
}

type Query {
  user(id: ID!): User
  users(limit: Int = 10): [User!]!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
}
```

### Event Interface

```yaml
event: user.registered
version: 1.0.0
payload:
  user_id: string
  email: string
  timestamp: datetime
  source: string
metadata:
  correlation_id: string
  causation_id: string
```

## Common Task Structures

### Database Task

```markdown
# Task: 01_create_user_table

## Objective
Create PostgreSQL user table with authentication fields

## Requirements
- [ ] Create migration file
- [ ] Define table schema
- [ ] Add indexes
- [ ] Create model

## Implementation
1. Generate migration
2. Define schema
3. Run migration
4. Test model
```

### API Task

```markdown
# Task: 02_implement_login_endpoint

## Objective
Implement POST /api/auth/login endpoint with JWT

## Requirements
- [ ] Validate credentials
- [ ] Generate JWT token
- [ ] Return user data
- [ ] Handle errors

## Implementation
1. Create route handler
2. Add validation
3. Implement JWT logic
4. Add tests
```

### Frontend Task

```markdown
# Task: 03_create_login_form

## Objective
Create responsive login form component

## Requirements
- [ ] Email/password fields
- [ ] Validation
- [ ] Error display
- [ ] Loading state

## Implementation
1. Create component
2. Add form handling
3. Connect to API
4. Add styling
```

## Anti-Pattern Examples

### ❌ BAD: Vague Task

```markdown
# Task: implement_authentication
## Objective
Add authentication to the system
```

### ✅ GOOD: Specific Task

```markdown
# Task: 01_implement_jwt_authentication
## Objective
Implement JWT-based authentication with refresh tokens for REST API
```

### ❌ BAD: Missing Dependencies

```markdown
Tasks:
- 01_create_frontend
- 02_create_backend
- 03_connect_them
```

### ✅ GOOD: Clear Dependencies

```markdown
Tasks:
- 01_define_api_contract
- 02_implement_backend (depends on 01)
- 02_implement_frontend (depends on 01)
- 03_integration_testing (depends on 02s)
```

## Real-World Adaptations

### Startup/MVP Festival

- Lean Phase 001 (1-2 sequences)
- Focus on Phase 002 interfaces
- Rapid Phase 003 with basics
- Minimal Phase 004

### Enterprise Festival

- Extensive Phase 001 (compliance, security)
- Detailed Phase 002 (governance review)
- Parallel Phase 003 (multiple teams)
- Comprehensive Phase 004 (audits)

### Open Source Festival

- Community-driven Phase 001
- Public Phase 002 (RFC process)
- Contributor Phase 003
- Community Phase 004

## Learning Resources

### From Examples

1. Start with `TASK_EXAMPLES.md` for task writing
2. Review `FESTIVAL_TODO_EXAMPLE.md` for tracking
3. Study patterns above for structure
4. Adapt to your specific needs

### Best Practices Derived

- Number tasks to show dependencies
- Define interfaces completely before coding
- Include quality gates in every sequence
- Make tasks concrete and testable
- Track progress visually

## Contributing Examples

Have a great festival pattern? Share it!

1. Document your festival structure
2. Include lessons learned
3. Submit via [CONTRIBUTING.md](../../../CONTRIBUTING.md)
4. Help others learn from your experience

## Summary

These examples demonstrate:

- Proper three-level hierarchy
- Interface-first development
- Quality gate integration
- Parallel work opportunities
- Clear dependency management
- Concrete task definition

Use them as inspiration, not prescription. Every project is unique - adapt the methodology to fit your needs while maintaining core principles.
