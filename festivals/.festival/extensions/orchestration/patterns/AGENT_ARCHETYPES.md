# Agent Archetypes for Orchestration

## Overview

This document defines common agent archetypes that appear across different orchestration patterns. Each archetype represents a distinct role with specific expertise, responsibilities, and interaction patterns. Understanding these archetypes helps in planning agent teams and ensuring all necessary capabilities are covered in orchestrated festivals.

## Core Archetype Categories

### Leadership and Coordination

- **Architect/Coordinator**: System design and team coordination
- **Project Manager**: Timeline, resource, and dependency management
- **Technical Lead**: Technical decisions and standards

### Domain Specialists

- **Backend Engineer**: Server-side logic and APIs
- **Frontend Engineer**: User interfaces and client-side logic
- **Database Engineer**: Data modeling and persistence
- **DevOps Engineer**: Infrastructure and deployment

### Quality and Support

- **Quality Assurance**: Testing and validation
- **Documentation Expert**: Technical writing and knowledge management
- **Performance Engineer**: Optimization and scalability
- **Security Specialist**: Security and compliance

## Leadership and Coordination Archetypes

### Architect/Coordinator

#### Role Description

The Architect/Coordinator serves as the central technical authority and coordination hub for orchestrated festivals. They maintain the overall system vision, make key architectural decisions, and ensure consistency across specialized agents.

#### Core Responsibilities

```
System Architecture:
- Define overall system architecture and design principles
- Create and maintain architectural documentation
- Make technology stack decisions
- Ensure architectural consistency across components

Coordination:
- Coordinate between specialist agents
- Facilitate interface definition and contracts
- Resolve technical conflicts and dependencies
- Manage handoffs between agents and phases

Quality Oversight:
- Establish coding standards and best practices
- Review critical technical decisions
- Ensure system coherence and maintainability
- Validate integration points and interfaces
```

#### Key Skills and Expertise

- **Technical Breadth**: Understanding of all domains in the festival
- **System Design**: Ability to design scalable, maintainable systems
- **Communication**: Strong ability to explain technical concepts
- **Decision Making**: Comfort with making architectural trade-offs

#### Interaction Patterns

- **Hub Role**: Central point of coordination in hub-and-spoke patterns
- **Interface Facilitator**: Leads interface definition phases
- **Conflict Resolver**: Final authority on technical disputes
- **Quality Gatekeeper**: Validates major transitions and deliverables

#### Example Usage (Blockchain Verification)

```
Agent: blockchain-verification-architect
- Led Phase 002 interface definition
- Coordinated between 6 specialist agents
- Maintained system coherence across blockchain, testing, and verification domains
- Resolved integration conflicts and dependencies
```

#### Success Metrics

- System architecture clarity and completeness
- Agent coordination effectiveness
- Technical debt and rework minimization
- Overall festival timeline adherence

### Project Manager

#### Role Description

The Project Manager focuses on timeline, resource allocation, dependency management, and risk mitigation. They ensure the orchestration runs smoothly from a process perspective while the Architect handles technical concerns.

#### Core Responsibilities

```
Timeline Management:
- Create and maintain project schedules
- Track milestone progress and deadlines
- Identify and resolve timeline risks
- Coordinate agent availability and workload

Dependency Management:
- Map dependencies between agents and phases
- Track dependency resolution
- Escalate blocking issues
- Maintain dependency documentation

Risk Management:
- Identify potential risks and bottlenecks
- Develop mitigation strategies
- Monitor risk indicators
- Coordinate contingency plans

Communication:
- Facilitate regular status meetings
- Maintain project dashboards and reports
- Handle stakeholder communication
- Document decisions and changes
```

#### Key Skills and Expertise

- **Project Management**: Experience with agile and waterfall methodologies
- **Risk Assessment**: Ability to identify and mitigate project risks
- **Communication**: Strong facilitation and reporting skills
- **Tool Proficiency**: Experience with project management tools

#### Interaction Patterns

- **Process Coordinator**: Manages the orchestration process flow
- **Status Hub**: Central point for status and progress information
- **Escalation Manager**: Handles issues that require management attention
- **Communication Bridge**: Connects technical teams with stakeholders

### Technical Lead

#### Role Description

The Technical Lead provides deep technical expertise in specific domains while also contributing to overall technical direction. They bridge the gap between Architect-level vision and implementation-level details.

