# Interface Planning Extension

The Interface Planning extension adds a dedicated phase for defining system contracts and enabling parallel development in multi-system architectures.

## When to Use Interface Planning

### Use This Extension When

- **Multiple services/systems** need to interact with each other
- **External integrations** require well-defined contracts
- **Team parallelization** would benefit from contract-first development
- **API-first development** is required for the project
- **Large systems** where integration complexity is a major risk

### Don't Use When

- Single application with no external integrations
- Simple tools or utilities
- Prototyping or proof-of-concept work
- Small team projects where coordination overhead exceeds benefit

## Phase Integration

The interface planning extension modifies the standard 3-phase pattern:

### Without Extension (Standard)

```
001_PLAN → 002_IMPLEMENT → 003_REVIEW_AND_UAT
```

### With Interface Planning Extension

```
001_PLAN → 002_DEFINE_INTERFACES → 003_IMPLEMENT → 004_REVIEW_AND_UAT
```

## Extension Benefits

### Parallel Development

- Teams can work simultaneously once contracts are defined
- Reduces development time for multi-system projects
- Eliminates integration surprises and rework

### Clear System Boundaries

- Forces explicit definition of system responsibilities
- Reduces coupling between system components
- Enables better testing through contract validation

### Quality Assurance

- Interface contracts serve as tests for system integration
- Prevents breaking changes during development
- Enables automated contract testing

## Phase 002: DEFINE_INTERFACES

### Purpose

Define all system interfaces, communication protocols, and data contracts before implementation begins.

### Key Deliverables

- **COMMON_INTERFACES.md** - Complete interface specification
- **System architecture diagrams** - Component relationships and data flow
- **API contracts** - Endpoint definitions and data schemas
- **Integration protocols** - Communication patterns and error handling
- **Performance requirements** - SLAs and scalability contracts

### Completion Criteria

- All system interfaces documented and reviewed
- Stakeholder sign-off on interface contracts
- Interface status shows FINALIZED (not DRAFT or UNDER_REVIEW)
- Implementation teams confirm contracts are sufficient

### Quality Gates

- Interface review checklist completed
- Contract testing strategy defined
- Performance requirements validated
- Security review passed

## Typical Sequence Structure

### 01_system_architecture/

- Define system components and boundaries
- Map data flow and communication patterns
- Identify external dependencies

### 02_api_contracts/

- Design REST/GraphQL/gRPC endpoints
- Define request/response schemas
- Specify authentication and authorization

### 03_data_contracts/

- Design database schemas
- Define data consistency requirements
- Plan migration strategies

### 04_integration_protocols/

- Choose communication protocols
- Define error handling patterns
- Plan monitoring and alerting

### 05_performance_contracts/

- Set SLA requirements
- Define scalability patterns
- Plan load testing approach

### 06_interface_validation/

- Review all contracts for completeness
- Get stakeholder sign-offs
- Finalize interface documentation

## Agent Behavior Changes

### Planning Agent

When interface planning extension is active:

- Emphasizes interface definition as critical phase
- Creates detailed interface sequences
- Ensures COMMON_INTERFACES.md template is included
- Validates system boundaries and dependencies

### Methodology Manager

When interface planning extension is active:

- Enforces interface finalization before implementation
- Monitors COMMON_INTERFACES.md status
- Blocks Phase 003 until interfaces are FINALIZED
- Validates interface contract adherence during implementation

### Review Agent

When interface planning extension is active:

- Validates interface completeness and consistency
- Checks for missing system interactions
- Ensures performance and security requirements are defined
- Verifies stakeholder sign-off process

## Extension Activation

### Automatic Suggestion

The planning agent suggests this extension when it detects:

- Multiple service components mentioned in requirements
- External API integrations required
- Team size indicates parallel development benefit
- Microservices or distributed architecture patterns

### Example Detection Triggers

```
Requirements mention:
- "microservices architecture"
- "API integration with [external system]"
- "frontend and backend teams"
- "multiple databases"
- "real-time communication"
- "third-party integrations"
```

### Manual Activation

Humans can explicitly request the extension:

```
"Use the interface-planning extension because we have 5 microservices that need to coordinate"
```

## Trade-offs

### Benefits

- **Faster overall development** through parallelization
- **Reduced integration issues** through contract-first design
- **Better system quality** through explicit interface design
- **Clearer team coordination** through defined boundaries

### Costs

- **Additional planning overhead** - Extra phase adds complexity
- **Coordination effort** - Requires stakeholder alignment and sign-offs
- **Change management** - Interface changes require formal process
- **Learning curve** - Teams need to understand contract-first development

## Success Metrics

### Project Success Indicators

- Implementation phase shows minimal integration issues
- Teams work in parallel without blocking dependencies
- Interface changes are rare and well-managed
- System integration testing passes with minimal rework

### Process Success Indicators

- Interface definition phase completes with clear deliverables
- All stakeholders sign off on contracts before implementation
- Implementation teams report sufficient interface definition
- Contract testing catches integration issues early

## Common Pitfalls

### Over-Engineering Interfaces

- **Problem**: Defining interfaces for every possible future need
- **Solution**: Focus on current requirements with extensibility hooks

### Premature Optimization

- **Problem**: Optimizing interface performance before understanding usage
- **Solution**: Design for clarity first, optimize based on actual metrics

### Incomplete Stakeholder Alignment

- **Problem**: Starting implementation before all teams agree on interfaces
- **Solution**: Enforce sign-off process and change control procedures

### Interface Drift

- **Problem**: Implementation deviates from defined interfaces
- **Solution**: Automated contract testing and regular interface validation

## Integration with Base Methodology

### Maintains Step-Based Focus

Interface planning maintains Festival Methodology's core principle of **step-based goal achievement**. The interface phase is about identifying the logical steps needed to enable parallel system development, not about time estimation or rigid scheduling.

### Quality Verification Patterns

All interface planning sequences include the standard quality gates:

- `XX_testing_and_verify.md` - Validate interface definitions
- `XX_code_review.md` - Review contracts for completeness and quality
- `XX_review_results_iterate.md` - Decide whether to iterate or proceed

### Goal-Oriented Planning

Interface planning serves the overall festival goal by enabling efficient implementation. The extension is only valuable when it accelerates goal achievement through better coordination and reduced integration risk.

## Examples of Good Interface Planning Projects

### E-commerce Platform

- Frontend web app + mobile app
- User service + product service + order service
- Payment gateway integration
- Search service integration
- Email notification service

### Healthcare System

- Patient portal frontend
- Electronic health records service
- Appointment scheduling service
- Insurance verification service
- Clinical decision support integration

### Financial Trading Platform

- Web trading interface
- Market data service
- Order execution service
- Risk management service
- Regulatory reporting integration

In each case, the interface planning extension enables multiple teams to work simultaneously while ensuring system components integrate correctly.

---

Remember: **Interface planning is powerful for multi-system projects, but adds complexity that isn't justified for simple single-system applications.** Use this extension when coordination overhead is smaller than the parallel development benefits it enables.
