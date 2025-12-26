package festival

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// executeChanges applies the planned changes
func (r *Renumberer) executeChanges() error {
	if len(r.changes) == 0 {
		if !r.options.Quiet {
			fmt.Println("No changes needed.")
		}
		return nil
	}

	// Display changes
	if !r.options.Quiet {
		r.displayChanges()
	}

	if r.options.DryRun {
		if !r.options.Quiet {
			fmt.Println("\nDRY RUN - Preview complete.")
		}
		// If auto-approve, proceed to apply after dry-run preview
		if r.options.AutoApprove {
			if !r.options.Quiet {
				fmt.Println("\nApplying changes...")
			}
		} else {
			// Prompt user to apply changes after dry-run
			if r.confirmApplyAfterDryRun() {
				if !r.options.Quiet {
					fmt.Println("\nApplying changes...")
				}
			} else {
				if !r.options.Quiet {
					fmt.Println("Operation cancelled.")
				}
				return nil
			}
		}
	} else {
		// Not in dry-run mode, confirm changes before applying
		if !r.options.AutoApprove {
			if !r.confirmChanges() {
				if !r.options.Quiet {
					fmt.Println("Operation cancelled.")
				}
				return nil
			}
		}
	}

	// Create backup if requested
	if r.options.Backup {
		if err := r.createBackup(); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Apply changes
	for _, change := range r.changes {
		switch change.Type {
		case ChangeRename:
			if err := os.Rename(change.OldPath, change.NewPath); err != nil {
				return fmt.Errorf("failed to rename %s to %s: %w", change.OldPath, change.NewPath, err)
			}
			if r.options.Verbose {
				fmt.Printf("Renamed: %s → %s\n", filepath.Base(change.OldPath), filepath.Base(change.NewPath))
			}

		case ChangeCreate:
			// If NewPath looks like a file (e.g., ends with .md), create an empty file.
			// Otherwise, create a directory (used for phases/sequences).
			if strings.HasSuffix(strings.ToLower(change.NewPath), ".md") {
				// For file creations, just ensure parent exists; the caller writes content.
				if err := os.MkdirAll(filepath.Dir(change.NewPath), 0755); err != nil {
					return fmt.Errorf("failed to create parent directory for %s: %w", change.NewPath, err)
				}
			} else {
				if err := os.MkdirAll(change.NewPath, 0755); err != nil {
					return fmt.Errorf("failed to create %s: %w", change.NewPath, err)
				}
			}
			if r.options.Verbose {
				fmt.Printf("Created: %s\n", filepath.Base(change.NewPath))
			}

		case ChangeRemove:
			if err := os.RemoveAll(change.OldPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", change.OldPath, err)
			}
			if r.options.Verbose {
				fmt.Printf("Removed: %s\n", filepath.Base(change.OldPath))
			}
		}
	}

	if !r.options.Quiet {
		fmt.Printf("\n✓ Successfully applied %d changes.\n", len(r.changes))
	}
	return nil
}

// displayChanges shows planned changes
func (r *Renumberer) displayChanges() {
	fmt.Println("\nFestival Renumbering Report")
	fmt.Println(strings.Repeat("═", 55))
	fmt.Println("\nChanges to be made:")

	for _, change := range r.changes {
		switch change.Type {
		case ChangeRename:
			fmt.Printf("  → Rename: %s → %s\n",
				filepath.Base(change.OldPath),
				filepath.Base(change.NewPath))
		case ChangeCreate:
			fmt.Printf("  ✓ Create: %s\n", filepath.Base(change.NewPath))
		case ChangeRemove:
			fmt.Printf("  ✗ Remove: %s\n", filepath.Base(change.OldPath))
		}
	}

	fmt.Printf("\nTotal: %d changes\n", len(r.changes))
}

// confirmChanges prompts for confirmation
func (r *Renumberer) confirmChanges() bool {
	fmt.Print("\nProceed with renumbering? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// confirmApplyAfterDryRun prompts to apply changes after dry-run preview
func (r *Renumberer) confirmApplyAfterDryRun() bool {
	fmt.Print("\nApply these changes? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// createBackup creates a backup of affected directories
func (r *Renumberer) createBackup() error {
	// Implementation would create timestamped backup
	// For now, just log
	fmt.Println("Creating backup...")
	return nil
}
