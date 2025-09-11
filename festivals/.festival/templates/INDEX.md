# Festival Methodology Template Index

Quick reference guide for selecting the right template for your needs.

## Template Selection Decision Tree

```
What do you need to create?
│
├─ Starting a new festival?
│  ├─ Use: FESTIVAL_OVERVIEW_TEMPLATE.md (overview and context)
│  └─ Use: FESTIVAL_GOAL_TEMPLATE.md (goals and evaluation)
│
├─ Defining project standards?
│  └─ Use: FESTIVAL_RULES_TEMPLATE.md
│
├─ Planning interfaces/contracts?
│  └─ Use: COMMON_INTERFACES_TEMPLATE.md
│
├─ Creating goals?
│  ├─ Festival level?
│  │  └─ Use: FESTIVAL_GOAL_TEMPLATE.md
│  ├─ Phase level?
│  │  └─ Use: PHASE_GOAL_TEMPLATE.md
│  └─ Sequence level?
│     └─ Use: SEQUENCE_GOAL_TEMPLATE.md
│
├─ Creating a task?
│  ├─ Complex/Critical task?
│  │  └─ Use: TASK_TEMPLATE.md
│  └─ Simple/Standard task?
│     └─ Use: TASK_TEMPLATE_SIMPLE.md
│
├─ Planning a sequence?
│  ├─ Use: SEQUENCE_TEMPLATE.md (structure)
│  └─ Use: SEQUENCE_GOAL_TEMPLATE.md (goals)
│
├─ Planning a phase?
│  ├─ Use: PHASE_TEMPLATE.md (structure)
│  └─ Use: PHASE_GOAL_TEMPLATE.md (goals)
│
└─ Tracking progress?
   ├─ Manual/Human tracking?
   │  └─ Use: FESTIVAL_TODO_TEMPLATE.md
   └─ Automated/CI tracking?
      └─ Use: FESTIVAL_TODO_TEMPLATE.yaml
```

## Template Catalog

### Project-Level Templates

| Template | Purpose | When to Use | Lines |
|----------|---------|-------------|-------|
| **FESTIVAL_OVERVIEW_TEMPLATE.md** | Define project context, stakeholders, approach | Starting any new festival | ~200 |
| **FESTIVAL_GOAL_TEMPLATE.md** | Track festival goals, KPIs, and evaluation | Alongside FESTIVAL_OVERVIEW | ~250 |
| **FESTIVAL_RULES_TEMPLATE.md** | Document project standards, conventions, guidelines | After initial planning, before implementation | ~150 |
| **COMMON_INTERFACES_TEMPLATE.md** | Define all system interfaces and contracts | Phase 002_DEFINE_INTERFACES | ~400 |

### Goal Templates

| Template | Purpose | When to Use | Lines |
|----------|---------|-------------|-------|
| **FESTIVAL_GOAL_TEMPLATE.md** | Festival-level goals and evaluation framework | Festival inception | ~250 |
| **PHASE_GOAL_TEMPLATE.md** | Phase-specific goals and success criteria | When creating each phase | ~150 |
| **SEQUENCE_GOAL_TEMPLATE.md** | Sequence-level goals and progress tracking | When creating each sequence | ~120 |

### Planning Templates

| Template | Purpose | When to Use | Lines |
|----------|---------|-------------|-------|
| **PHASE_TEMPLATE.md** | Structure a complete phase with objectives and sequences | Planning major milestones | ~60 |
| **SEQUENCE_TEMPLATE.md** | Plan a sequence of related tasks | Organizing work within a phase | ~80 |

### Task Templates

| Template | Purpose | When to Use | Lines |
|----------|---------|-------------|-------|
| **TASK_TEMPLATE.md** | Comprehensive task with all sections | Complex, critical, or unfamiliar tasks | 188 |
| **TASK_TEMPLATE_SIMPLE.md** | Streamlined task essentials only | Simple, routine, or well-understood tasks | ~40 |

### Tracking Templates

| Template | Purpose | When to Use | Lines |
|----------|---------|-------------|-------|
| **FESTIVAL_TODO_TEMPLATE.md** | Markdown progress tracking with checkboxes | Human-readable tracking, documentation | ~300 |
| **FESTIVAL_TODO_TEMPLATE.yaml** | Structured YAML for automation | CI/CD integration, automated reporting | ~400 |

## Template Quick Reference

### FESTIVAL_OVERVIEW_TEMPLATE.md
**Creates:** `FESTIVAL_OVERVIEW.md` in your festival root

**Key Sections:**
- Project Goal (one clear sentence)
- Success Criteria (measurable outcomes)
- Problem Statement (current vs desired state)
- Stakeholder Matrix (users, team, constraints)
- High-Level Phases (initial structure)

**Example Output:**
```markdown
# Festival Overview: User Authentication System

## Project Goal
Build a secure user authentication system with email/password login, 
social auth, and role-based access control.

## Success Criteria
- [ ] Users can register, login, and logout
- [ ] OAuth with Google and GitHub works
- [ ] Role-based permissions enforced
```

