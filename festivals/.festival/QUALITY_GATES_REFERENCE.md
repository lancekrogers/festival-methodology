# Quality Gates Reference

Quality gates are the standard testing, review, and iteration tasks that should be included at the end of EVERY implementation sequence. They ensure work meets standards before proceeding and provide consistent quality control across your festival.

## Standard Quality Gates

Every implementation sequence should end with these three tasks in order:

```
XX_testing_and_verify.md      ← Verify functionality works as specified
XX_code_review.md             ← Review code quality and standards compliance  
XX_review_results_iterate.md  ← Address findings and iterate until acceptable
```

## Why Quality Gates Matter

**Without Quality Gates:**

- Bugs compound across sequences
- Code quality degrades over time
- Standards drift as team members interpret them differently
- Integration issues surface late in the project
- Technical debt accumulates unchecked

**With Quality Gates:**

- Issues caught early when they're cheaper to fix
- Consistent quality standards maintained
- Knowledge shared through reviews
- Team learns and improves continuously
- Stakeholder confidence in deliverables

## Task 1: Testing and Verify

### Purpose

Verify that all sequence deliverables work as specified and meet functional requirements.

### Standard Template

```markdown
# Task: XX_testing_and_verify.md

**Autonomy Level:** medium

## Objective
Verify that all sequence deliverables work as specified and meet quality standards.

## Requirements
- [ ] All functionality works as described in task objectives
- [ ] Unit tests pass for new/modified code
- [ ] Integration tests cover main workflows 
- [ ] Manual testing confirms user stories
- [ ] Performance meets documented requirements
- [ ] Security validation passes (if applicable)
- [ ] Error handling works correctly
- [ ] Edge cases handled appropriately

## Testing Steps
1. **Automated Tests**
   - Run: `[specific test command for your project]`
   - Verify all tests pass
   - Check coverage meets project standards

2. **Manual Verification**  
   - Test each deliverable against its requirements
   - Verify user workflows function correctly
   - Check error conditions behave appropriately

3. **Performance Testing** (if applicable)
   - Load test under expected conditions
   - Verify response times meet requirements
   - Check resource utilization is reasonable

## Definition of Done
- [ ] All automated tests pass
- [ ] Manual testing confirms functionality
- [ ] Performance benchmarks met
- [ ] No critical or high-severity issues found
- [ ] Documentation updated with any new requirements
```

### Domain-Specific Testing

**Backend API Testing:**

```markdown
## Additional Requirements
- [ ] API endpoints respond correctly to valid requests
- [ ] Invalid requests return appropriate error codes  
- [ ] Authentication/authorization works as designed
- [ ] Database transactions handle edge cases
- [ ] API documentation matches implementation
```

**Frontend Component Testing:**

```markdown
## Additional Requirements  
- [ ] Component renders correctly in target browsers
- [ ] User interactions work as expected
- [ ] Form validation provides clear feedback
- [ ] Responsive design works on target screen sizes
- [ ] Accessibility requirements met
```

**Database Schema Testing:**

```markdown
## Additional Requirements
- [ ] Migrations run successfully on clean database
- [ ] Constraints prevent invalid data entry
- [ ] Indexes improve query performance as expected
- [ ] Backup/restore procedures work correctly
- [ ] Data integrity maintained under concurrent access
```

**DevOps/Infrastructure Testing:**

```markdown
## Additional Requirements
- [ ] Deployment process completes successfully
- [ ] Service starts correctly in target environment
- [ ] Monitoring and logging capture expected events
- [ ] Rollback process works if needed
- [ ] Security configurations applied correctly
```

## Task 2: Code Review

### Purpose

Review code quality, architecture alignment, and standards compliance before finalizing the sequence.

### Standard Template

