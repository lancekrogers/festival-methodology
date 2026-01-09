package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newResearchLinkCmd() *cobra.Command {
	var phases, sequences, tasks []string
	var bidirectional bool
	var unlink bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "link <research-doc>",
		Short: "Link research findings to implementation phases/tasks",
		Long: `Link a research document to phases, sequences, or tasks.

This creates references in the research document's frontmatter that indicate
which implementation work is informed by this research. With --bidirectional,
it also adds a reference in the target documents.`,
		Example: `  fest research link api-auth.md --phase 002_IMPLEMENT
  fest research link db-choice.md --sequence 002_IMPLEMENT/01_core
  fest research link spec.md --task 002_IMPLEMENT/01_core/03_design.md --bidirectional`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResearchLink(cmd.Context(), cmd, args[0], phases, sequences, tasks, bidirectional, unlink, jsonOutput)
		},
	}

	cmd.Flags().StringSliceVar(&phases, "phase", nil, "Phase to link (can be repeated)")
	cmd.Flags().StringSliceVar(&sequences, "sequence", nil, "Sequence to link (can be repeated)")
	cmd.Flags().StringSliceVar(&tasks, "task", nil, "Task to link (can be repeated)")
	cmd.Flags().BoolVarP(&bidirectional, "bidirectional", "b", false, "Create bidirectional references")
	cmd.Flags().BoolVarP(&unlink, "unlink", "u", false, "Remove the specified links")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runResearchLink(ctx context.Context, cmd *cobra.Command, docPath string, phases, sequences, tasks []string, bidirectional, unlink, jsonOutput bool) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("runResearchLink")
	}

	// Validate inputs
	if len(phases) == 0 && len(sequences) == 0 && len(tasks) == 0 {
		return errors.Validation("at least one of --phase, --sequence, or --task is required")
	}

	// Resolve document path
	absDocPath, err := filepath.Abs(docPath)
	if err != nil {
		return errors.Wrap(err, "resolving document path").WithField("path", docPath)
	}

	// Check document exists
	if _, err := os.Stat(absDocPath); os.IsNotExist(err) {
		return errors.NotFound("research document").WithField("path", docPath)
	}

	// Read document
	content, err := os.ReadFile(absDocPath)
	if err != nil {
		return errors.IO("reading document", err).WithField("path", absDocPath)
	}

	// Parse and update frontmatter
	newContent, linksAdded, linksRemoved, err := updateDocumentLinks(string(content), phases, sequences, tasks, unlink)
	if err != nil {
		return errors.Wrap(err, "updating links")
	}

	// Write updated document
	if err := os.WriteFile(absDocPath, []byte(newContent), 0644); err != nil {
		return errors.IO("writing document", err).WithField("path", absDocPath)
	}

	// Handle bidirectional linking (if adding, not unlinking)
	var bidirectionalCount int
	if bidirectional && !unlink {
		bidirectionalCount, err = createBidirectionalLinks(absDocPath, phases, sequences, tasks)
		if err != nil {
			// Log warning but don't fail
			fmt.Fprintln(cmd.ErrOrStderr(), ui.Warning(fmt.Sprintf("Failed to create some bidirectional links: %v", err)))
		}
	}

	if jsonOutput {
		output := struct {
			OK                bool     `json:"ok"`
			Action            string   `json:"action"`
			Document          string   `json:"document"`
			LinksAdded        int      `json:"links_added,omitempty"`
			LinksRemoved      int      `json:"links_removed,omitempty"`
			BidirectionalRefs int      `json:"bidirectional_refs_added,omitempty"`
			Phases            []string `json:"phases,omitempty"`
			Sequences         []string `json:"sequences,omitempty"`
			Tasks             []string `json:"tasks,omitempty"`
		}{
			OK:                true,
			Action:            "research_link",
			Document:          docPath,
			LinksAdded:        linksAdded,
			LinksRemoved:      linksRemoved,
			BidirectionalRefs: bidirectionalCount,
			Phases:            phases,
			Sequences:         sequences,
			Tasks:             tasks,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, ui.H1("Research Links Updated"))
	fmt.Fprintf(out, "%s %s\n", ui.Label("Document"), ui.Value(docPath))
	if unlink {
		fmt.Fprintf(out, "%s %s\n", ui.Label("Removed"), ui.Value(fmt.Sprintf("%d link(s)", linksRemoved)))
	} else {
		fmt.Fprintf(out, "%s %s\n", ui.Label("Added"), ui.Value(fmt.Sprintf("%d link(s)", linksAdded)))
		if bidirectionalCount > 0 {
			fmt.Fprintf(out, "%s %s\n", ui.Label("Bidirectional"), ui.Value(fmt.Sprintf("%d reference(s)", bidirectionalCount)))
		}
	}

	return nil
}

