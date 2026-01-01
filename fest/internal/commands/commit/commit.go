// Package commit provides the fest commit command for git integration.
package commit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

var (
	message string
	taskRef string
	noTag   bool
	jsonOut bool
)

// NewCommitCommand creates the fest commit command
func NewCommitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Create git commit with task reference",
		Long: `Create a git commit with the current task ID embedded in the message.

The fest commit command wraps git commit and automatically prepends the
current task's fest_ref ID to the commit message.

Examples:
  fest commit -m "Implement feature"
  # → git commit -m "[FEST-a3b2c1] Implement feature"

  fest commit --task FEST-b4c5d6 -m "Related work"
  # → git commit -m "[FEST-b4c5d6] Related work"

  fest commit --no-tag -m "No reference"
  # → git commit -m "No reference"`,
		RunE: runCommit,
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "commit message")
	cmd.Flags().StringVar(&taskRef, "task", "", "task reference ID to use (e.g., FEST-a3b2c1)")
	cmd.Flags().BoolVar(&noTag, "no-tag", false, "don't prepend task reference")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output result as JSON")

	cmd.MarkFlagRequired("message")

	return cmd
}

// CommitResult represents the result of a commit operation
type CommitResult struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash,omitempty"`
	Message string `json:"message"`
	TaskRef string `json:"task_ref,omitempty"`
	Error   string `json:"error,omitempty"`
}

func runCommit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	result := &CommitResult{}

	// Check if we're in a git repository
	if !isGitRepo() {
		result.Success = false
		result.Error = "not in a git repository"
		return outputResult(result)
	}

	// Get task reference
	var ref string
	if !noTag {
		if taskRef != "" {
			// Validate provided ref
			if !id.Validate(taskRef) {
				result.Success = false
				result.Error = fmt.Sprintf("invalid task reference format: %s (expected FEST-xxxxxx)", taskRef)
				return outputResult(result)
			}
			ref = taskRef
		} else {
			// Try to detect from current location
			var err error
			ref, err = detectCurrentTaskRef(ctx)
			if err != nil {
				// Not an error - just no ref available
				ref = ""
			}
		}
	}

	// Build commit message
	commitMessage := message
	if ref != "" {
		commitMessage = fmt.Sprintf("[%s] %s", ref, message)
	}

	// Execute git commit
	hash, err := executeGitCommit(commitMessage)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return outputResult(result)
	}

	result.Success = true
	result.Hash = hash
	result.Message = commitMessage
	result.TaskRef = ref

	return outputResult(result)
}

func outputResult(result *CommitResult) error {
	if jsonOut {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		if result.Success {
			fmt.Printf("Committed: %s\n", result.Hash)
			fmt.Printf("Message: %s\n", result.Message)
			if result.TaskRef != "" {
				fmt.Printf("Task: %s\n", result.TaskRef)
			}
		} else {
			return errors.New(result.Error)
		}
	}
	return nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func executeGitCommit(message string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "git commit failed")
	}

	// Get the commit hash
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	var out bytes.Buffer
	hashCmd.Stdout = &out
	if err := hashCmd.Run(); err != nil {
		return "", errors.Wrap(err, "failed to get commit hash")
	}

	return strings.TrimSpace(out.String()), nil
}

func detectCurrentTaskRef(ctx context.Context) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Try to find fest.yaml to locate festival root
	festivalPath, err := tpl.FindFestivalRoot(cwd)
	if err != nil {
		return "", errors.NotFound("festival")
	}

	// Check if we're in a task directory
	relPath, err := filepath.Rel(festivalPath, cwd)
	if err != nil {
		return "", err
	}

	// Walk up looking for a task file
	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) >= 3 {
		// We might be in sequence/task level
		seqDir := filepath.Join(festivalPath, parts[0], parts[1])
		entries, err := os.ReadDir(seqDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				if strings.HasPrefix(strings.ToUpper(entry.Name()), "SEQUENCE_") {
					continue
				}

				// Try to parse frontmatter from this task
				taskPath := filepath.Join(seqDir, entry.Name())
				content, err := os.ReadFile(taskPath)
				if err != nil {
					continue
				}

				fm, _, err := frontmatter.Parse(content)
				if err != nil || fm == nil {
					continue
				}

				if fm.Ref != "" {
					return fm.Ref, nil
				}
			}
		}
	}

	return "", errors.NotFound("task reference")
}
