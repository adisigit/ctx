package cli

import (
	"ctx/internal/git"
	"ctx/internal/store"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagLogLimit  int
	flagLogBranch string
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show context history for this repo",
	RunE:  runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().IntVar(&flagLogLimit, "limit", 10, "max number of entries to show")
	logCmd.Flags().StringVar(&flagLogBranch, "branch", "", "filter by branch name")
}

func runLog(cmd *cobra.Command, args []string) error {
	repo, err := git.RepoIdentity()
	if err != nil {
		return fmt.Errorf("not a git repo or git not detected: %w", err)
	}
	s, err := store.NewSQLiteStore()
	if err != nil {
		return err
	}
	entries, err := s.AllByRepo(repo)
	if err != nil {
		return err
	}
	if flagLogBranch != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Branch == flagLogBranch {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if len(entries) == 0 {
		fmt.Println("No context history found")
		return nil
	}
	if len(entries) > flagLogLimit {
		entries = entries[:flagLogLimit]
	}
	for i, e := range entries {
		fmt.Printf("[%s] %s (%s @ %s)\n",
			e.CreatedAt.Format("02 Jan 15:04"), e.Session, e.Branch, e.CommitSHA)
		fmt.Printf("  completed: %s\n", e.Completed)
		fmt.Printf("  remaining: %s\n", e.Remaining)
		if e.Blocker != "" && e.Blocker != "None" {
			fmt.Printf("  blocker:   %s\n", e.Blocker)
		}
		if i != len(entries)-1 {
			fmt.Println()
		}
	}
	return nil
}
