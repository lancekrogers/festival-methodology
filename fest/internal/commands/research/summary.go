package research

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ResearchDocInfo represents information about a research document
type ResearchDocInfo struct {
	Name            string   `json:"name" yaml:"name"`
	Path            string   `json:"path" yaml:"path"`
	Type            string   `json:"type" yaml:"type"`
	Title           string   `json:"title,omitempty" yaml:"title,omitempty"`
	Status          string   `json:"status,omitempty" yaml:"status,omitempty"`
	LinkedPhases    []string `json:"linked_phases,omitempty" yaml:"linked_phases,omitempty"`
	LinkedSequences []string `json:"linked_sequences,omitempty" yaml:"linked_sequences,omitempty"`
	LinkedTasks     []string `json:"linked_tasks,omitempty" yaml:"linked_tasks,omitempty"`
}

// ResearchSummary contains the summary of research documents
type ResearchSummary struct {
	GeneratedAt time.Time         `json:"generated_at" yaml:"generated_at"`
	Scope       string            `json:"scope" yaml:"scope"`
	ScopeID     string            `json:"scope_id" yaml:"scope_id"`
	Total       int               `json:"total" yaml:"total"`
	ByType      map[string]int    `json:"by_type" yaml:"by_type"`
	ByStatus    map[string]int    `json:"by_status" yaml:"by_status"`
	Documents   []ResearchDocInfo `json:"documents" yaml:"documents"`
}