func updateDocumentLinks(content string, phases, sequences, tasks []string, unlink bool) (string, int, int, error) {
	// Check for frontmatter
	if !strings.HasPrefix(content, "---") {
		return "", 0, 0, errors.Validation("document has no frontmatter")
	}

	endIndex := strings.Index(content[3:], "\n---")
	if endIndex == -1 {
		return "", 0, 0, errors.Validation("invalid frontmatter format")
	}

	frontmatterStr := content[4 : 3+endIndex]
	bodyContent := content[3+endIndex+4:]

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return "", 0, 0, errors.Parse("parsing frontmatter", err)
	}

	linksAdded := 0
	linksRemoved := 0

	// Helper to update a slice field
	updateSlice := func(key string, values []string) {
		existing, _ := frontmatter[key].([]interface{})
		existingSet := make(map[string]bool)
		for _, v := range existing {
			if s, ok := v.(string); ok {
				existingSet[s] = true
			}
		}

		if unlink {
			// Remove values
			for _, v := range values {
				if existingSet[v] {
					delete(existingSet, v)
					linksRemoved++
				}
			}
		} else {
			// Add values
			for _, v := range values {
				if !existingSet[v] {
					existingSet[v] = true
					linksAdded++
				}
			}
		}

		// Convert back to slice
		var newSlice []string
		for v := range existingSet {
			newSlice = append(newSlice, v)
		}

		if len(newSlice) > 0 {
			frontmatter[key] = newSlice
		} else {
			delete(frontmatter, key)
		}
	}

	if len(phases) > 0 {
		updateSlice("linked_phases", phases)
	}
	if len(sequences) > 0 {
		updateSlice("linked_sequences", sequences)
	}
	if len(tasks) > 0 {
		updateSlice("linked_tasks", tasks)
	}

	// Serialize frontmatter
	newFrontmatter, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", 0, 0, errors.Wrap(err, "serializing frontmatter")
	}

	newContent := "---\n" + string(newFrontmatter) + "---" + bodyContent

	return newContent, linksAdded, linksRemoved, nil
}

func createBidirectionalLinks(researchDocPath string, phases, sequences, tasks []string) (int, error) {
	count := 0

	// Get research doc title for the reference
	content, err := os.ReadFile(researchDocPath)
	if err != nil {
		return 0, err
	}

	frontmatter := extractFrontmatter(string(content))
	title := filepath.Base(researchDocPath)
	if frontmatter != nil {
		if t, ok := frontmatter["title"].(string); ok && t != "" {
			title = t
		}
	}

	docType := "research"
	if frontmatter != nil {
		if t, ok := frontmatter["research_type"].(string); ok && t != "" {
			docType = t
		}
	}

	// Create reference section
	reference := fmt.Sprintf("\n## Research References\n\n- [%s](%s) - %s\n", title, researchDocPath, docType)

	// Add to phase PHASE_GOAL.md files
	for _, phase := range phases {
		goalPath := filepath.Join(phase, "PHASE_GOAL.md")
		if _, err := os.Stat(goalPath); os.IsNotExist(err) {
			continue
		}

		goalContent, err := os.ReadFile(goalPath)
		if err != nil {
			continue
		}

		// Check if reference section already exists
		if strings.Contains(string(goalContent), "## Research References") {
			// Append to existing section
			newContent := strings.Replace(string(goalContent), "## Research References\n\n", "## Research References\n\n"+fmt.Sprintf("- [%s](%s) - %s\n", title, researchDocPath, docType), 1)
			if err := os.WriteFile(goalPath, []byte(newContent), 0644); err == nil {
				count++
			}
		} else {
			// Add new section at end
			newContent := string(goalContent) + reference
			if err := os.WriteFile(goalPath, []byte(newContent), 0644); err == nil {
				count++
			}
		}
	}

	// Similar logic for sequences and tasks
	for _, seq := range sequences {
		goalPath := filepath.Join(seq, "SEQUENCE_GOAL.md")
		if _, err := os.Stat(goalPath); os.IsNotExist(err) {
			continue
		}

		goalContent, err := os.ReadFile(goalPath)
		if err != nil {
			continue
		}

		if !strings.Contains(string(goalContent), "## Research References") {
			newContent := string(goalContent) + reference
			if err := os.WriteFile(goalPath, []byte(newContent), 0644); err == nil {
				count++
			}
		}
	}

	for _, task := range tasks {
		if _, err := os.Stat(task); os.IsNotExist(err) {
			continue
		}

		taskContent, err := os.ReadFile(task)
		if err != nil {
			continue
		}

		if !strings.Contains(string(taskContent), "## Research References") {
			newContent := string(taskContent) + reference
			if err := os.WriteFile(task, []byte(newContent), 0644); err == nil {
				count++
			}
		}
	}

	return count, nil
}