```markdown
# Task: XX_code_review.md

**Autonomy Level:** low

## Objective
Review all code and deliverables for quality, standards compliance, and architectural alignment.

## Requirements
- [ ] Code follows project style guidelines (see FESTIVAL_RULES.md)
- [ ] Architecture aligns with COMMON_INTERFACES.md
- [ ] Security best practices followed
- [ ] Performance considerations addressed
- [ ] Documentation is complete and accurate
- [ ] Error handling is comprehensive
- [ ] Code is maintainable and readable
- [ ] Dependencies are justified and secure

## Review Process
1. **Automated Checks**
   - Run linting: `[linting command]`
   - Run security scan: `[security scan command]`  
   - Check dependency vulnerabilities: `[dependency check command]`

2. **Manual Review**
   - Review each file for clarity and maintainability
   - Verify adherence to project conventions
   - Check for security vulnerabilities
   - Assess performance implications

3. **Architecture Review**
   - Verify alignment with COMMON_INTERFACES.md
   - Check integration points match specifications
   - Assess impact on system complexity
   - Review for technical debt

## Review Checklist
- [ ] **Code Quality**: Clean, readable, well-structured
- [ ] **Standards**: Follows project conventions and style guide
- [ ] **Security**: No vulnerabilities, secrets, or unsafe practices
- [ ] **Performance**: Efficient algorithms and resource usage
- [ ] **Documentation**: Clear comments, updated README/docs
- [ ] **Testing**: Adequate test coverage and quality
- [ ] **Architecture**: Aligns with system design
- [ ] **Dependencies**: Justified additions, security checked
```

### Code Review Guidelines

**What to Look For:**

**Code Quality:**

- Clear variable and function names
- Appropriate code organization and structure
- Consistent formatting and style
- Elimination of dead or commented-out code

**Security:**

- No hardcoded secrets or credentials  
- Input validation on all external data
- Proper authentication/authorization
- Safe handling of sensitive data

**Performance:**

- Efficient algorithms for data processing
- Proper resource cleanup (memory, connections, etc.)
- Appropriate caching strategies
- Database query optimization

**Maintainability:**

- Code is self-documenting where possible
- Complex logic has explanatory comments
- Dependencies are minimal and justified
- Code follows SOLID principles

## Task 3: Review Results and Iterate

### Purpose

Address all findings from testing and code review, iterating until work meets acceptance criteria.

### Standard Template

```markdown
# Task: XX_review_results_iterate.md

**Autonomy Level:** medium

## Objective
Address all findings from testing and code review, iterating until sequence meets acceptance criteria.

## Requirements
- [ ] All test failures resolved or requirements clarified
- [ ] Code review findings addressed satisfactorily
- [ ] Performance issues resolved
- [ ] Security concerns resolved
- [ ] Documentation updated based on feedback
- [ ] Stakeholder acceptance obtained (if required)
- [ ] Quality standards met for sequence completion

## Iteration Process
1. **Triage Findings**
   - Categorize issues by severity (critical/high/medium/low)
   - Identify quick fixes vs significant rework needed
   - Determine if any findings require requirement clarification

2. **Address Issues**
   - Fix critical and high-priority issues immediately
   - Create plan for medium-priority issues
   - Document low-priority issues for future consideration
   - Update tests and documentation as needed

3. **Re-verify**
   - Re-run tests after changes
   - Re-review code if significant changes made
   - Confirm issues are actually resolved

4. **Stakeholder Sign-off** (if required)
   - Demonstrate functionality to stakeholders
   - Address any feedback or concerns
   - Obtain formal acceptance

## Definition of Done
- [ ] All critical and high-priority findings resolved
- [ ] Medium-priority findings resolved or explicitly deferred
- [ ] All tests pass after changes
- [ ] Code quality standards met
- [ ] Stakeholder acceptance obtained
- [ ] Sequence deliverables finalized and documented
```

### Common Iteration Patterns

**Minor Issues (1-2 iteration cycles):**

- Fix bugs found in testing
- Address code style violations
- Update documentation gaps
- Resolve merge conflicts

**Moderate Issues (3-5 iteration cycles):**