func newResearchSummaryCmd() *cobra.Command {
	var phase string
	var festival bool
	var output string
	var format string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Generate summary/index of research documents",
		Long: `Generate a summary of all research documents in a research phase or festival.

The summary includes document counts by type and status, and a list of all
research documents with their metadata.`,
		Example: `  fest research summary
  fest research summary --phase 001_RESEARCH
  fest research summary --festival
  fest research summary --format json --output research_index.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResearchSummary(cmd.Context(), cmd, phase, festival, output, format, jsonOutput)
		},
	}

	cmd.Flags().StringVarP(&phase, "phase", "p", "", "Phase to summarize")
	cmd.Flags().BoolVarP(&festival, "festival", "f", false, "Summarize entire festival")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write to file (default: stdout)")
	cmd.Flags().StringVar(&format, "format", "markdown", "Output format (markdown|json)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (shorthand for --format json)")

	return cmd
}

func runResearchSummary(ctx context.Context, cmd *cobra.Command, phase string, festival bool, output, format string, jsonOutput bool) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("runResearchSummary")
	}

	if jsonOutput {
		format = "json"
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting working directory", err)
	}

	// Find festivals root
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals root")
	}

	// Determine scope
	var searchPath string
	var scope, scopeID string

	if festival {
		// Get current festival
		festivalPath := detectFestivalPath(cwd, festivalsRoot)
		if festivalPath == "" {
			return errors.NotFound("festival directory")
		}
		searchPath = festivalPath
		scope = "festival"
		scopeID = filepath.Base(festivalPath)
	} else if phase != "" {
		festivalPath := detectFestivalPath(cwd, festivalsRoot)
		if festivalPath == "" {
			return errors.NotFound("festival directory")
		}
		searchPath = filepath.Join(festivalPath, phase)
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			return errors.NotFound("phase directory").WithField("phase", phase)
		}
		scope = "phase"
		scopeID = phase
	} else {
		// Use current directory or detect phase
		phaseID := detectPhaseID(cwd)
		if phaseID != "" {
			searchPath = cwd
			scope = "phase"
			scopeID = phaseID
		} else {
			// Try to detect festival
			festivalPath := detectFestivalPath(cwd, festivalsRoot)
			if festivalPath == "" {
				return errors.NotFound("festival or phase directory")
			}
			searchPath = festivalPath
			scope = "festival"
			scopeID = filepath.Base(festivalPath)
		}
	}

	// Collect research documents
	docs, err := collectResearchDocs(searchPath)
	if err != nil {
		return errors.Wrap(err, "collecting research documents")
	}

	// Build summary
	summary := ResearchSummary{
		GeneratedAt: time.Now(),
		Scope:       scope,
		ScopeID:     scopeID,
		Total:       len(docs),
		ByType:      make(map[string]int),
		ByStatus:    make(map[string]int),
		Documents:   docs,
	}

	for _, doc := range docs {
		summary.ByType[doc.Type]++
		if doc.Status != "" {
			summary.ByStatus[doc.Status]++
		}
	}

	// Generate output
	var result string
	if format == "json" {
		var buffer bytes.Buffer
		if err := shared.EncodeJSON(&buffer, summary); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
		result = strings.TrimRight(buffer.String(), "\n")
	} else {
		result = formatMarkdownSummary(summary)
	}

	// Write output
	if output != "" {
		if err := os.WriteFile(output, []byte(result), 0644); err != nil {
			return errors.IO("writing output file", err).WithField("path", output)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ui.Success("Summary written"), ui.Dim(output))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), result)
	}

	return nil
}

func collectResearchDocs(searchPath string) ([]ResearchDocInfo, error) {
	var docs []ResearchDocInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Skip goal files
		name := info.Name()
		if name == "PHASE_GOAL.md" || name == "SEQUENCE_GOAL.md" || strings.HasPrefix(name, "RESEARCH_") && strings.HasSuffix(name, "_TEMPLATE.md") {
			return nil
		}

		// Check if it's a research document by prefix
		docType := ""
		for _, t := range validResearchTypes {
			if strings.HasPrefix(name, t+"_") {
				docType = t
				break
			}
		}

		if docType == "" {
			return nil // Not a research document
		}

		// Parse frontmatter for additional info
		doc := ResearchDocInfo{
			Name: name,
			Path: path,
			Type: docType,
		}

		// Try to read frontmatter
		content, err := os.ReadFile(path)
		if err == nil {
			frontmatter := extractFrontmatter(string(content))
			if frontmatter != nil {
				if title, ok := frontmatter["title"].(string); ok {
					doc.Title = title
				}
				if status, ok := frontmatter["status"].(string); ok {
					doc.Status = status
				}
				if phases, ok := frontmatter["linked_phases"].([]interface{}); ok {
					for _, p := range phases {
						if s, ok := p.(string); ok {
							doc.LinkedPhases = append(doc.LinkedPhases, s)
						}
					}
				}
				if seqs, ok := frontmatter["linked_sequences"].([]interface{}); ok {
					for _, s := range seqs {
						if str, ok := s.(string); ok {
							doc.LinkedSequences = append(doc.LinkedSequences, str)
						}
					}
				}
				if tasks, ok := frontmatter["linked_tasks"].([]interface{}); ok {
					for _, t := range tasks {
						if str, ok := t.(string); ok {
							doc.LinkedTasks = append(doc.LinkedTasks, str)
						}
					}
				}
			}
		}

		docs = append(docs, doc)
		return nil
	})

	return docs, err
}

func extractFrontmatter(content string) map[string]interface{} {
	// Look for YAML frontmatter between --- markers
	if !strings.HasPrefix(content, "---") {
		return nil
	}

	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return nil
	}

	frontmatterStr := content[3 : 3+endIndex]

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return nil
	}

	return frontmatter
}

func formatMarkdownSummary(summary ResearchSummary) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Research Summary: %s\n\n", summary.ScopeID))
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", summary.GeneratedAt.Format(time.RFC3339)))

	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Type | Count |\n")
	sb.WriteString("|------|-------|\n")
	for _, t := range validResearchTypes {
		count := summary.ByType[t]
		if count > 0 {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", t, count))
		}
	}
	sb.WriteString(fmt.Sprintf("| **Total** | **%d** |\n", summary.Total))
	sb.WriteString("\n")

	if len(summary.ByStatus) > 0 {
		sb.WriteString("### By Status\n\n")
		sb.WriteString("| Status | Count |\n")
		sb.WriteString("|--------|-------|\n")
		for status, count := range summary.ByStatus {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", status, count))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Documents\n\n")

	// Group by type
	for _, docType := range validResearchTypes {
		var typeDocs []ResearchDocInfo
		for _, doc := range summary.Documents {
			if doc.Type == docType {
				typeDocs = append(typeDocs, doc)
			}
		}

		if len(typeDocs) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(docType)))
		for _, doc := range typeDocs {
			title := doc.Title
			if title == "" {
				title = doc.Name
			}
			status := doc.Status
			if status == "" {
				status = "unknown"
			}
			sb.WriteString(fmt.Sprintf("- **[%s](%s)** (%s)\n", title, doc.Path, status))

			if len(doc.LinkedPhases) > 0 {
				sb.WriteString(fmt.Sprintf("  - Linked to: %s\n", strings.Join(doc.LinkedPhases, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func detectFestivalPath(cwd, festivalsRoot string) string {
	// Look for active, planned, or any festival directory
	relPath, err := filepath.Rel(festivalsRoot, cwd)
	if err != nil {
		return ""
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) < 2 {
		return ""
	}

	// First part should be status (active, planned, completed, archived)
	// Second part should be festival name
	status := parts[0]
	if status != "active" && status != "planned" && status != "completed" && status != "archived" {
		return ""
	}

	return filepath.Join(festivalsRoot, parts[0], parts[1])
}
