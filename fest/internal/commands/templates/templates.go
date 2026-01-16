package templates

import (
	"context"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewTemplatesCommand creates the templates command group
func NewTemplatesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage agent-created templates within a festival",
		Long: `Create, apply, and manage templates for repetitive tasks.

Agent templates use simple {{variable}} syntax for substitution.
Templates are stored in .templates/ within the festival directory.

Examples:
  fest templates create component_test
  fest templates apply component_test --vars '{"name": "UserService"}'
  fest templates list`,
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newApplyCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}

func newCreateCmd() *cobra.Command {
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new template",
		Long: `Create a new agent template in the current festival.

The template will be stored in .templates/<name>.md

Example template content:
  # {{component_name}} Test

  ## Setup
  {{setup_steps}}

  ## Test Cases
  - {{test_case_1}}
  - {{test_case_2}}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd.Context(), args[0], fromFile)
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "create from existing file")

	return cmd
}

func runCreate(ctx context.Context, name, fromFile string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := template.NewAgentTemplateStore(festivalPath)

	var content string
	if fromFile != "" {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			return errors.IO("reading source file", err).WithField("path", fromFile)
		}
		content = string(data)
	} else {
		// Create a starter template
		content = fmt.Sprintf(`# Template: %s

## Description
[Describe what this template is for]

## Content
{{variable_1}}

{{variable_2}}
`, name)
	}

	tmpl, err := store.Save(ctx, name, content)
	if err != nil {
		return err
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	display.Success("Created template: %s", tmpl.Path)

	if len(tmpl.Variables) > 0 {
		display.Info("Variables: %v", tmpl.Variables)
	} else {
		display.Info("No {{variables}} found. Edit the template to add variables.")
	}

	return nil
}

func newApplyCmd() *cobra.Command {
	var (
		varsJSON string
		output   string
		preview  bool
	)

	cmd := &cobra.Command{
		Use:   "apply <name>",
		Short: "Apply a template with variables",
		Long: `Apply a template, substituting {{variables}} with provided values.

Variables can be provided as:
  - JSON string: --vars '{"name": "value"}'
  - File reference: --vars @variables.json

Examples:
  fest templates apply component_test --vars '{"name": "UserService"}'
  fest templates apply api_endpoint -o ./api.md --vars @vars.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd.Context(), args[0], varsJSON, output, preview)
		},
	}

	cmd.Flags().StringVar(&varsJSON, "vars", "{}", "JSON object or @file with variable values")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: stdout)")
	cmd.Flags().BoolVar(&preview, "preview", false, "show result without writing")

	return cmd
}

func runApply(ctx context.Context, name, varsJSON, output string, preview bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := template.NewAgentTemplateStore(festivalPath)

	// Load template
	tmpl, err := store.Load(ctx, name)
	if err != nil {
		return err
	}

	// Parse variables
	vars, err := template.ParseVariablesJSON(varsJSON)
	if err != nil {
		return err
	}

	// Apply variables
	result := template.ApplyVariables(tmpl.Content, vars)

	// Check for unfilled variables
	remaining := template.ExtractVariables(result)
	if len(remaining) > 0 && !preview {
		display := ui.New(shared.IsNoColor(), shared.IsVerbose())
		display.Warning("Unfilled variables: %v", remaining)
	}

	// Output result
	if preview || output == "" {
		fmt.Println(result)
		return nil
	}

	if err := os.WriteFile(output, []byte(result), 0644); err != nil {
		return errors.IO("writing output file", err).WithField("path", output)
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	display.Success("Created: %s", output)
	return nil
}

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Long: `List all agent templates available in the current festival.

Templates are stored in .templates/ within the festival directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func runList(ctx context.Context, jsonOutput bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := template.NewAgentTemplateStore(festivalPath)
	templates, err := store.List(ctx)
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			display := ui.New(shared.IsNoColor(), shared.IsVerbose())
			display.Info("No templates found. Create one with 'fest templates create <name>'")
		}
		return nil
	}

	if jsonOutput {
		fmt.Print("[")
		for i, tmpl := range templates {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf(`{"name": "%s", "path": "%s", "variables": %q}`,
				tmpl.Name, tmpl.Path, tmpl.Variables)
		}
		fmt.Println("]")
		return nil
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	fmt.Printf("Templates in %s:\n\n", festivalPath)
	for _, tmpl := range templates {
		fmt.Printf("  %s\n", tmpl.Name)
		if len(tmpl.Variables) > 0 {
			fmt.Printf("    Variables: %v\n", tmpl.Variables)
		}
	}
	_ = display // silence unused warning
	return nil
}
