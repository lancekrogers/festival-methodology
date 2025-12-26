# Festival Methodology Extensions

Extensions provide specialized workflow patterns for specific project needs beyond the core 3-phase Festival Methodology. The base methodology handles most development scenarios, but some projects require additional specialized phases or different sequencing.

## Core 3-Phase Base Methodology

The standard Festival Methodology follows a simple, effective pattern:

```
001_PLAN → 002_IMPLEMENT → 003_REVIEW_AND_UAT
```

This covers the vast majority of software development scenarios:

- **001_PLAN**: Define goals, requirements, and approach
- **002_IMPLEMENT**: Execute the planned work in logical sequences  
- **003_REVIEW_AND_UAT**: Validate, test, and complete the goal

## When Extensions Are Needed

Extensions are **optional** and should only be used when the base 3-phase methodology isn't sufficient:

### Multi-System Projects

- **Extension**: `interface-planning/`
- **When**: Multiple services/systems need to interact
- **Adds**: 002_DEFINE_INTERFACES phase between PLAN and IMPLEMENT
- **Benefit**: Enables parallel development through contract-first design

### Research-Heavy Projects  

- **Extension**: `research-heavy/`
- **When**: Significant unknowns require exploration before planning
- **Adds**: Research and prototyping phases before planning
- **Benefit**: Reduces risk through early validation

### Compliance Workflows

- **Extension**: `compliance-workflow/`
- **When**: Regulatory requirements need formal documentation and approval
- **Adds**: Documentation and approval phases throughout workflow
- **Benefit**: Ensures regulatory compliance and audit trails

### MVP Iteration

- **Extension**: `mvp-iteration/`
- **When**: Need to build and validate quickly, then enhance
- **Adds**: Multiple build-validate-enhance cycles
- **Benefit**: Early feedback and iterative improvement

## Extension Activation

Extensions are **suggested by AI agents** based on project characteristics discovered during planning:

### Planning Agent Detection

When the festival planning agent detects:

- Multiple interacting systems → Suggests `interface-planning`
- Unknown technical feasibility → Suggests `research-heavy`
- Regulatory requirements → Suggests `compliance-workflow`
- MVP-first approach → Suggests `mvp-iteration`

### Manual Selection

Humans can also explicitly request extensions:

```
"I need the interface-planning extension because we have 5 microservices that need to interact"
```

## Extension Structure

Each extension contains:

```
extension-name/
├── extension.md          # When to use, benefits, trade-offs
├── phase-patterns.md     # How phases integrate with base methodology
├── templates/           # Extension-specific templates
│   ├── phase-templates/
│   └── sequence-templates/
└── examples/           # Real project examples using this extension
```

## Available Extensions

### [Interface Planning](interface-planning/)

**Purpose**: Enable parallel development in multi-system architectures  
**Phase Addition**: 002_DEFINE_INTERFACES between PLAN and IMPLEMENT  
**Use When**: Multiple services, external integrations, team parallelization needed

### [Research Heavy](research-heavy/)

**Purpose**: Handle projects with significant technical unknowns  
**Phase Addition**: Research and prototyping phases before main workflow  
**Use When**: New technologies, feasibility questions, innovation projects

### [Compliance Workflow](compliance-workflow/)

**Purpose**: Meet regulatory and audit requirements  
**Phase Addition**: Documentation and approval gates throughout  
**Use When**: Healthcare, finance, government, or other regulated domains

### [MVP Iteration](mvp-iteration/)

**Purpose**: Build-measure-learn cycles for rapid validation  
**Phase Addition**: Multiple plan-implement-validate cycles  
**Use When**: Startup environments, new product development, user validation needed

## Extension Philosophy

### Keep Base Simple

The core 3-phase methodology should handle 80% of projects cleanly without extensions. Extensions add complexity and should only be used when the additional structure provides clear value.

### Step-Based Focus Maintained

All extensions maintain the core Festival principle: **thinking in steps toward goals, not time estimates**. Extensions add logical step progressions, not time-based scheduling.

### Optional, Not Mandatory

Extensions are suggestions from AI agents based on project characteristics. Humans always have final decision on whether an extension adds value or unnecessary complexity.

### Composable Design

Some extensions can be combined when projects have multiple specialized needs:

- Interface Planning + Compliance Workflow (regulated multi-system projects)
- Research Heavy + MVP Iteration (innovation projects with validation cycles)

## Implementation Guidelines

### For AI Agents

1. **Start with base methodology** - Always assume 3-phase unless clear indicators suggest extension
2. **Ask, don't assume** - Suggest extensions based on detected characteristics, get human confirmation
3. **Explain the trade-off** - Extensions add complexity; ensure the benefit justifies the cost
4. **Focus on the goal** - Extensions should accelerate goal achievement, not slow it down

### For Humans  

1. **Default to simple** - Use base 3-phase methodology unless you have specific need for extension
2. **Understand the cost** - Extensions add planning overhead and coordination complexity
3. **Choose based on value** - Extensions should solve real problems, not theoretical ones
4. **Start small** - Can always add extensions later if needs emerge during execution

## Extension Development

To create new extensions:

1. **Identify clear use case** - What specific project characteristics require different workflow?
2. **Define phase integration** - How do new phases integrate with base methodology?
3. **Create templates** - What specialized templates support the extension workflow?
4. **Document activation criteria** - When should AI agents suggest this extension?
5. **Provide examples** - Real projects that demonstrate extension value

Extensions should be **narrowly focused** on specific project characteristics rather than trying to be general-purpose methodology alternatives.

---

Remember: **Extensions are powerful tools for specific needs, but the base 3-phase methodology should handle most projects effectively.** Choose extensions when they clearly accelerate your specific goal achievement, not because they seem theoretically useful.
