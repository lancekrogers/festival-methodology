# Error Audit Report

## Executive Summary

This report provides a comprehensive analysis of error messages across the fest CLI codebase.
The audit evaluates message clarity, consistency, and actionability to improve user experience.

### Statistics

| Category | Count | Percentage |
|----------|-------|------------|
| NotFound | 121 | 12.9% |
| Validation | 133 | 14.1% |
| IO | 215 | 22.9% |
| Parse | 30 | 3.2% |
| New | 17 | 1.8% |
| Wrap | 425 | 45.1% |
| **Total** | **941** | **100%** |

### Overall Assessment

- **Strengths**: Consistent use of structured errors, good field annotations
- **Improvements Made**: Default hints now wired into all error helpers
- **Remaining Opportunities**: Some messages could be more actionable

## Error Categories

### NOT_FOUND Errors

| Message Pattern | Clarity | Priority | Notes |
|-----------------|---------|----------|-------|
| `festival not found` | 5/5 | Low | Clear, has default hint |
| `phase not found` | 5/5 | Low | Clear, has default hint |
| `sequence not found` | 5/5 | Low | Clear, has default hint |
| `festivals root not found` | 4/5 | Medium | Could specify expected location |
| `task not found` | 5/5 | Low | Clear with task ID field |
| `policy not found` | 4/5 | Medium | Technical term, may confuse users |
| `config repo not found` | 4/5 | Low | Has field annotation |
| `source path not found` | 5/5 | Low | Clear with path field |

**Recommendations:**
- All NotFound errors now include context-aware hints via `hintForResource()`
- Consider adding more specific hints for technical resources like "policy"

### VALIDATION Errors

| Message Pattern | Clarity | Priority | Notes |
|-----------------|---------|----------|-------|
| `festivalsRoot cannot be empty` | 5/5 | Low | Clear parameter validation |
| `task ID required` | 5/5 | Low | Clear requirement |
| `task path is outside festival` | 5/5 | Low | Clear boundary violation |
| `no progress data to save` | 4/5 | Medium | Could suggest how to create data |
| `progress must be between 0 and 100` | 5/5 | Low | Clear range constraint |
| `blocker message required` | 5/5 | Low | Clear requirement |
| `config repo already exists` | 5/5 | Low | Clear duplicate prevention |
| `source is not a directory` | 5/5 | Low | Clear type constraint |
| `path is not a directory` | 5/5 | Low | Clear with path field |

**Recommendations:**
- Validation errors now include default `HintSeeHelp` hint
- Messages are generally clear and actionable
- Consider using typo detection for status/entity type validation

### IO Errors

| Message Pattern | Clarity | Priority | Notes |
|-----------------|---------|----------|-------|
| `I/O operation failed` | 3/5 | High | Generic, wrapped error provides details |
| Various operation-specific | 4/5 | Medium | Operation name provides context |

**Recommendations:**
- IO errors now include `HintCheckPermissions` by default
- The operation name (Op field) provides good context
- Consider more specific hints for common failures (disk full, locked files)

### PARSE Errors

| Message Pattern | Clarity | Priority | Notes |
|-----------------|---------|----------|-------|
| `failed to parse YAML` | 4/5 | Medium | Could include line number |
| `invalid frontmatter` | 4/5 | Medium | Technical term |
| `failed to parse JSON` | 4/5 | Medium | Could include position |

**Recommendations:**
- Parse errors now include `HintCheckConfig` by default
- Consider extracting line numbers from underlying parse errors
- Could add syntax highlighting or snippet context

## New Features Implemented

### 1. Default Hints (Task 02)

All error helper functions now include appropriate default hints:

```go
NotFound("festival")  // Includes HintFestivalNotFound
Validation(msg)       // Includes HintSeeHelp
IO(op, err)           // Includes HintCheckPermissions
Config(msg)           // Includes HintCheckConfig
Template(msg)         // Includes HintCheckTemplate
Parse(msg, err)       // Includes HintCheckConfig
```

### 2. Typo Detection (Task 03)

New `suggestions.go` provides:

- `LevenshteinDistance()` for edit distance calculation
- `SuggestSimilar()` for finding similar strings
- `DidYouMean()` for creating suggestion errors
- `ValidateWithSuggestions()` for validating against known values

Example usage:
```go
err := errors.ValidateWithSuggestions(status, ValidStatuses, "status")
// Returns: invalid status: "actve"
// Hint: Did you mean "active"?
```

### 3. Standard Hint Constants

Available hints for common scenarios:

| Constant | Message |
|----------|---------|
| `HintFestivalNotFound` | Navigate to a festival directory or run 'fest show all' |
| `HintPhaseNotFound` | Run 'fest status list --type phase' |
| `HintSequenceNotFound` | Run 'fest status list --type sequence' |
| `HintCreateFestival` | Run 'fest create festival' or 'fest tui' |
| `HintCheckPath` | Check the path and try again |
| `HintCheckConfig` | Check your fest.yaml configuration |
| `HintCheckTemplate` | Run 'fest validate' to check for template issues |
| `HintRunInit` | Run 'fest init' to initialize a festival workspace |
| `HintCheckPermissions` | Check file/directory permissions |
| `HintSeeHelp` | Run 'fest help' for more information |

## Priority Matrix

### High Priority (Should Address)

| Issue | Recommendation | Effort |
|-------|----------------|--------|
| Generic IO message | Add operation-specific context | Low |
| Parse errors lack position | Extract line/column from underlying errors | Medium |

### Medium Priority (Nice to Have)

| Issue | Recommendation | Effort |
|-------|----------------|--------|
| Technical terms (policy, frontmatter) | Add glossary or simpler explanations | Low |
| Missing progress data message | Suggest how to track progress | Low |

### Low Priority (Polish)

| Issue | Recommendation | Effort |
|-------|----------------|--------|
| Status validation | Wire in typo detection | Low |
| Entity type validation | Wire in typo detection | Low |

## Testing Coverage

Error handling tests exist in:

- `internal/errors/errors_test.go` - Core error types
- `internal/errors/suggestions_test.go` - Typo detection

All tests pass with the new hint system.

## Conclusion

The fest CLI error system is well-structured with:

1. Consistent use of error codes for categorization
2. Field annotations for context
3. JSON-serializable errors for programmatic handling
4. Default hints for all error types
5. Typo detection for improved UX

The improvements made in this festival (default hints, typo detection, suggestions)
significantly enhance the actionability of error messages.
