package commit

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

var (
	queryTask     string
	querySequence string
	queryPhase    string
	queryLimit    int
	queryJson     bool
)

var taskRefPattern = regexp.MustCompile(`\[FEST-[a-z0-9]{6}\]`)

// NewCommitsCommand creates the fest commits query command
func NewCommitsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "Query commits by festival element",
		Long: `Query git commits that reference festival elements.

Search commits by task ID, sequence, or phase. Uses git log with grep
to find commits that contain festival references in their messages.

Examples:
  fest commits                           # All commits for current festival
  fest commits --task FEST-a3b2c1        # Commits for specific task
  fest commits --sequence 01_foundation  # Commits for sequence
  fest commits --phase 002_IMPLEMENT     # Commits for phase
  fest commits --limit 20                # Limit results`,
		RunE: runCommits,
	}

	cmd.Flags().StringVar(&queryTask, "task", "", "query commits for specific task (e.g., FEST-a3b2c1)")
	cmd.Flags().StringVar(&querySequence, "sequence", "", "query commits for sequence")
	cmd.Flags().StringVar(&queryPhase, "phase", "", "query commits for phase")
	cmd.Flags().IntVar(&queryLimit, "limit", 50, "maximum number of commits to return")
	cmd.Flags().BoolVar(&queryJson, "json", false, "output as JSON")

	return cmd
}

// Commit represents a git commit
type Commit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Author    string    `json:"author"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	TaskRefs  []string  `json:"task_refs"`
}

// CommitsResult represents the query result
type CommitsResult struct {
	Query   string   `json:"query"`
	Count   int      `json:"count"`
	Commits []Commit `json:"commits"`
}

func runCommits(cmd *cobra.Command, args []string) error {
	// Build grep pattern
	var grepPattern string
	var queryDesc string

	switch {
	case queryTask != "":
		if !id.Validate(queryTask) {
			return errors.Validation("invalid task reference format").
				WithField("task", queryTask)
		}
		grepPattern = fmt.Sprintf("\\[%s\\]", queryTask)
		queryDesc = fmt.Sprintf("task:%s", queryTask)
	case querySequence != "":
		// Match any task in this sequence
		grepPattern = fmt.Sprintf("\\[FEST-[a-z0-9]{6}\\].*%s", querySequence)
		queryDesc = fmt.Sprintf("sequence:%s", querySequence)
	case queryPhase != "":
		// Match any task in this phase
		grepPattern = fmt.Sprintf("\\[FEST-[a-z0-9]{6}\\].*%s", queryPhase)
		queryDesc = fmt.Sprintf("phase:%s", queryPhase)
	default:
		// All festival commits
		grepPattern = "\\[FEST-[a-z0-9]{6}\\]"
		queryDesc = "all"
	}

	commits, err := queryCommits(grepPattern, queryLimit)
	if err != nil {
		return err
	}

	result := &CommitsResult{
		Query:   queryDesc,
		Count:   len(commits),
		Commits: commits,
	}

	if queryJson {
		if err := shared.EncodeJSON(cmd.OutOrStdout(), result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		printCommits(result)
	}

	return nil
}

func queryCommits(grepPattern string, limit int) ([]Commit, error) {
	// Use git log with grep
	args := []string{
		"log",
		"--grep=" + grepPattern,
		"-E",
		fmt.Sprintf("-n%d", limit),
		"--format=%H|%h|%an|%aI|%s",
	}

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if just no matches
		if strings.Contains(stderr.String(), "fatal") {
			return nil, errors.Wrap(err, "git log failed").
				WithField("stderr", stderr.String())
		}
		// No matches is fine
		return []Commit{}, nil
	}

	return parseGitLog(out.String())
}

func parseGitLog(output string) ([]Commit, error) {
	var commits []Commit

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		date, _ := time.Parse(time.RFC3339, parts[3])

		commit := Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      date,
			Message:   parts[4],
			TaskRefs:  id.ExtractFromMessage(parts[4]),
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func printCommits(result *CommitsResult) {
	if result.Count == 0 {
		fmt.Printf("%s %s\n", ui.Warning("No commits found"), ui.Dim(fmt.Sprintf("query: %s", result.Query)))
		return
	}

	fmt.Println(ui.H1("Commits"))
	fmt.Printf("%s %s\n", ui.Label("Query"), ui.Value(result.Query))
	fmt.Printf("%s %s\n", ui.Label("Count"), ui.Value(fmt.Sprintf("%d", result.Count)))

	for _, commit := range result.Commits {
		message := highlightTaskRefs(commit.Message)

		fmt.Printf("\n%s %s\n", ui.Value(commit.ShortHash), message)
		fmt.Printf("  %s %s\n", ui.Label("Author"), ui.Value(commit.Author))
		fmt.Printf("  %s %s\n", ui.Label("Date"), ui.Dim(commit.Date.Format("2006-01-02 15:04")))
		if len(commit.TaskRefs) > 0 {
			fmt.Printf("  %s %s\n", ui.Label("Tasks"), ui.Value(strings.Join(commit.TaskRefs, ", "), ui.TaskColor))
		}
	}
}

func highlightTaskRefs(message string) string {
	if message == "" {
		return message
	}

	return taskRefPattern.ReplaceAllStringFunc(message, func(match string) string {
		return ui.Value(match, ui.TaskColor)
	})
}
