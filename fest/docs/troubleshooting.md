# Troubleshooting

Common issues and fixes for Fest CLI output and styling.

## Styled Output and Color

### Output has no color in a TTY

Possible causes:
- `--no-color` flag
- `NO_COLOR=1` environment variable
- `TERM=dumb` or a terminal without color support
- Output is not actually a TTY (piped or redirected)

Fixes:
- Remove `--no-color`
- Unset `NO_COLOR`
- Use a terminal with color support
- Run without piping or redirection

### ANSI codes appear in logs or redirected output

Possible causes:
- `CLICOLOR_FORCE=1` environment variable
- Output forced through a pseudo-TTY

Fixes:
- Unset `CLICOLOR_FORCE`
- Use `--no-color` or `NO_COLOR=1`

### JSON output is not plain

Fixes:
- Use the command's `--json` flag
- Avoid piping styled output into parsers

## Output Performance

If output feels slow on large festivals:
- Use narrower commands (`fest status`, `fest next`) instead of broad listings
- Use `--json` for machine parsing and filter with jq
- Prefer non-verbose output unless needed

## What to Report

When filing issues, include:
- Command and flags used
- Whether stdout is a TTY or redirected
- Relevant environment variables (`NO_COLOR`, `CLICOLOR_FORCE`, `TERM`)
- Sample output (redacted if needed)
