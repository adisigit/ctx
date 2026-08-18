package cli

import (
	"ctx/internal/git"
	"ctx/internal/store"
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current git state and last recorded context",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	repo, err := git.RepoIdentity()
	if err != nil {
		return fmt.Errorf("not a git repo or git not detected: %w", err)
	}
	branch, err := git.CurrentBranch()
	if err != nil {
		return err
	}
	var commit string
	if git.HasAnyCommit() {
		commit, err = git.LastCommit()
		if err != nil {
			return err
		}
	} else {
		commit = "no commits yet"
	}
	dirty, err := git.IsDirty()
	if err != nil {
		return err
	}
	var changedCount int
	if dirty {
		file, err := git.UncommittedFiles()
		if err != nil {
			return err
		}
		changedCount = len(file)
	}
	fmt.Printf("Repo:   %s\n", repo)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Printf("Commit: %s\n", commit)
	if dirty {
		fmt.Printf("Status: %d uncommitted file(s)\n", changedCount)
	} else {
		fmt.Println("Status: clean")
	}
	s, err := store.NewSQLiteStore()
	if err != nil {
		return err
	}
	last, err := s.LastByRepo(repo)
	if err != nil {
		return err
	}
	fmt.Println()
	if last == nil {
		fmt.Println("Last context: none recorded yet")
	} else {
		fmt.Printf("Last context: %s (\"%s\")\n", last.CreatedAt.Format("02 Jan 15:04"), last.Session)
		if last.Blocker != "" && last.Blocker != "None" {
			fmt.Printf("⚠ Open blocker: %s\n", last.Blocker)
		}
	}

	return nil
}
