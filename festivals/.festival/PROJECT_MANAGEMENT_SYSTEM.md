# Festival Methodology - Markdown/YAML Project Management System

## Overview

Festival Methodology uses markdown/YAML for project tracking without external tools. This system tracks festival progress using standard markdown checkboxes and YAML structured data.

## Why Markdown/YAML?

### **Universal Compatibility**

- Works in any text editor, IDE, or markdown viewer
- No vendor lock-in or proprietary formats
- Version control friendly (git diff shows exact changes)
- Platform independent (works on any OS)

### **AI Agent Friendly**

- AI agents can read, parse, and update markdown naturally
- Structured YAML provides programmatic access to data
- Checkbox syntax is universally understood
- Easy to generate reports and status updates

### **Automation Ready**

- Parse checkboxes to calculate completion percentages
- Extract project status for CI/CD integration
- Generate dashboards from structured data
- Validate dependencies and gate compliance

### **Human Readable**

- Clear visual progress indicators
- No special tools required to understand status
- Easy to edit manually or programmatically
- Natural documentation format

## Three-Level Hierarchy Tracking

The system tracks the complete Festival Methodology hierarchy:

```
Festival
├── Phase 001: PLAN                    [✅] Completed
│   ├── 01_requirements_analysis       [✅] Completed  
│   │   ├── 01_user_research.md        [✅] Completed
│   │   ├── 01_security_requirements   [✅] Completed
│   │   └── 02_requirements_spec.md    [✅] Completed
│   ├── 02_architecture_design         [🚧] In Progress
│   │   ├── 01_system_architecture.md  [✅] Completed
│   │   └── 02_security_architecture   [🚧] In Progress
│   └── 03_feasibility_study          [ ] Not Started
└── Phase 002: DEFINE_INTERFACES       [ ] Not Started
```

### Status Indicators

| Symbol | Meaning | Usage |
|--------|---------|--------|
| `[ ]` | Not Started | Work hasn't begun |
| `[🚧]` | In Progress | Currently being worked on |
| `[✅]` | Completed | All requirements met and verified |
| `[❌]` | Blocked | Cannot proceed due to external dependency |
| `[🔄]` | Needs Review | Work done but requires approval |
| `[⚡]` | Ready | Dependencies met, can start immediately |

## File Structure

### Core TODO Files

```
festival_project_name/
├── TODO.md                    # Main tracking file (markdown format)
├── TODO.yaml                  # Structured data (YAML format) 
├── COMMON_INTERFACES.md       # Interface specifications
├── 001_PLAN/
│   ├── 01_requirements_analysis/
│   │   ├── 01_user_research.md
│   │   ├── 02_requirements_spec.md
│   │   └── results/           # Completed work artifacts
│   └── 02_architecture_design/
└── 002_DEFINE_INTERFACES/
```

### Template Files Available

| File | Purpose | Format |
|------|---------|--------|
| `FESTIVAL_TODO_TEMPLATE.md` | Main project tracking | Markdown with checkboxes |
| `FESTIVAL_TODO_TEMPLATE.yaml` | Structured project data | YAML for automation |
| `FESTIVAL_TODO_EXAMPLE.md` | Real-world usage example | Markdown example |

## Key Features

### 1. **Complete Project Visibility**

**Phase-Level Status:**

```markdown
### Phase Completion Status
- [✅] **001_PLAN** - Requirements and Architecture
- [🚧] **002_DEFINE_INTERFACES** - System Contracts (Critical Gate)
- [ ] **003_IMPLEMENT** - Build Solution  
- [ ] **004_REVIEW_AND_UAT** - User Acceptance
```

**Sequence-Level Status:**

```markdown
### Sequence Progress
- [✅] **01_requirements_analysis** (Foundation for all other work)
- [🚧] **02_architecture_design** (System blueprint)  
- [ ] **03_feasibility_study** (Risk and resource validation)
```

**Task-Level Status:**

```markdown
**Tasks**:
- [✅] 01_user_research.md
- [✅] 01_security_requirements.md *(parallel)*
- [🚧] 02_requirements_spec.md
- [ ] 03_testing_and_verify.md
```

### 2. **Progress Dashboard**

Real-time metrics calculated from checkboxes:

```markdown
## Progress Dashboard

Festival Progress: [████__________] 30%

Phase Breakdown:
001_PLAN:              [███████████████] 3/3 sequences ✅
002_DEFINE_INTERFACES: [█████__________] 2/3 sequences 🚧  
003_IMPLEMENT:         [_______________] 0/4 sequences ⏳

Total: 5/13 sequences completed (38%)
Tasks: 22/72 completed (31%)
```

### 3. **Critical Gate Enforcement**

Methodology compliance built into tracking:

```markdown
### Phase Gates (Must Complete Before Next Phase)
1. **001 → 002**: [✅] Requirements documented, [✅] Architecture approved
2. **002 → 003**: [❌] ALL interfaces FINALIZED, [❌] COMMON_INTERFACES.md status = FINALIZED
3. **003 → 004**: [ ] Implementation complete, [ ] Tests passing
4. **004 → DONE**: [ ] User acceptance passed, [ ] Production ready
```

### 4. **Dependency Tracking**

Clear prerequisite management:

```markdown
#### 02_architecture_design
**Status**: [🚧] In Progress
**Dependencies**: 01_requirements_analysis must be completed

#### 03_service_integration  
**Status**: [ ] Not Started
**Dependencies**: 01_backend_foundation must be completed
```

### 5. **Parallel Work Indicators**

Shows what can run simultaneously:

```markdown
**Tasks**:
- [✅] 01_user_research.md
- [✅] 01_security_requirements.md *(parallel)*  # Can run with 01_user_research
- [🚧] 02_requirements_spec.md                   # Waits for both 01_ tasks
```

### 6. **Risk & Blocker Management**

Structured tracking of issues:

```markdown
### Active Blockers
❌ BLOCKER_001: OAuth provider approvals pending
   Impact: Cannot complete external service integration
   Owner: Sarah (following up daily)
   ETA: 2024-01-25

### Risk Register  
🔺 HIGH: OAuth approval delays could push timeline
🔸 MED:  Database performance under load unknown
🔹 LOW:  Team capacity during holidays
```

### 7. **Decision Documentation**

Audit trail of key decisions:

```markdown
### Recent Decisions
2024-01-17 DECISION: Use JWT tokens with 24-hour expiry
  Rationale: Balance between security and user experience
  Impact: Reduces server storage, requires refresh flow
  Made by: Architecture team
```

## Automation Capabilities

### Parsing Checkboxes for Metrics

```python
# Example: Extract progress from markdown
import re

def calculate_progress(markdown_content):
    completed = len(re.findall(r'- \[✅\]', markdown_content))
    total = len(re.findall(r'- \[[\s\S]\]', markdown_content))
    return (completed / total) * 100 if total > 0 else 0
```

### Dependency Validation

```python
# Example: Validate phase gates
def validate_phase_transition(current_phase, todo_content):
    if current_phase == "002_DEFINE_INTERFACES":
        if "COMMON_INTERFACES.md status = FINALIZED" not in todo_content:
            return False, "Interfaces must be FINALIZED before Phase 003"
    return True, "Gate criteria met"
```

### Status Dashboard Generation

```python
# Example: Generate progress dashboard
def generate_dashboard(yaml_data):
    phases = yaml_data['festival']['phases']
    for phase in phases:
        completed_sequences = sum(1 for seq in phase['sequences'] if seq['status'] == 'completed')
        total_sequences = len(phase['sequences'])
        progress = (completed_sequences / total_sequences) * 100
        print(f"{phase['id']}: [{progress_bar(progress)}] {completed_sequences}/{total_sequences}")
```

## Integration Examples

### CI/CD Integration

```yaml
# .github/workflows/festival-status.yml
name: Festival Progress
on: [push]
jobs:
  update-status:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Calculate Progress
        run: |
          python .festival/scripts/calculate_progress.py TODO.md
      - name: Update Dashboard
        run: |
          python .festival/scripts/update_dashboard.py
```

### Slack Notifications

```python
# Send progress updates to Slack
def notify_progress_update(webhook_url, progress_data):
    message = {
        "text": f"🎯 Festival Progress Update",
        "blocks": [
            {
                "type": "section", 
                "text": {"type": "mrkdwn", "text": f"Overall Progress: {progress_data['overall']}%"}
            },
            {
                "type": "section",
                "text": {"type": "mrkdwn", "text": f"Active Phase: {progress_data['current_phase']}"}
            }
        ]
    }
    requests.post(webhook_url, json=message)
```

### Project Management Tool Sync

```python
# Sync with external tools (Jira, GitHub Issues, etc.)
def sync_to_github_issues(repo, todo_data):
    for phase in todo_data['phases']:
        for sequence in phase['sequences']:
            if sequence['status'] == 'in_progress':
                # Create or update GitHub issue
                create_github_issue(repo, sequence)
```

## Best Practices

### 1. **Daily Updates**

- Update checkboxes as work progresses: `[ ]` → `[🚧]` → `[✅]`
- Update sequence status when all tasks complete
- Update phase status when all sequences complete
- Mark blockers immediately with `[❌]`

### 2. **Weekly Reviews**

- Review overall progress metrics
- Update active work section
- Assess risks and dependencies  
- Plan next week's focus

### 3. **Phase Gate Reviews**

- Ensure ALL criteria met before phase transition
- **Phase 002 → 003 is CRITICAL**: No implementation until interfaces FINALIZED
- Document gate decisions in decision log

### 4. **Methodology Compliance**

- **Interface-First**: Phase 002 gates all implementation
- **Parallel Work**: Tasks with same numbers can run simultaneously
- **Quality Gates**: Every sequence ends with test → review → iterate
- **Step-Based Progress**: Track completed deliverables, not time

## Benefits Summary

### For Teams

- **Single Source of Truth**: All project status in one place
- **Clear Dependencies**: Know what's blocking progress
- **Methodology Enforcement**: Built-in festival principles
- **Progress Visibility**: Real-time completion tracking

### For AI Agents  

- **Native Format**: Easy to read and update markdown/YAML
- **Structured Data**: Programmatic access to project state
- **Automation Ready**: Parse, validate, and report automatically
- **Integration Friendly**: Connect to any external system

### For Project Managers

- **No Tool Dependencies**: Works with any text editor
- **Version Controlled**: Track changes over time
- **Customizable**: Adapt to any project structure  
- **Audit Trail**: Complete history of decisions and progress

---

## Getting Started

1. **Copy Templates**: Use `FESTIVAL_TODO_TEMPLATE.md` as starting point
2. **Customize Structure**: Adapt phases/sequences for your project
3. **Start Tracking**: Begin updating checkboxes as work progresses
4. **Automate**: Add scripts for progress calculation and reporting
5. **Integrate**: Connect to CI/CD, Slack, or project management tools

**The result**: A pure markdown/YAML project management system that provides complete festival tracking without external dependencies, while enabling powerful automation and integration capabilities.