#### Core Responsibilities

```
Technical Direction:
- Provide domain-specific technical leadership
- Mentor and guide specialist agents
- Review technical implementations
- Contribute to architectural decisions

Standards and Practices:
- Establish domain-specific coding standards
- Define testing and quality requirements
- Create implementation guidelines
- Ensure best practices adoption

Code Quality:
- Conduct code reviews
- Validate technical designs
- Ensure performance and security standards
- Maintain technical documentation
```

#### Key Skills and Expertise

- **Deep Technical Expertise**: Expert-level knowledge in specific domains
- **Mentoring**: Ability to guide and develop other engineers
- **Code Review**: Strong skills in evaluating code quality
- **Standards Development**: Experience creating technical guidelines

## Domain Specialist Archetypes

### Backend Engineer

#### Role Description

The Backend Engineer specializes in server-side logic, APIs, databases, and system integration. They handle the core business logic and data processing components of the system.

#### Core Responsibilities

```
API Development:
- Design and implement REST/GraphQL APIs
- Handle request/response processing
- Implement authentication and authorization
- Manage API versioning and documentation

Business Logic:
- Implement core business rules and workflows
- Handle data validation and transformation
- Manage state transitions and processes
- Integrate with external services

Data Management:
- Design database schemas and migrations
- Implement data access layers
- Handle caching and performance optimization
- Manage data consistency and integrity

Integration:
- Integrate with third-party services
- Handle message queues and async processing
- Implement monitoring and logging
- Manage service-to-service communication
```

#### Key Skills and Expertise

- **Programming Languages**: Node.js, Python, Java, Go, etc.
- **Databases**: SQL, NoSQL, caching strategies
- **API Design**: REST, GraphQL, microservices
- **Integration**: Message queues, webhooks, third-party APIs

#### Interaction Patterns

- **Service Provider**: Provides APIs and services to other components
- **Data Guardian**: Manages data consistency and business rules
- **Integration Hub**: Connects with external systems and services
- **Performance Owner**: Responsible for backend performance and scalability

### Frontend Engineer

#### Role Description

The Frontend Engineer specializes in user interfaces, client-side logic, and user experience. They handle everything users see and interact with directly.

#### Core Responsibilities

```
User Interface:
- Implement responsive web interfaces
- Create mobile applications
- Handle user interaction patterns
- Implement accessibility standards

Client-Side Logic:
- Manage application state
- Handle client-side routing
- Implement real-time features
- Manage data fetching and caching

User Experience:
- Optimize performance and loading times
- Implement smooth animations and transitions
- Handle error states and loading indicators
- Ensure cross-browser compatibility

Integration:
- Integrate with backend APIs
- Handle authentication flows
- Implement real-time communication
- Manage client-side security
```

#### Key Skills and Expertise

- **Technologies**: React, Vue, Angular, HTML/CSS, JavaScript/TypeScript
- **Design**: UI/UX principles, responsive design, accessibility
- **Performance**: Bundle optimization, lazy loading, caching
- **Testing**: Unit testing, integration testing, end-to-end testing

### Database Engineer

#### Role Description

The Database Engineer specializes in data modeling, storage optimization, and database administration. They ensure data is stored efficiently, securely, and reliably.

#### Core Responsibilities

```
Data Modeling:
- Design database schemas and relationships
- Optimize data structures for performance
- Handle data normalization and denormalization
- Manage data versioning and migrations

Performance Optimization:
- Create and optimize database indexes
- Analyze and improve query performance
- Implement caching strategies
- Monitor database metrics and health

Administration:
- Manage database configuration and tuning
- Handle backup and recovery procedures
- Implement security and access controls
- Plan capacity and scaling strategies

Integration:
- Design data access patterns
- Implement repository and ORM layers
- Handle data synchronization
- Manage data pipelines and ETL processes
```

#### Key Skills and Expertise

- **Database Systems**: PostgreSQL, MongoDB, Redis, etc.
- **Query Optimization**: SQL tuning, index design
- **Administration**: Backup, recovery, monitoring
- **Data Modeling**: Schema design, normalization

### DevOps Engineer

#### Role Description

The DevOps Engineer specializes in infrastructure, deployment, monitoring, and operational concerns. They ensure systems run reliably in production.

#### Core Responsibilities