- Performance optimizations
- Security improvements
- Architecture adjustments
- Significant test additions

**Major Issues (multiple sequences):**

- Requirement clarification needed
- Significant architectural changes
- Major performance problems
- Security vulnerabilities requiring design changes

## Quality Gate Customization

### By Project Type

**Startup/MVP Projects:**

```markdown
## Streamlined Requirements
- [ ] Core functionality works
- [ ] No security vulnerabilities
- [ ] Basic tests pass
- [ ] Code is readable
```

**Enterprise Projects:**

```markdown  
## Comprehensive Requirements
- [ ] Full test suite passes (90%+ coverage)
- [ ] Security audit completed
- [ ] Performance benchmarks met
- [ ] Compliance requirements satisfied
- [ ] Documentation review completed
- [ ] Stakeholder approval obtained
```

**Open Source Projects:**

```markdown
## Community-Focused Requirements
- [ ] Contribution guidelines followed
- [ ] License compatibility verified
- [ ] Breaking changes documented
- [ ] Community feedback incorporated
- [ ] Backward compatibility maintained
```

### By Sequence Risk Level

**High-Risk Sequences** (security, data, critical path):

- More thorough testing requirements
- Additional security validation
- Performance benchmarking required
- External review required

**Medium-Risk Sequences** (standard features):

- Standard quality gate requirements
- Normal test coverage expectations
- Regular review process

**Low-Risk Sequences** (documentation, styling):

- Streamlined testing requirements
- Faster review cycle
- Focus on functionality over performance

## Integration with Project Standards

### Link to Festival Rules

Quality gates should reference your project's specific standards:

```markdown
## Rules Compliance
Before starting review, confirm adherence to:
- FESTIVAL_RULES.md sections 2.1-2.3 (Code Standards)
- FESTIVAL_RULES.md section 4.2 (Security Requirements)
- COMMON_INTERFACES.md (Architecture compliance)
```

### Automation Opportunities

**Automated Testing:**

- CI/CD pipeline runs tests automatically
- Test results posted to pull requests
- Coverage reports generated automatically

**Automated Code Review:**

- Linting runs on every commit
- Security scanning integrated into workflow
- Dependency vulnerability checks

**Automated Quality Metrics:**

- Code quality scores tracked over time
- Test coverage trending
- Review cycle time measurement

## Common Anti-Patterns to Avoid

### ❌ Skip Quality Gates When Under Pressure

**Problem:** "We'll come back and test later"
**Reality:** Technical debt compounds, bugs become harder to fix
**Solution:** Streamline quality gates if needed, but never skip them entirely

### ❌ Generic Quality Gates for All Work

**Problem:** Same requirements for documentation and critical backend services
**Solution:** Customize requirements based on risk and impact

### ❌ Quality Gates as Checklist Theater  

**Problem:** Going through motions without finding real issues
**Solution:** Focus on value - what could actually go wrong?

### ❌ No Clear Acceptance Criteria

**Problem:** Quality gates drag on without clear completion criteria
**Solution:** Define specific, measurable acceptance criteria upfront

## Metrics and Improvement

### Track These Metrics

- **Defect Escape Rate**: Issues found after sequence completion
- **Review Cycle Time**: Average time from review start to acceptance
- **Rework Rate**: Percentage of sequences requiring significant iteration
- **Quality Gate Completion Rate**: Percentage of sequences with full quality gate completion

### Continuous Improvement

- Review quality gate effectiveness quarterly
- Adjust requirements based on defect patterns
- Streamline processes that don't add value
- Add rigor where quality issues persist

## Summary

Quality gates are not bureaucracy - they're insurance against technical debt and quality degradation. Customize them for your project's needs, but never skip them entirely. The three-task pattern (test → review → iterate) provides a consistent framework that scales from simple to complex projects while maintaining quality standards.

Remember: It's cheaper to find and fix issues during the sequence than after the festival is complete.