### COMMON_INTERFACES_TEMPLATE.md
**Creates:** `COMMON_INTERFACES.md` in your festival root

**Key Sections:**
- API Endpoints (REST/GraphQL/RPC)
- Data Models (schemas, types)
- Function Signatures (public interfaces)
- Event Contracts (pub/sub, webhooks)
- Error Codes (standardized responses)

**Critical:** Complete this BEFORE implementation begins!

### FESTIVAL_RULES_TEMPLATE.md
**Creates:** `FESTIVAL_RULES.md` in your festival root

**Key Sections:**
- Code Standards (style, patterns)
- Quality Gates (coverage, performance)
- Security Requirements (auth, encryption)
- Team Agreements (review process, communication)
- Technology Constraints (stack, dependencies)

### TASK_TEMPLATE.md vs TASK_TEMPLATE_SIMPLE.md

**Full Template Includes:**
- Objective & Context
- Requirements & Deliverables
- Pre-Task Checklist
- Implementation Steps (detailed)
- Testing Commands
- Technical Notes
- Resources & Links
- Completion Checklist
- Good vs Bad Examples

**Simple Template Includes:**
- Objective
- Requirements (checklist)
- Implementation Steps (brief)
- Definition of Done

**Choose Simple When:**
- Task is well-understood
- Team has done similar work
- Risk is low
- Time is critical

**Choose Full When:**
- Task is complex or novel
- Multiple people involved
- High risk or critical path
- Need detailed documentation

### Progress Tracking Templates

**Markdown Format (FESTIVAL_TODO_TEMPLATE.md):**
```markdown
## Phase 001_PLAN [████████░░] 80%
### 01_requirements [✅] Complete
- [x] 01_user_research.md
- [x] 02_requirements_spec.md
### 02_architecture [🚧] In Progress
- [x] 01_system_design.md
- [ ] 02_database_schema.md
```

**YAML Format (FESTIVAL_TODO_TEMPLATE.yaml):**
```yaml
phases:
  - id: "001_PLAN"
    status: "in_progress"
    completion: 80
    sequences:
      - id: "01_requirements"
        status: "completed"
        tasks:
          - id: "01_user_research"
            status: "completed"
```

## Usage Patterns

### Pattern 1: New Festival Setup
1. `FESTIVAL_OVERVIEW_TEMPLATE.md` → Define goals
2. `FESTIVAL_RULES_TEMPLATE.md` → Set standards
3. `PHASE_TEMPLATE.md` → Structure phases
4. `SEQUENCE_TEMPLATE.md` → Plan sequences
5. `COMMON_INTERFACES_TEMPLATE.md` → Define interfaces
6. `TASK_TEMPLATE.md` → Create tasks
7. `FESTIVAL_TODO_TEMPLATE.md` → Track progress

### Pattern 2: Quick Task Creation
1. Assess complexity
2. Choose template (simple vs full)
3. Fill required sections
4. Add to sequence directory
5. Update TODO.md

### Pattern 3: Interface-First Development
1. Complete Phase 001 planning
2. Use `COMMON_INTERFACES_TEMPLATE.md` extensively
3. Define ALL interfaces before coding
4. Distribute interface docs to all developers
5. Begin parallel implementation

## Template Customization Tips

### Making Templates Your Own

1. **Remove Unnecessary Sections**
   - Delete sections that don't apply
   - Keep templates lean and relevant

2. **Add Domain-Specific Sections**
   - Add sections for your industry
   - Include compliance requirements
   - Add team-specific needs

3. **Create Variations**
   - `TASK_TEMPLATE_FRONTEND.md`
   - `TASK_TEMPLATE_BACKEND.md`
   - `TASK_TEMPLATE_DEVOPS.md`

4. **Adjust Complexity**
   - Startup: Lean templates
   - Enterprise: Comprehensive templates
   - Open Source: Documentation-heavy

### Template Evolution

Templates should evolve with your project:

**Early Stage:** Simple templates, focus on speed
**Growth Stage:** Add quality sections
**Mature Stage:** Comprehensive documentation
**Maintenance:** Streamline based on experience

## Common Questions

**Q: Can I modify templates?**
A: Yes! Templates are starting points. Customize freely.

**Q: Which tracking format should I use?**
A: Use .md for humans, .yaml for machines. Many teams use both.

**Q: How detailed should tasks be?**
A: Detailed enough that someone unfamiliar could complete them.

**Q: Must I use all sections?**
A: No. Use what adds value, remove what doesn't.

**Q: Can I create new templates?**
A: Absolutely! Share them with the community.

## Template Metrics

**Average completion times with templates:**
- Festival Overview: 15-30 minutes
- Common Interfaces: 1-2 hours
- Task (Simple): 5-10 minutes
- Task (Full): 15-20 minutes
- Phase Planning: 20-30 minutes
- Sequence Planning: 10-15 minutes

**Without templates:** 2-3x longer, less consistent

## Next Steps

1. Start with `FESTIVAL_OVERVIEW_TEMPLATE.md`
2. Review examples in `../examples/`
3. Customize templates for your project
4. Share improvements via [CONTRIBUTING.md](../../../CONTRIBUTING.md)