```
Infrastructure:
- Design and manage cloud infrastructure
- Implement infrastructure as code
- Handle containerization and orchestration
- Manage networking and security groups

Deployment:
- Create CI/CD pipelines
- Implement automated testing and deployment
- Handle environment management
- Manage configuration and secrets

Monitoring:
- Implement application and infrastructure monitoring
- Create alerting and notification systems
- Handle log aggregation and analysis
- Implement performance tracking

Operations:
- Handle incident response and troubleshooting
- Implement backup and disaster recovery
- Manage capacity planning and scaling
- Handle security updates and patching
```

#### Key Skills and Expertise

- **Cloud Platforms**: AWS, Azure, GCP
- **Containerization**: Docker, Kubernetes
- **CI/CD**: Jenkins, GitHub Actions, GitLab CI
- **Monitoring**: Prometheus, Grafana, ELK stack

## Quality and Support Archetypes

### Quality Assurance Engineer

#### Role Description

The Quality Assurance Engineer ensures system quality through comprehensive testing, validation, and quality processes. They catch issues before they reach production.

#### Core Responsibilities

```
Test Strategy:
- Develop comprehensive test plans
- Design test cases and scenarios
- Implement automated testing frameworks
- Manage test data and environments

Testing Execution:
- Perform manual and automated testing
- Execute integration and end-to-end tests
- Validate performance and security requirements
- Test error handling and edge cases

Quality Processes:
- Establish quality gates and criteria
- Review code and architectural decisions
- Validate requirements and specifications
- Ensure compliance with standards

Bug Management:
- Identify, document, and track defects
- Verify bug fixes and regressions
- Prioritize issues based on impact
- Communicate quality metrics
```

#### Key Skills and Expertise

- **Testing Tools**: Selenium, Jest, Cypress, JMeter
- **Test Design**: Test case design, boundary analysis
- **Automation**: Test automation frameworks and tools
- **Quality Processes**: QA methodologies, standards compliance

### Documentation Expert

#### Role Description

The Documentation Expert creates and maintains comprehensive documentation that enables effective use, maintenance, and evolution of the system.

#### Core Responsibilities

```
Technical Documentation:
- Create API documentation and guides
- Document system architecture and design
- Write installation and configuration guides
- Maintain troubleshooting documentation

User Documentation:
- Create user manuals and guides
- Write tutorials and getting-started guides
- Develop training materials
- Maintain FAQ and knowledge base

Process Documentation:
- Document development processes and workflows
- Create coding standards and guidelines
- Maintain project documentation
- Document lessons learned and best practices

Knowledge Management:
- Organize and structure documentation
- Maintain documentation versioning
- Ensure documentation accuracy and currency
- Facilitate knowledge transfer
```

#### Key Skills and Expertise

- **Technical Writing**: Clear, concise technical communication
- **Documentation Tools**: Markdown, wikis, documentation platforms
- **Information Architecture**: Organizing and structuring information
- **Domain Knowledge**: Understanding of technical domains being documented

### Performance Engineer

#### Role Description

The Performance Engineer focuses on system performance, scalability, and optimization. They ensure systems meet performance requirements under load.

#### Core Responsibilities

```
Performance Analysis:
- Analyze system performance bottlenecks
- Conduct load and stress testing
- Profile application performance
- Monitor system metrics and trends

Optimization:
- Optimize database queries and indexes
- Improve application code performance
- Implement caching strategies
- Optimize infrastructure configuration

Scalability:
- Design for horizontal and vertical scaling
- Implement load balancing strategies
- Plan capacity requirements
- Design auto-scaling mechanisms

Testing:
- Create performance test suites
- Implement continuous performance testing
- Validate performance requirements
- Benchmark system performance
```

#### Key Skills and Expertise

- **Performance Testing**: JMeter, LoadRunner, k6
- **Profiling**: Application profilers, APM tools
- **Optimization**: Code optimization, database tuning
- **Monitoring**: Performance monitoring tools and metrics

### Security Specialist

#### Role Description

The Security Specialist ensures system security through threat analysis, security implementation, and compliance validation.

#### Core Responsibilities

```
Security Analysis:
- Conduct security assessments and audits
- Identify security vulnerabilities and threats
- Perform penetration testing
- Review security architecture

Security Implementation:
- Implement authentication and authorization
- Handle encryption and key management
- Configure security controls and policies
- Manage security monitoring and logging

Compliance:
- Ensure compliance with security standards
- Implement regulatory requirements
- Manage security documentation
- Conduct security training

Incident Response:
- Respond to security incidents
- Investigate security breaches
- Implement security patches and updates
- Manage security communications
```

