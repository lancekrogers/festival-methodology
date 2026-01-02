# Gates Directory

This directory contains custom quality gate configurations for the Festival Methodology.

## Structure

```
.festival/
├── gates/
│   └── policies/   # Named policy YAML files (e.g., strict.yml, lightweight.yml)
└── templates/
    └── gates/      # Global gate templates (e.g., QUALITY_GATE_TESTING.md)
```

## Policies

Policy files define sets of quality gates. Create a YAML file in `policies/` to define a named policy:

```yaml
# policies/team-custom.yml
version: 1
name: team-custom
description: Our team's custom gate workflow
append:
  - id: testing_and_verify
    template: QUALITY_GATE_TESTING
    enabled: true
  - id: team_review
    template: TEAM_REVIEW
    enabled: true
```

## Templates

Template files are Markdown files that define the content of each gate task. They are searched in order of precedence:

1. Sequence level: `<sequence>/.fest.templates/`
2. Phase level: `<phase>/.fest.templates/`
3. Festival level: `<festival>/.festival/templates/gates/`
4. Global gates: `festivals/.festival/templates/gates/`

If a template ID starts with `gates/`, fest checks `<festival>/gates/` first.

## Usage

```bash
# List available policies
fest gates list

# Show effective gates
fest gates show

# Apply a policy
fest gates apply strict --phase 002_IMPLEMENT

# Initialize an override file
fest gates init --sequence 002_IMPLEMENT/01_core
```

See `fest understand gates` for more information.
