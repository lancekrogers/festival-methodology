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