#### Key Skills and Expertise

- **Security Tools**: Vulnerability scanners, penetration testing tools
- **Compliance**: GDPR, SOX, HIPAA, etc.
- **Cryptography**: Encryption, hashing, digital signatures
- **Threat Analysis**: Risk assessment, threat modeling

## Archetype Selection and Combination

### Matching Archetypes to Festival Needs

#### Small Festivals (3-5 agents)

```
Essential Archetypes:
- Architect/Coordinator (combined role)
- Domain Specialist (primary domain)
- Quality Assurance Engineer

Optional Archetypes:
- Documentation Expert (can be combined with QA)
- Second Domain Specialist (if needed)
```

#### Medium Festivals (5-7 agents)

```
Core Team:
- Architect/Coordinator
- Backend Engineer
- Frontend Engineer (if applicable)
- Quality Assurance Engineer
- Documentation Expert

Extended Team:
- Database Engineer (if needed)
- DevOps Engineer (if deployment complexity requires)
```

#### Large Festivals (7+ agents)

```
Leadership:
- Architect/Coordinator
- Project Manager
- Technical Lead

Specialists:
- Backend Engineer
- Frontend Engineer
- Database Engineer
- DevOps Engineer

Support:
- Quality Assurance Engineer
- Documentation Expert
- Performance Engineer (if needed)
- Security Specialist (if needed)
```

### Hybrid Roles and Combinations

#### Full-Stack Engineer

Combines Backend and Frontend Engineer capabilities:

```
Responsibilities:
- Both server-side and client-side development
- End-to-end feature implementation
- API design and consumption
- Database integration and UI implementation
```

#### Site Reliability Engineer (SRE)

Combines DevOps and Performance Engineer capabilities:

```
Responsibilities:
- Infrastructure management and optimization
- Performance monitoring and improvement
- Incident response and reliability
- Automation and tooling
```

#### Technical Writer-Developer

Combines Documentation Expert with domain expertise:

```
Responsibilities:
- Writing documentation with deep technical understanding
- Creating code examples and tutorials
- Maintaining documentation currency through code involvement
- Bridging technical and user perspectives
```

## Archetype Evolution During Festivals

### Role Flexibility

Agents may need to adapt their roles based on:

- Festival phase requirements
- Discovered needs and gaps
- Agent availability and capacity
- Emerging technical challenges

### Cross-Training Opportunities

Orchestration enables knowledge sharing:

- Domain specialists learn from each other
- Support roles gain domain knowledge
- Leadership roles understand implementation details
- Quality processes improve through specialist input

### Archetype Emergence

New archetypes may emerge from:

- Unique festival requirements
- Technology stack specifics
- Organization-specific needs
- Lessons learned from previous festivals

## Implementation Guidelines

### Agent Assignment Process

1. **Analyze Festival Requirements**: Identify needed capabilities
2. **Map to Archetypes**: Select appropriate archetype combinations
3. **Assess Agent Capabilities**: Match available agents to archetypes
4. **Plan Role Development**: Identify training or support needs
5. **Define Hybrid Roles**: Combine archetypes when beneficial

### Role Clarity and Communication

- Clearly define each agent's archetype and responsibilities
- Document interaction patterns and dependencies
- Establish communication protocols between archetypes
- Create escalation procedures for role conflicts

### Success Metrics by Archetype

Track effectiveness of each archetype:

- **Architects**: System coherence and design quality
- **Specialists**: Domain implementation quality and efficiency
- **Quality Roles**: Defect rates and testing coverage
- **Support Roles**: Documentation quality and knowledge transfer

## Conclusion

Agent archetypes provide a framework for understanding and organizing the diverse skills needed in orchestrated festivals. By understanding these archetypes, you can:

1. **Plan Better Teams**: Ensure all necessary capabilities are covered
2. **Improve Coordination**: Understand interaction patterns between roles
3. **Develop Agents**: Identify growth opportunities and skill development needs
4. **Scale Orchestration**: Apply proven archetype patterns to new festivals

Remember that archetypes are guidelines, not rigid constraints. Adapt them to your specific context, and don't hesitate to create hybrid roles or new archetypes when your festival's needs require it.
