package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

// ResearchDocType represents the type of research document
type ResearchDocType string

const (
	ResearchInvestigation ResearchDocType = "investigation"
	ResearchComparison    ResearchDocType = "comparison"
	ResearchAnalysis      ResearchDocType = "analysis"
	ResearchSpecification ResearchDocType = "specification"
)

var validResearchTypes = []string{
	string(ResearchInvestigation),
	string(ResearchComparison),
	string(ResearchAnalysis),
	string(ResearchSpecification),
}

func newResearchCreateCmd() *cobra.Command {
	var docType, title, path string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new research document from template",
		Long: `Create a new research document using one of the research templates.

Available document types:
  investigation  - Exploring unknowns, gathering information
  comparison     - Evaluating options, making decisions
  analysis       - Deep-dive technical analysis
  specification  - Defining requirements and design`,
		Example: `  fest research create --type investigation --title "API Authentication Options"
  fest research create --type comparison --title "Database Selection"
  fest research create --type analysis --title "Performance Baseline"
  fest research create --type specification --title "User API Design"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResearchCreate(cmd.Context(), cmd, docType, title, path, jsonOutput)
		},
	}

	cmd.Flags().StringVarP(&docType, "type", "t", "", "Document type (investigation|comparison|analysis|specification)")
	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.Flags().StringVarP(&path, "path", "p", ".", "Destination directory")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runResearchCreate(ctx context.Context, cmd *cobra.Command, docType, title, path string, jsonOutput bool) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Validate inputs
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("--title is required")
	}
	if strings.TrimSpace(docType) == "" {
		return fmt.Errorf("--type is required (investigation|comparison|analysis|specification)")
	}

	// Validate document type
	validType := false
	for _, t := range validResearchTypes {
		if strings.ToLower(docType) == t {
			validType = true
			docType = t
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid document type %q: must be one of %v", docType, validResearchTypes)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return fmt.Errorf("finding template root: %w", err)
	}

	// Resolve destination path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Determine template file
	templateMap := map[string]string{
		"investigation": "RESEARCH_INVESTIGATION_TEMPLATE.md",
		"comparison":    "RESEARCH_COMPARISON_TEMPLATE.md",
		"analysis":      "RESEARCH_ANALYSIS_TEMPLATE.md",
		"specification": "RESEARCH_SPECIFICATION_TEMPLATE.md",
	}

	templateFile := templateMap[docType]
	templatePath := filepath.Join(tmplRoot, templateFile)

	// Load template
	loader := tpl.NewLoader()
	t, err := loader.Load(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", templateFile, err)
	}

	// Build context
	researchCtx := tpl.NewContext()

	// Try to detect phase info from path
	phaseID := detectPhaseID(absPath)
	if phaseID != "" {
		researchCtx.PhaseID = phaseID
	}

	// Set custom variables
	researchCtx.SetCustom("title", title)
	researchCtx.SetCustom("research_type", docType)
	researchCtx.SetCustom("created", time.Now().Format(time.RFC3339))
	researchCtx.SetCustom("research_id", slugify(title))

	// Render template
	mgr := tpl.NewManager()
	content, err := mgr.Render(t, researchCtx)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create output file
	slug := slugify(title)
	outputFile := fmt.Sprintf("%s_%s.md", docType, slug)
	outputPath := filepath.Join(absPath, outputFile)

	// Ensure directory exists
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if jsonOutput {
		output := struct {
			OK       bool   `json:"ok"`
			Action   string `json:"action"`
			Type     string `json:"type"`
			Title    string `json:"title"`
			Path     string `json:"path"`
			Filename string `json:"filename"`
		}{
			OK:       true,
			Action:   "research_create",
			Type:     docType,
			Title:    title,
			Path:     outputPath,
			Filename: outputFile,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Created research document: %s\n", outputFile)
	fmt.Fprintf(out, "  Type: %s\n", docType)
	fmt.Fprintf(out, "  Path: %s\n", outputPath)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintln(out, "  1. Open the document and fill in the [REPLACE: ...] markers")
	fmt.Fprintln(out, "  2. Link to implementation phases with 'fest research link'")

	return nil
}

// detectPhaseID tries to detect the phase ID from the current path
func detectPhaseID(path string) string {
	// Walk up the path looking for a directory matching NNN_NAME pattern
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if len(parts[i]) >= 4 && parts[i][0:3] >= "000" && parts[i][0:3] <= "999" && parts[i][3] == '_' {
			return parts[i]
		}
	}
	return ""
}

// slugify converts a title to a filename-safe slug
func slugify(title string) string {
	// Lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with underscores
	slug = strings.ReplaceAll(slug, " ", "_")
	slug = strings.ReplaceAll(slug, "-", "_")

	// Remove or replace other special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	// Collapse multiple underscores
	slug = result.String()
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}

	// Trim leading/trailing underscores
	slug = strings.Trim(slug, "_")

	return slug
}
