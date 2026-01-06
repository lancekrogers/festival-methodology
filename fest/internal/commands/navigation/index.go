package navigation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/index"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewIndexCommand creates the index command group
func NewIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage festival indices",
		Long: `Generate and validate festival indices for Guild integration.

The index file (.festival/index.json) provides a machine-readable representation
of the festival structure, including phases, sequences, and tasks.

For workspace-wide indexing (Guild v3), use the 'tree' subcommand.`,
	}

	cmd.AddCommand(newIndexWriteCommand())
	cmd.AddCommand(newIndexValidateCommand())
	cmd.AddCommand(newIndexShowCommand())
	cmd.AddCommand(newIndexTreeCommand())
	cmd.AddCommand(newIndexDiffCommand())

	return cmd
}

func newIndexWriteCommand() *cobra.Command {
	var outputPath string
	var prettyPrint bool

	cmd := &cobra.Command{
		Use:   "write [festival-path]",
		Short: "Generate festival index",
		Long: `Generate a festival index from the filesystem structure.

The index is written to .festival/index.json within the festival directory.
Use --output to write to a different location.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			writer := index.NewIndexWriter(festivalRoot)
			idx, err := writer.Generate()
			if err != nil {
				return errors.Wrap(err, "generating index")
			}

			// Determine output path
			if outputPath == "" {
				outputPath = filepath.Join(festivalRoot, ".festival", index.IndexFileName)
			}

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return errors.IO("creating output directory", err).WithField("path", filepath.Dir(outputPath))
			}

			// Save the index
			if err := idx.Save(outputPath); err != nil {
				return errors.IO("saving index", err).WithField("path", outputPath)
			}

			summary := idx.Summary()
			fmt.Println(ui.H1("Festival Index"))
			fmt.Printf("%s %s\n", ui.Label("Path"), ui.Dim(outputPath))
			fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(idx.FestivalID, ui.FestivalColor))
			fmt.Printf("%s %s\n", ui.Label("Phases"), ui.Value(fmt.Sprintf("%d", summary.PhaseCount)))
			fmt.Printf("%s %s\n", ui.Label("Sequences"), ui.Value(fmt.Sprintf("%d", summary.SequenceCount)))
			fmt.Printf("%s %s\n", ui.Label("Tasks"), ui.Value(fmt.Sprintf("%d", summary.TaskCount)))
			if summary.ManagedCount > 0 {
				fmt.Printf("%s %s\n", ui.Label("Managed gates"), ui.Value(fmt.Sprintf("%d", summary.ManagedCount), ui.GateColor))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path (default: .festival/index.json)")
	cmd.Flags().BoolVar(&prettyPrint, "pretty", true, "Pretty print JSON output")

	return cmd
}

func newIndexValidateCommand() *cobra.Command {
	var indexPath string

	cmd := &cobra.Command{
		Use:   "validate [festival-path]",
		Short: "Validate festival index against filesystem",
		Long: `Validate that the festival index matches the actual filesystem structure.

Reports:
- Entries in index that don't exist on disk (missing)
- Files on disk that aren't in the index (extra)
- Missing goal files (warnings)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			// Determine index path
			if indexPath == "" {
				indexPath = filepath.Join(festivalRoot, ".festival", index.IndexFileName)
			}

			// Check if index exists
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				return errors.NotFound("index file").WithField("path", indexPath).WithField("hint", "run 'fest index write' to generate it")
			}

			result, err := index.ValidateFromFile(festivalRoot, indexPath)
			if err != nil {
				return errors.Wrap(err, "validating index")
			}

			// Print results
			fmt.Println(ui.H1("Index Validation"))
			if result.Valid && len(result.Warnings) == 0 && len(result.ExtraInFS) == 0 {
				fmt.Println(ui.Success("✓ Index is valid and synchronized with filesystem."))
				return nil
			}

			if len(result.Errors) > 0 {
				fmt.Println(ui.H2("Errors"))
				for _, e := range result.Errors {
					fmt.Printf("%s %s\n", ui.Error(strings.ToUpper(e.Type)), ui.Dim(e.Path))
					fmt.Printf("  %s\n", e.Message)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Println()
				fmt.Println(ui.H2("Warnings"))
				for _, w := range result.Warnings {
					fmt.Printf("%s %s\n", ui.Warning("WARN"), ui.Dim(w))
				}
			}

			if len(result.ExtraInFS) > 0 {
				fmt.Println()
				fmt.Println(ui.H2("Files Not In Index"))
				for _, f := range result.ExtraInFS {
					fmt.Printf("%s %s\n", ui.Warning("EXTRA"), ui.Dim(f))
				}
			}

			if !result.Valid {
				return errors.Validation("index validation failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&indexPath, "index", "i", "", "Path to index file (default: .festival/index.json)")

	return cmd
}

func newIndexShowCommand() *cobra.Command {
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "show [festival-path]",
		Short: "Show festival index contents",
		Long:  `Display the contents of the festival index file.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			indexPath := filepath.Join(festivalRoot, ".festival", index.IndexFileName)

			// Check if index exists
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				return errors.NotFound("index file").WithField("path", indexPath).WithField("hint", "run 'fest index write' to generate it")
			}

			idx, err := index.LoadIndex(indexPath)
			if err != nil {
				return errors.IO("loading index", err).WithField("path", indexPath)
			}

			if showJSON {
				data, err := json.MarshalIndent(idx, "", "  ")
				if err != nil {
					return errors.Wrap(err, "formatting index as JSON")
				}
				fmt.Println(string(data))
				return nil
			}

			// Human-readable output
			fmt.Println(ui.H1("Festival Index"))
			fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(idx.FestivalID, ui.FestivalColor))
			fmt.Printf("%s %s\n", ui.Label("Spec Version"), ui.Value(fmt.Sprintf("%d", idx.FestSpec)))
			fmt.Printf("%s %s\n", ui.Label("Generated"), ui.Dim(idx.GeneratedAt.Format("2006-01-02 15:04:05")))
			fmt.Println(ui.Dim(strings.Repeat("─", 60)))

			for _, phase := range idx.Phases {
				fmt.Println()
				fmt.Println(ui.H2(fmt.Sprintf("Phase %s", phase.PhaseID)))
				if phase.GoalFile != "" {
					fmt.Printf("%s %s\n", ui.Label("Goal"), ui.Dim(phase.GoalFile))
				}

				for _, seq := range phase.Sequences {
					fmt.Printf("%s %s\n", ui.Label("Sequence"), ui.Value(seq.SequenceID, ui.SequenceColor))
					if seq.GoalFile != "" {
						fmt.Printf("  %s %s\n", ui.Label("Goal"), ui.Dim(seq.GoalFile))
					}

					for _, task := range seq.Tasks {
						if task.Managed {
							fmt.Printf("  %s %s %s\n",
								ui.Dim("•"),
								ui.Value(task.TaskID, ui.TaskColor),
								ui.Dim(fmt.Sprintf("(gate: %s)", task.GateID)))
							continue
						}
						fmt.Printf("  %s %s\n", ui.Dim("•"), ui.Value(task.TaskID, ui.TaskColor))
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

func newIndexTreeCommand() *cobra.Command {
	var outputPath string
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Generate workspace-wide tree index",
		Long: `Generate a tree index of all festivals in the workspace.

The tree index groups festivals by status (planned, active, completed, dungeon)
and provides a complete hierarchical view for Guild v3 integration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return errors.IO("getting working directory", err)
			}

			workspaceRoot, err := tpl.FindFestivalsRoot(cwd)
			if err != nil {
				return errors.NotFound("festivals/ directory")
			}

			syncer := index.NewTreeSyncer(workspaceRoot)
			tree, err := syncer.Sync()
			if err != nil {
				return errors.Wrap(err, "syncing tree index")
			}

			// Save if output path specified
			if outputPath != "" {
				if err := tree.Save(outputPath); err != nil {
					return errors.IO("saving tree index", err).WithField("path", outputPath)
				}
				fmt.Printf("Tree index written to: %s\n", outputPath)
			}

			if showJSON {
				data, err := json.MarshalIndent(tree, "", "  ")
				if err != nil {
					return errors.Wrap(err, "formatting tree as JSON")
				}
				fmt.Println(string(data))
				return nil
			}

			// Human-readable output
			fmt.Println(ui.H1("Workspace Index"))
			fmt.Printf("%s %s\n", ui.Label("Workspace"), ui.Dim(tree.Workspace.Path))
			fmt.Printf("%s %s\n", ui.Label("Festivals"), ui.Value(fmt.Sprintf("%d", tree.Workspace.FestivalCount)))
			fmt.Printf("%s %s\n", ui.Label("Tasks"), ui.Value(fmt.Sprintf("%d/%d completed", tree.Workspace.CompletedTasks, tree.Workspace.TotalTasks)))
			fmt.Println(ui.Dim(strings.Repeat("─", 60)))

			printFestivalGroup("Planned", tree.Festivals.Planned)
			printFestivalGroup("Active", tree.Festivals.Active)
			printFestivalGroup("Completed", tree.Festivals.Completed)
			printFestivalGroup("Dungeon", tree.Festivals.Dungeon)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Save tree index to file")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

func printFestivalGroup(name string, festivals []index.FestivalNode) {
	if len(festivals) == 0 {
		return
	}
	fmt.Println()
	fmt.Println(ui.H2(name))
	for _, f := range festivals {
		progress := int(f.Progress * 100)
		fmt.Printf("%s %s %s\n",
			ui.Value(f.Name, ui.FestivalColor),
			ui.Dim(fmt.Sprintf("(%d phases, %d/%d tasks)", f.PhaseCount, f.CompletedTasks, f.TaskCount)),
			ui.Value(fmt.Sprintf("%d%%", progress)))
	}
}

func newIndexDiffCommand() *cobra.Command {
	var oldPath string
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare tree indexes to detect changes",
		Long: `Compare two tree indexes to detect changes between them.

This is useful for tracking progress over time or detecting changes
since the last sync.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldPath == "" {
				return errors.Validation("--old flag is required")
			}

			// Load old index
			oldTree, err := index.LoadTreeIndex(oldPath)
			if err != nil {
				return errors.IO("loading old tree index", err).WithField("path", oldPath)
			}

			// Generate current index
			cwd, err := os.Getwd()
			if err != nil {
				return errors.IO("getting working directory", err)
			}

			workspaceRoot, err := tpl.FindFestivalsRoot(cwd)
			if err != nil {
				return errors.NotFound("festivals/ directory")
			}

			syncer := index.NewTreeSyncer(workspaceRoot)
			newTree, err := syncer.Sync()
			if err != nil {
				return errors.Wrap(err, "syncing current tree index")
			}

			// Compute diff
			diff := index.ComputeDiff(oldTree, newTree)

			if showJSON {
				data, err := json.MarshalIndent(diff, "", "  ")
				if err != nil {
					return errors.Wrap(err, "formatting diff as JSON")
				}
				fmt.Println(string(data))
				return nil
			}

			// Human-readable output
			if !diff.HasChanges() {
				fmt.Println(ui.H1("Index Diff"))
				fmt.Println(ui.Success("✓ No changes detected."))
				return nil
			}

			fmt.Println(ui.H1("Index Diff"))
			fmt.Printf("%s %s\n", ui.Label("Since"), ui.Dim(oldTree.IndexedAt.Format("2006-01-02 15:04:05")))

			s := diff.Summary
			if s.FestivalsAdded > 0 || s.FestivalsRemoved > 0 || s.FestivalsMoved > 0 {
				fmt.Printf("%s %s\n", ui.Label("Festivals"),
					ui.Value(fmt.Sprintf("+%d -%d ~%d moved", s.FestivalsAdded, s.FestivalsRemoved, s.FestivalsMoved)))
			}
			if s.PhasesAdded > 0 || s.PhasesRemoved > 0 {
				fmt.Printf("%s %s\n", ui.Label("Phases"),
					ui.Value(fmt.Sprintf("+%d -%d", s.PhasesAdded, s.PhasesRemoved)))
			}
			if s.SequencesAdded > 0 || s.SequencesRemoved > 0 {
				fmt.Printf("%s %s\n", ui.Label("Sequences"),
					ui.Value(fmt.Sprintf("+%d -%d", s.SequencesAdded, s.SequencesRemoved)))
			}
			if s.TasksAdded > 0 || s.TasksRemoved > 0 || s.TasksCompleted > 0 {
				fmt.Printf("%s %s\n", ui.Label("Tasks"),
					ui.Value(fmt.Sprintf("+%d -%d ✓%d completed", s.TasksAdded, s.TasksRemoved, s.TasksCompleted)))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&oldPath, "old", "", "Path to old tree index file (required)")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

// resolveFestivalRoot determines the festival root directory
func resolveFestivalRoot(args []string) (string, error) {
	if len(args) > 0 {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return "", errors.Wrap(err, "resolving path").WithField("path", args[0])
		}
		return absPath, nil
	}

	// Try to find festival root from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting working directory", err)
	}

	// Look for festivals/ parent or a festival indicator
	root, err := tpl.FindFestivalsRoot(cwd)
	if err == nil {
		// If we're in a festivals/ tree, try to find the specific festival
		rel, _ := filepath.Rel(root, cwd)
		if rel != "." && rel != "" {
			// We might be inside a festival - find its root
			parts := filepath.SplitList(rel)
			if len(parts) > 0 {
				return filepath.Join(root, parts[0]), nil
			}
		}
	}

	// Default to current directory
	return cwd, nil
}
