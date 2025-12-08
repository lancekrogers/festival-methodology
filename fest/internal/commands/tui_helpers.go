package commands

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"

    tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

func collectRequiredVars(templateRoot string, paths []string) []string {
    loader := tpl.NewLoader()
    vars := []string{}
    for _, p := range paths {
        if _, err := os.Stat(p); err != nil {
            continue
        }
        t, err := loader.Load(p)
        if err != nil || t.Metadata == nil {
            continue
        }
        vars = append(vars, t.Metadata.RequiredVariables...)
    }
    return vars
}

func uniqueStrings(in []string) []string {
    m := map[string]struct{}{}
    out := []string{}
    for _, s := range in {
        if s == "" {
            continue
        }
        if _, ok := m[s]; !ok {
            m[s] = struct{}{}
            out = append(out, s)
        }
    }
    return out
}

func writeTempVarsFile(vars map[string]interface{}) (string, error) {
    if len(vars) == 0 {
        return "", nil
    }
    f, err := os.CreateTemp("", "fest-vars-*.json")
    if err != nil {
        return "", fmt.Errorf("failed to create temp vars file: %w", err)
    }
    enc := json.NewEncoder(f)
    if err := enc.Encode(vars); err != nil {
        _ = f.Close()
        return "", fmt.Errorf("failed to write temp vars file: %w", err)
    }
    _ = f.Close()
    return f.Name(), nil
}

func atoiDefault(s string, def int) int {
    n, err := strconv.Atoi(strings.TrimSpace(s))
    if err != nil {
        return def
    }
    return n
}

func defaultFestivalTemplatePaths(tmplRoot string) []string {
    return []string{
        filepath.Join(tmplRoot, "FESTIVAL_OVERVIEW_TEMPLATE.md"),
        filepath.Join(tmplRoot, "FESTIVAL_GOAL_TEMPLATE.md"),
        filepath.Join(tmplRoot, "FESTIVAL_RULES_TEMPLATE.md"),
        filepath.Join(tmplRoot, "FESTIVAL_TODO_TEMPLATE.md"),
    }
}

// slugify mirrors the create_festival.go behavior
func slugify(s string) string {
    lower := strings.ToLower(strings.TrimSpace(s))
    re := regexp.MustCompile(`[^a-z0-9]+`)
    slug := re.ReplaceAllString(lower, "-")
    slug = strings.Trim(slug, "-")
    if slug == "" {
        slug = "festival"
    }
    return slug
}

// resolvePhaseDirInput attempts to resolve a user-provided phase path shortcut
// like "02" or "002" or a phase name to a concrete directory path.
// It searches under the detected festival directory near cwd.
func resolvePhaseDirInput(input, cwd string) (string, error) {
    input = strings.TrimSpace(input)
    absCwd, _ := filepath.Abs(cwd)
    festivalDir := findFestivalDir(absCwd)

    // If direct path exists (relative or absolute), use it
    if input == "" || input == "." {
        if isPhaseDirPath(absCwd) {
            return absCwd, nil
        }
        // No specific phase; default to CWD if it looks like a festival directory
        return "", fmt.Errorf("please specify a phase (e.g., 002_IMPLEMENT or 02)")
    }
    if filepath.IsAbs(input) {
        if info, err := os.Stat(input); err == nil && info.IsDir() {
            return input, nil
        }
    } else {
        try := filepath.Join(absCwd, input)
        if info, err := os.Stat(try); err == nil && info.IsDir() {
            return try, nil
        }
    }

    // Collect phase dirs under festivalDir
    phases := listPhaseDirs(festivalDir)
    if len(phases) == 0 {
        return "", fmt.Errorf("no phase directories found under %s", festivalDir)
    }

    // If numeric, pad to 3 digits and match prefix
    if isDigits(input) {
        n, _ := strconv.Atoi(input)
        code := fmt.Sprintf("%03d", n)
        for _, name := range phases {
            if strings.HasPrefix(name, code+"_") || name == code {
                return filepath.Join(festivalDir, name), nil
            }
        }
        return "", fmt.Errorf("phase %s not found under %s", code, festivalDir)
    }

    // Match by name suffix after underscore (case-insensitive)
    needle := strings.ToLower(input)
    for _, name := range phases {
        if name == input {
            return filepath.Join(festivalDir, name), nil
        }
        parts := strings.SplitN(name, "_", 2)
        if len(parts) == 2 {
            if strings.ToLower(parts[1]) == needle {
                return filepath.Join(festivalDir, name), nil
            }
        }
    }

    return "", fmt.Errorf("could not resolve phase '%s'", input)
}

func isDigits(s string) bool {
    if s == "" { return false }
    for _, r := range s {
        if r < '0' || r > '9' { return false }
    }
    return true
}

// findFestivalDir attempts to find the nearest festival directory from cwd.
// If cwd is a phase dir (NNN_NAME), returns its parent. Otherwise, returns cwd
// if it looks like a festival root (has phase dirs or key festival files). Fallback: cwd.
func findFestivalDir(cwd string) string {
    if isPhaseDirPath(cwd) {
        return filepath.Dir(cwd)
    }
    if looksLikeFestivalDir(cwd) {
        return cwd
    }
    // Fallback one level up
    parent := filepath.Dir(cwd)
    if looksLikeFestivalDir(parent) {
        return parent
    }
    return cwd
}

func looksLikeFestivalDir(dir string) bool {
    // If typical files exist or numbered phase dirs present, assume festival dir
    if exists(filepath.Join(dir, "FESTIVAL_OVERVIEW.md")) || exists(filepath.Join(dir, "FESTIVAL_GOAL.md")) {
        return true
    }
    return len(listPhaseDirs(dir)) > 0
}

func isPhaseDirPath(path string) bool {
    base := filepath.Base(path)
    re := regexp.MustCompile(`^[0-9]{3}_.+`)
    return re.MatchString(base)
}

func listPhaseDirs(dir string) []string {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil
    }
    re := regexp.MustCompile(`^[0-9]{3}_.+`)
    out := []string{}
    for _, e := range entries {
        if e.IsDir() && re.MatchString(e.Name()) {
            out = append(out, e.Name())
        }
    }
    return out
}

func exists(p string) bool {
    _, err := os.Stat(p)
    return err == nil
}
