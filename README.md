# Festival Methodology

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Methodology Version](https://img.shields.io/badge/version-0.3.0-blue)](CHANGELOG.md)

A goal-oriented, AI-native project management methodology that prioritizes concrete objectives over process overhead. Festival Methodology uses a three-level hierarchy (**Phases → Sequences → Tasks**) with interface-first planning to enable parallel development and systematic progress.

## Highlights

- **AI-native design** - Built specifically for AI agent orchestration, not adapted from human methodologies
- **Outcome-focused** - Complete meaningful features instead of endless incremental progress  
- **Three-level hierarchy** - Phases → Sequences → Tasks with systematic numbering for clear progress tracking
- **Interface-first development** - Define all system contracts before implementation to enable parallel work
- **Rapid iteration** - Test methodology changes in days, not months, with immediate feedback on results
- **Framework, not prescription** - Adapt to your workflow and project needs

**Intended for:** Developers building AI-orchestrated workflows and complex software projects with AI assistance. Not a generic PM tool.

### Quick Example

Festival tasks are concrete and actionable, not abstract:

```markdown
# Task: 01_create_user_table_and_model.md
## Objective
Create PostgreSQL user table and Sequelize model with email/password authentication

## Requirements  
- [ ] Create `users` table with id, email, password_hash, created_at, updated_at
- [ ] Create `models/User.js` with Sequelize model definition
- [ ] Add email validation and password hashing methods

## Implementation Steps
1. Run: `npx sequelize-cli migration:generate --name create-users-table`
2. Edit migration file with SQL schema
3. Create `models/User.js` with Sequelize model
4. Test with: `npm test -- --grep "User model"`
```

This level of detail enables AI agents to execute tasks systematically with clear success criteria.

## Why Festival Methodology?

Traditional agile methodologies were designed for human teams to deliver constant incremental progress through fixed 2-week sprints. This artificial timebox exists because projects can drift indefinitely without deadlines. But the real problem with agile isn't just that it's designed for humans - it's that **constant minor progress isn't the same as completing something meaningful**.

With AI agents, we can finally escape this trap. AI agents work continuously, parallelize complex tasks, and maintain context across long development cycles. This means we can actually **complete ambitious goals** rather than just iterating forever.

Festival Methodology is designed for **completion AND iteration through completed features**. Instead of endless minor updates, you iterate by completing meaningful chunks of work. Whether you're building a major feature, refactoring a codebase, or creating an entire product, this framework helps you:

- **Define a clear goal** and work systematically toward completion
- **Scale appropriately** - use it for anything requiring more than a handful of prompts
- **Automate your workflow** - let AI agents handle the execution while you focus on direction

## Why Festival Works with AI

Festival Methodology thrives with AI agents because of:

- **Rapid feedback loops** - Test and refine your approach in days, not months
- **Simple, readable rules** - Both humans and AI agents can understand and follow the structure
- **Continuous monitoring** - See immediately what's working and what needs adjustment
- **Outcome-based iteration** - Adjust based on completed features, not abstract metrics

The methodology evolves through actual usage. When something doesn't work, you'll know quickly and can adjust. When something works well, you can double down on it.

**This is a framework, not a prescription.** Clone it, adapt it to your workflow, and discover what works for your projects through real-world application.

## Getting Started

**New to Festival Methodology?** See [GETTING_STARTED.md](GETTING_STARTED.md) for a practical walkthrough of using Festival with Claude Code.

## Quick Start - New Project Setup

### 1. Copy the Starter Kit to Your Project

Copy the entire `festivals/` directory from this repository to your AI workspace root:

```bash
cp -r festivals/ /path/to/your/ai-workspace/
```

Your workspace structure should look like:

```
your-ai-workspace/
├── your-project-repos/     # Your existing repositories
└── festivals/              # Festival planning directory (copied from this repo)
    ├── .festival/          # Methodology templates and agents
    │   ├── templates/      # All methodology templates
    │   ├── agents/         # Custom AI agents for festival workflow
    │   └── README.md       # Implementation guide
    └── README.md           # How to create your first festival
```

### 2. Create Your First Festival

1. Navigate to your `festivals/` directory
2. Follow the instructions in `festivals/README.md` to create your first festival
3. Use the custom AI agents in `festivals/.festival/agents/` for guided setup

### 3. Use the Custom AI Agents

This starter kit includes three specialized AI agents:

- **festival-planning-agent** - Conducts structured interviews to create complete festival plans
- **festival-review-agent** - Validates festival structures for quality and methodology compliance
- **festival-methodology-manager** - Enforces methodology principles during execution

## What's Included

### Methodology Templates (`festivals/.festival/templates/`)

- **FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md** - Core methodology documentation
- **COMMON_INTERFACES_TEMPLATE.md** - Protocol-agnostic interface definition system
- **TASK_TEMPLATE.md** - Template for individual tasks
- **TASK_EXAMPLES.md** - Concrete examples of well-written tasks
- **FESTIVAL_RULES_TEMPLATE.md** - Template for project standards
- **FESTIVAL_OVERVIEW_TEMPLATE.md** - Template for project overview and goals

### Custom AI Agents (`festivals/.festival/agents/`)

- **festival_planning_agent.md** - Systematic project planning and structure creation
- **festival_review_agent.md** - Quality assurance and methodology compliance validation
- **festival_methodology_manager.md** - Process enforcement during festival execution

### Documentation

- **festivals/README.md** - Complete guide to using the methodology
- **festivals/.festival/README.md** - Implementation details and agent usage

## Key Methodology Principles

1. **Interface-First Development** - Define all system interfaces before implementation begins
2. **Three-Level Hierarchy** - Phases → Sequences → Tasks with systematic numbering
3. **Quality Verification** - Testing and review at every sequence level
4. **Parallel Development** - Interface contracts enable simultaneous team work
5. **Step-Based Planning** - Focus on development steps, not time estimates

## Workflow Overview

1. **Copy starter kit** to your AI workspace
2. **Use festival-planning-agent** to create festival structure through guided interview
3. **Use festival-review-agent** to validate structure before execution
4. **Use festival-methodology-manager** during execution to maintain methodology compliance
5. **Execute festivals** following the systematic three-level hierarchy

## Support & Documentation

- **Primary Documentation**: See `festivals/README.md` for complete usage guide
- **Implementation Guide**: See `festivals/.festival/README.md` for technical details
- **Agent Usage**: Custom AI agents include detailed usage instructions
- **Templates**: All templates include examples and guidance for proper usage

## Community & Support

- **Share Experience**: Open an [Issue](../../issues) to share what worked, what didn't, and what you discovered
- **Ask Questions**: Use [Discussions](../../discussions) for methodology questions and brainstorming
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) - we value real-world usage reports most

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Festival Methodology emerged from iterative refinement across a dozen AI-assisted software projects and my personal experience planning and developing software projects for >10 years. It continues to evolve based on real-world usage.

---

**Quick Start**: `cp -r festivals/ /your/ai-workspace/` → Navigate to `festivals/` → Follow `README.md` → Create your first festival!

