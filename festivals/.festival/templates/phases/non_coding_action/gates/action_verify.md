---
id: action_verify
aliases:
  - action-check
  - ag-verify
description: Quality gate for verifying non-coding actions were completed correctly
---

# Task: Action Verification

**Task Number:** {{ .TaskNumber }} | **Parallel Group:** None | **Dependencies:** All action tasks | **Autonomy:** medium

## Objective

Verify that all non-coding actions (deployments, configurations, migrations, etc.) were completed correctly.

## Verification Checklist

### Pre-Action Verification

- [ ] Backup/rollback plan is in place
- [ ] Prerequisites were met
- [ ] Stakeholders were notified
- [ ] Maintenance window confirmed (if applicable)

### Action Execution

- [ ] All steps completed in order
- [ ] Logs captured for each step
- [ ] No unexpected errors occurred
- [ ] Timing was within expectations

### Post-Action Verification

- [ ] System is functioning normally
- [ ] All services are healthy
- [ ] Monitoring shows expected metrics
- [ ] Users can access the system

### Rollback Readiness

- [ ] Rollback procedure is documented
- [ ] Rollback has been tested
- [ ] Point of no return is identified
- [ ] Recovery time is acceptable

## Action Log

| Step | Action | Result | Timestamp |
|------|--------|--------|-----------|
| 1 | [Action description] | [ ] Success / [ ] Fail | [Time] |
| 2 | [Action description] | [ ] Success / [ ] Fail | [Time] |
| 3 | [Action description] | [ ] Success / [ ] Fail | [Time] |

## Verification Results

### Health Checks

| Check | Expected | Actual | Status |
|-------|----------|--------|--------|
| [Service 1] | [Expected state] | [Actual state] | [ ] Pass / [ ] Fail |
| [Service 2] | [Expected state] | [Actual state] | [ ] Pass / [ ] Fail |

### Metrics Validation

| Metric | Baseline | Current | Acceptable |
|--------|----------|---------|------------|
| [Metric 1] | [Before] | [After] | [ ] Yes / [ ] No |

## Issues Encountered

| Issue | Impact | Resolution |
|-------|--------|------------|
| [Issue 1] | [Impact] | [How resolved] |

## Definition of Done

- [ ] All actions completed
- [ ] Verification checks pass
- [ ] No critical issues
- [ ] System is stable

## Sign-Off

**Executor:** [Name/Agent]
**Date:** [Date]
**Status:** [ ] Verified / [ ] Issues Found

**Notes:**
[Summary of verification and any concerns]
