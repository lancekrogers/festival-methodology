# Getting Started with Festival Methodology and Claude Code

This guide walks through how to use Festival Methodology in practice with Claude Code or similar AI assistants.

## Initial Setup

### 1. Create Your Planning Directory

Place the Festival Methodology alongside your project repository:

```
your-workspace/
├── your-project/           # Your actual code repository
└── ai_planning/           # AI planning directory
    └── festivals/         # Copy the festivals/ directory here
```

### 2. Start Planning Your Festival

1. **Open Claude Code** in your workspace
2. **Ask Claude to review** the `festivals/README.md` in your planning directory
3. **Discuss your project goals** - Chat with Claude about what you want to accomplish
4. **Use Claude Code to create** the festival structure and documentation

## Two Approaches to Festival Planning

### Approach 1: Comprehensive Planning

Plan everything upfront:
- Create the complete festival structure
- Define all phases, sequences, and tasks
- Review and refine the entire plan
- Make adjustments before starting work

### Approach 2: Iterative Development

Start simple and evolve:
- Create basic festival structure with high-level goals
- Detail out only the immediate sequences
- Expand the plan as you learn more
- Adjust based on what you discover

## Using Custom Agents (Optional)

Custom agents can help maintain consistency, but they're not required. Often you'll get better results working directly with Claude Code.

### Setting Up Custom Agents

1. Copy agent files to `.claude/agents/` in your workspace
2. Ask Claude to tailor these agents for your specific festival
3. Tell Claude to use these agents when working on the festival

### Working Without Custom Agents

Simply:
1. Create the festival structure
2. Give Claude your requirements
3. Point Claude to specific tasks or sequences to work on
4. Review progress and adjust as needed

## Execution Workflow

### Starting Work

Tell Claude something like:
- "Let's work on sequence 01 of phase 001"
- "Please complete task 01_api_design.md"
- "Update the TODO.md file as you complete tasks"

### Monitoring Progress

- Review what Claude is doing regularly
- Ensure it aligns with your goals
- Make adjustments to the plan as needed
- Keep the TODO.md file updated

## Important Notes

### This Is Not a Polished Product

Festival Methodology is a framework that requires adaptation:
- Results will vary based on your approach
- You'll need to find what works for your workflow
- Expect to iterate and refine your process

### The Power Is in the Process

Using this methodology, you can achieve:
- Years worth of progress in months
- Complex software products built systematically
- Clear documentation of decisions and progress
- Ability to parallelize work across multiple agents

## Tips for Success

1. **Start Simple** - Don't over-plan your first festival
2. **Iterate Quickly** - Adjust based on what's working
3. **Trust the Process** - Let the structure guide the work
4. **Stay Flexible** - Adapt the methodology to your needs
5. **Document Learnings** - Note what works for future festivals

## Example First Session

```
You: Please review the festivals/README.md file in my ai_planning directory

Claude: [Reviews the methodology]

You: I want to create a festival for building a user authentication system. 
     Let's start by planning out the high-level phases.

Claude: [Creates festival structure]

You: Great, now let's detail out phase 001_PLAN. What sequences should we have?

Claude: [Develops sequences and tasks]

You: Let's start working on sequence 01_requirements_gathering

Claude: [Begins executing tasks]
```

## Remember

The methodology is designed to evolve with use. Don't worry about getting it perfect the first time - you'll refine your approach as you see what works for your specific needs.