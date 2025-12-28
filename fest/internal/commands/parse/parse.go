// Package parse provides the fest parse command for structured output.
package parse

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/parser"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

var (
	format      string
	typeFilter  string
	parseAll    bool
	compact     bool
	full        bool
	inferMissing bool
	outputFile  string
)

// NewParseCommand creates the fest parse command
func NewParseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [path]",
		Short: "Parse festival documents into structured output",
		Long: `Parse festival documents into structured JSON or YAML output.

This command walks the festival hierarchy and produces structured output
suitable for external tool integration (e.g., Guild v3, visualization tools).

Examples:
  fest parse                         # Parse current festival as JSON
  fest parse --format yaml           # Output as YAML
  fest parse --type task             # Output flat list of tasks
  fest parse --type gate             # Output only gates
  fest parse --all                   # Parse all festivals
  fest parse --compact               # Summary only (no children)
  fest parse --full                  # Include document content
  fest parse -o output.json          # Write to file`,
		RunE: runParse,
	}

	cmd.Flags().StringVar(&format, "format", "json", "output format (json, yaml)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "filter by entity type (task, gate, phase, sequence)")
	cmd.Flags().BoolVar(&parseAll, "all", false, "parse all festivals in workspace")
	cmd.Flags().BoolVar(&compact, "compact", false, "compact output (summary only)")
	cmd.Flags().BoolVar(&full, "full", false, "include document content")
	cmd.Flags().BoolVar(&inferMissing, "infer", true, "infer frontmatter when missing")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "write output to file")

	return cmd
}

func runParse(cmd *cobra.Command, args []string) error {
	opts := parser.ParseOptions{
		IncludeContent: full,
		TypeFilter:     typeFilter,
		Compact:        compact,
		Format:         format,
		InferMissing:   inferMissing,
	}

	p := parser.NewParser(opts)

	var output []byte
	var err error

	if parseAll {
		output, err = parseAllFestivals(p, opts)
	} else {
		output, err = parseSingleFestival(p, args, opts)
	}

	if err != nil {
		return err
	}

	// Output to file or stdout
	if outputFile != "" {
		return os.WriteFile(outputFile, output, 0644)
	}
	fmt.Println(string(output))
	return nil
}

func parseSingleFestival(p *parser.Parser, args []string, opts parser.ParseOptions) ([]byte, error) {
	var festivalPath string

	if len(args) > 0 {
		festivalPath = args[0]
	} else {
		// Try to find current festival
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		festivalPath, err = tpl.FindFestivalRoot(cwd)
		if err != nil {
			return nil, errors.NotFound("festival")
		}
	}

	festival, err := p.ParseFestival(festivalPath)
	if err != nil {
		return nil, err
	}

	// Apply type filter for flat output
	if opts.TypeFilter != "" {
		entities := parser.FlattenByType(festival, opts.TypeFilter)
		result := &parser.FlattenedResult{
			Query:    fmt.Sprintf("type:%s", opts.TypeFilter),
			Count:    len(entities),
			Entities: entities,
		}
		return parser.Format(result, opts.Format, opts.Compact)
	}

	// Compact mode: exclude children
	if opts.Compact {
		summary := struct {
			Type    string                  `json:"type" yaml:"type"`
			ID      string                  `json:"id" yaml:"id"`
			Name    string                  `json:"name" yaml:"name"`
			Status  string                  `json:"status" yaml:"status"`
			Path    string                  `json:"path" yaml:"path"`
			Summary *parser.FestivalSummary `json:"summary" yaml:"summary"`
		}{
			Type:    festival.Type,
			ID:      festival.ID,
			Name:    festival.Name,
			Status:  festival.Status,
			Path:    festival.Path,
			Summary: festival.Summary,
		}
		return parser.Format(summary, opts.Format, opts.Compact)
	}

	return parser.Format(festival, opts.Format, false)
}

func parseAllFestivals(p *parser.Parser, opts parser.ParseOptions) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return nil, errors.NotFound("festivals/ directory")
	}

	result := &parser.AllFestivalsResult{
		Summary: &parser.WorkspaceSummary{},
	}

	// Parse each status directory
	statusDirs := []string{"planned", "active", "completed", "dungeon"}
	for _, status := range statusDirs {
		statusPath := filepath.Join(festivalsRoot, status)
		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if entry.Name()[0] == '.' {
				continue
			}

			festival, err := p.ParseFestival(filepath.Join(statusPath, entry.Name()))
			if err != nil {
				continue
			}

			// Set status from directory location
			festival.Status = status

			result.Festivals = append(result.Festivals, *festival)
			result.Summary.FestivalCount++
			result.Summary.PhaseCount += festival.Summary.PhaseCount
			result.Summary.TaskCount += festival.Summary.TaskCount
		}
	}

	return parser.Format(result, opts.Format, opts.Compact)
}
