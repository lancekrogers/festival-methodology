package markers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/markers"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ScaffoldMeta contains metadata about the scaffold output
type ScaffoldMeta struct {
	Template    string `json:"template" yaml:"template"`
	Source      string `json:"source" yaml:"source"`
	Generated   string `json:"generated" yaml:"generated"`
	MarkerCount int    `json:"marker_count" yaml:"marker_count"`
	FestVersion string `json:"fest_version" yaml:"fest_version"`
}

// ScaffoldOutput is the structure for scaffold JSON/YAML output
type ScaffoldOutput struct {
	Meta    ScaffoldMeta      `json:"_meta" yaml:"_meta"`
	Markers map[string]string `json:"markers" yaml:"markers"`
}

type scaffoldOptions struct {
	template  string
	file      string
	output    string
	format    string
	withHints bool
}

// templateAliases maps short names to template filenames
var templateAliases = map[string]string{
	"task":              "tasks/TASK.md",
	"task-simple":       "tasks/SIMPLE.md",
	"task-minimal":      "tasks/MINIMAL.md",
	"sequence":          "sequences/GOAL.md",
	"sequence-minimal":  "sequences/GOAL_MINIMAL.md",
	"phase-impl":        "phases/implementation/GOAL.md",
	"phase-planning":    "phases/planning/GOAL.md",
	"phase-research":    "phases/research/GOAL.md",
	"phase-review":      "phases/review/GOAL.md",
	"phase-action":      "phases/non_coding_action/GOAL.md",
	"festival":          "festival/GOAL.md",
	"festival-goal":     "festival/GOAL.md",
	"festival-overview": "festival/OVERVIEW.md",
	"festival-todo":     "festival/TODO.md",
	"festival-rules": "festival/RULES.md",
}

// newScaffoldCommand creates the scaffold subcommand
func newScaffoldCommand(opts *markersOptions) *cobra.Command {
	scaffoldOpts := &scaffoldOptions{}

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate marker JSON from template",
		Long: `Generate a JSON or YAML file with all template markers pre-populated as keys.

This allows agents to fill marker values without manually typing hint strings,
eliminating typos and reducing token usage.

Examples:
  # Generate from built-in template
  fest markers scaffold --template task-simple

  # Generate from existing file
  fest markers scaffold --file PHASE_GOAL.md

  # Output as YAML to a file
  fest markers scaffold --template sequence --format yaml --output markers.yaml

Available template aliases:
  task, task-simple, task-minimal    Task templates
  sequence, sequence-minimal         Sequence templates
  phase, phase-impl, phase-planning  Phase templates
  festival, festival-overview        Festival templates
  gate-testing, gate-review          Quality gate templates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScaffold(scaffoldOpts, opts)
		},
	}

	cmd.Flags().StringVar(&scaffoldOpts.template, "template", "", "Built-in template name (e.g., task, phase, sequence)")
	cmd.Flags().StringVar(&scaffoldOpts.file, "file", "", "Path to file with markers")
	cmd.Flags().StringVar(&scaffoldOpts.output, "output", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&scaffoldOpts.format, "format", "json", "Output format: json or yaml")
	cmd.Flags().BoolVar(&scaffoldOpts.withHints, "with-hints", false, "Include hint descriptions as comments")

	return cmd
}

func runScaffold(scaffoldOpts *scaffoldOptions, opts *markersOptions) error {
	// Validate inputs
	if scaffoldOpts.template == "" && scaffoldOpts.file == "" {
		return errors.Validation("must specify either --template or --file")
	}
	if scaffoldOpts.template != "" && scaffoldOpts.file != "" {
		return errors.Validation("cannot specify both --template and --file")
	}

	// Validate format
	format := strings.ToLower(scaffoldOpts.format)
	if format != "json" && format != "yaml" {
		return errors.Validation("format must be 'json' or 'yaml'")
	}

	// Get source file path and content
	var sourcePath string
	var content []byte
	var err error

	if scaffoldOpts.template != "" {
		sourcePath, content, err = resolveTemplate(scaffoldOpts.template)
		if err != nil {
			return err
		}
	} else {
		sourcePath = scaffoldOpts.file
		content, err = os.ReadFile(scaffoldOpts.file)
		if err != nil {
			return errors.Wrap(err, "reading file").WithField("path", scaffoldOpts.file)
		}
	}

	// Extract markers
	parsedMarkers := markers.Parse(string(content))
	hints := markers.ExtractHints(parsedMarkers)

	// Deduplicate hints while preserving order
	uniqueHints := deduplicateHints(hints)

	// Build output structure
	output := ScaffoldOutput{
		Meta: ScaffoldMeta{
			Template:    scaffoldOpts.template,
			Source:      sourcePath,
			Generated:   time.Now().UTC().Format(time.RFC3339),
			MarkerCount: len(uniqueHints),
			FestVersion: "1.0.0",
		},
		Markers: make(map[string]string),
	}

	for _, hint := range uniqueHints {
		output.Markers[hint] = ""
	}

	// Generate output
	var outputBytes []byte
	if format == "yaml" {
		outputBytes, err = yaml.Marshal(output)
	} else {
		outputBytes, err = json.MarshalIndent(output, "", "  ")
	}
	if err != nil {
		return errors.Wrap(err, "marshaling output")
	}

	// Write output
	if scaffoldOpts.output != "" {
		if err := os.WriteFile(scaffoldOpts.output, outputBytes, 0644); err != nil {
			return errors.Wrap(err, "writing output file").WithField("path", scaffoldOpts.output)
		}
		if !opts.jsonOutput {
			display := ui.New(false, false)
			display.Success("Scaffold written to %s (%d markers)", scaffoldOpts.output, len(uniqueHints))
		}
	} else {
		fmt.Println(string(outputBytes))
	}

	return nil
}

// resolveTemplate finds a template by alias or filename
func resolveTemplate(name string) (string, []byte, error) {
	// Check if it's an alias
	templateName := name
	if alias, ok := templateAliases[strings.ToLower(name)]; ok {
		templateName = alias
	}

	// Try to find the template starting from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil, errors.Wrap(err, "getting working directory")
	}

	templateRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return "", nil, errors.Wrap(err, "finding template root")
	}

	templatePath := filepath.Join(templateRoot, templateName)

	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try without .md extension
		if !strings.HasSuffix(templateName, ".md") {
			templatePath = filepath.Join(templateRoot, templateName+".md")
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				return "", nil, suggestTemplate(name)
			}
		} else {
			return "", nil, suggestTemplate(name)
		}
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", nil, errors.Wrap(err, "reading template").WithField("path", templatePath)
	}

	return templatePath, content, nil
}

// suggestTemplate returns an error with suggestions for similar templates
func suggestTemplate(name string) error {
	suggestions := []string{}
	nameLower := strings.ToLower(name)

	for alias := range templateAliases {
		if strings.Contains(alias, nameLower) || strings.Contains(nameLower, alias) {
			suggestions = append(suggestions, alias)
		}
	}

	if len(suggestions) > 0 {
		return errors.NotFound("template").
			WithField("name", name).
			WithField("suggestions", strings.Join(suggestions, ", "))
	}

	// List all available aliases
	available := make([]string, 0, len(templateAliases))
	for alias := range templateAliases {
		available = append(available, alias)
	}

	return errors.NotFound("template").
		WithField("name", name).
		WithField("available", strings.Join(available, ", "))
}

// deduplicateHints removes duplicate hints while preserving order
func deduplicateHints(hints []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(hints))

	for _, hint := range hints {
		if !seen[hint] {
			seen[hint] = true
			result = append(result, hint)
		}
	}

	return result
}
