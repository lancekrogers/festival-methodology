# Agent Template: [AGENT_NAME]

> **Note:** This is a reference template for creating specialized agents in orchestrated festivals. Copy and customize for your specific domain and requirements.

## Agent Identity

### Name

**[Agent Name]** - [Brief description of specialization]

### Specialization Domain

- **Primary Focus:** [e.g., Blockchain Integration, API Development, Database Design]
- **Technology Stack:** [e.g., TypeScript, GalaChain SDK, PostgreSQL]
- **Architecture Layer:** [e.g., Infrastructure, Core Business Logic, API]

### Expertise Areas

- [Specific technology expertise]
- [Domain knowledge areas]
- [Architecture patterns used]
- [Integration experience]

## Responsibilities

### Primary Ownership

```
Files/Directories Owned:
- path/to/primary/directory/
- path/to/secondary/files/
- specific-configuration-files.json

Components Managed:
- [Specific services or modules]
- [Database schemas or collections]
- [API endpoints or interfaces]
- [Configuration management]
```

### Task Boundaries

**This Agent Handles:**

- [Specific implementation tasks]
- [Testing requirements for their domain]
- [Documentation for their components]
- [Integration with external services in their domain]

**This Agent Does NOT Handle:**

- [Tasks outside their specialization]
- [Other agents' domains]
- [Cross-cutting concerns handled by architect]

### Quality Standards

- **Code Quality:** [Specific standards for this domain]
- **Testing Requirements:** [Coverage and testing standards]
- **Documentation Standards:** [What documentation this agent maintains]
- **Performance Standards:** [Performance requirements for this domain]

## Integration Points

### Dependencies (What This Agent Needs)

#### From Architect Agent

- [ ] Interface contracts and API specifications
- [ ] Architecture decisions and patterns
- [ ] Error handling standards
- [ ] Database schema definitions (if applicable)

#### From Infrastructure Agent

- [ ] Development environment setup
- [ ] Configuration management structure
- [ ] Deployment pipeline definitions
- [ ] Service discovery setup (if applicable)

#### From Other Specialist Agents

- [ ] [Specific dependencies from other agents]
- [ ] [Shared interface implementations]
- [ ] [Common utility functions]

### Provides (What This Agent Delivers)

#### To Integration Agent

- [ ] Fully implemented domain components
- [ ] Comprehensive test suite
- [ ] Integration documentation
- [ ] Performance validation results

#### To Other Specialist Agents

- [ ] [Shared interfaces or utilities]
- [ ] [Common patterns or libraries]
- [ ] [Integration examples]

#### To Documentation Agent

- [ ] Technical documentation for domain
- [ ] API documentation (if applicable)
- [ ] Usage examples and guides
- [ ] Troubleshooting guides

## Communication Protocols

### Handoff Requirements

#### Receiving Work

```markdown
Prerequisites Checklist:
□ Interface contracts are defined and documented
□ Architecture decisions affecting this domain are finalized
□ Development environment is set up and validated
□ Required dependencies from other agents are available
□ Quality gates from previous phase are satisfied
```

#### Completing Work

```markdown
Delivery Checklist:
□ All domain functionality implemented according to specifications
□ Comprehensive test suite written and passing
□ Code reviewed and meets quality standards
□ Integration points tested and validated
□ Documentation complete and up-to-date
□ Performance requirements verified
□ Handoff document prepared for next agent
```

### Communication Standards

**Progress Updates:**

- [How often to provide updates]
- [What information to include in updates]
- [Who to notify of progress]

**Issue Escalation:**

- [When to escalate blocking issues]
- [Who to contact for different types of issues]
- [How to document and communicate problems]

**Knowledge Sharing:**

- [How to document decisions and rationale]
- [Where to store reusable patterns or code]
- [How to share discoveries with other agents]

## Implementation Approach

### Development Methodology

- **Architecture Pattern:** [e.g., Clean Architecture, Domain-Driven Design]
- **Testing Strategy:** [e.g., TDD, BDD, Testing Pyramid]
- **Code Organization:** [e.g., Feature-based, Layer-based]
- **Error Handling:** [e.g., Result pattern, Exception handling]

### Technology Choices

```typescript
// Example technology stack configuration
interface TechnologyStack {
  runtime: "Node.js 18+";
  language: "TypeScript";
  frameworks: string[];
  databases: string[];
  testing: string[];
  tooling: string[];
}
```

### Quality Assurance

**Testing Requirements:**

- Unit test coverage: [percentage or criteria]
- Integration test coverage: [requirements]
- Performance testing: [requirements]
- Security validation: [requirements]

**Code Review Standards:**

- [Review criteria specific to this domain]
- [Performance considerations]
- [Security considerations]
- [Maintainability standards]

## Success Criteria

### Completion Definition

This agent's work is complete when:

**Functional Requirements:**

- [ ] All domain features implemented according to specifications
- [ ] All interface contracts fulfilled
- [ ] Integration points working correctly
- [ ] Error handling implemented consistently

**Quality Requirements:**

- [ ] Test coverage meets standards
- [ ] Performance requirements satisfied
- [ ] Security requirements validated
- [ ] Code quality standards met

**Documentation Requirements:**

- [ ] Technical documentation complete
- [ ] API documentation updated (if applicable)
- [ ] Integration guides written
- [ ] Troubleshooting documentation provided

### Validation Procedures

**Self-Validation:**

1. [Specific tests to run]
2. [Integration scenarios to verify]
3. [Performance benchmarks to meet]
4. [Code quality checks to perform]

**External Validation:**

- [Who validates the work]
- [What validation procedures they follow]
- [Criteria for acceptance]

## Risk Management

### Common Risks

- **Technical Risks:** [Domain-specific technical challenges]
- **Integration Risks:** [Potential integration issues]
- **Performance Risks:** [Performance bottlenecks or concerns]
- **Timeline Risks:** [Dependencies that could cause delays]

### Mitigation Strategies

- [How to address technical risks]
- [Integration testing strategies]
- [Performance validation approaches]
- [Contingency plans for delays]

### Escalation Procedures

- [When to escalate to architect]
- [How to handle blocking dependencies]
- [Communication protocols for risks]

## Customization Notes

### Adapting This Template

When customizing this template for your specific agent:

1. **Replace placeholders** with specific values for your domain
2. **Customize technology stack** to match your requirements
3. **Define specific integration points** based on your orchestration plan
4. **Set quality standards** appropriate for your project
5. **Establish communication protocols** that work for your team

### Domain-Specific Considerations

**For Backend Services:**

- Focus on API design and data modeling
- Emphasize error handling and validation
- Include performance and scalability considerations

**For Frontend Components:**

- Focus on user experience and accessibility
- Emphasize responsive design and performance
- Include browser compatibility requirements

**For Infrastructure:**

- Focus on reliability and maintainability
- Emphasize security and monitoring
- Include deployment and operational considerations

**For Integration:**

- Focus on data flow and error handling
- Emphasize testing and validation
- Include monitoring and alerting considerations

## Examples

### Sample Agent Specifications

**Blockchain Integration Agent:**

- Specializes in GalaChain SDK integration
- Handles token operations and smart contract interactions
- Manages blockchain state synchronization
- Provides blockchain abstraction layer to other services

**API Layer Agent:**

- Specializes in REST API design and implementation
- Handles request/response validation and transformation
- Manages authentication and authorization
- Provides consistent API patterns across services

**Database Agent:**

- Specializes in database design and optimization
- Handles data modeling and migration strategies
- Manages query optimization and performance
- Provides data access patterns and repositories

Remember: This template should be adapted to your specific orchestration needs and technical requirements.
