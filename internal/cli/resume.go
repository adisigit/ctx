package cli

import (
	"ctx/internal/git"
	"ctx/internal/store"
	"fmt"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Show last context from this repo",
	RunE:  runResume,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func runResume(smd *cobra.Command, args []string) error {
	repo, err := git.RepoIdentity()
	if err != nil {
		return fmt.Errorf("not git repo or git repo not detected: %w", err)
	}
	s, err := store.NewSQLiteStore()
	if err != nil {
		return err
	}
	last, err := s.LastByRepo(repo)
	if err != nil {
		return err
	}
	if last == nil {
		fmt.Printf("no context found for repo %s", repo)
		return nil
	}
	fmt.Println("Session:")
	fmt.Printf("  %s\n", last.Session)
	fmt.Println("Git state:")
	fmt.Printf("  commit = %s\n", last.CommitSHA)
	fmt.Println("Completed:")
	fmt.Printf("  %s\n", last.Completed)
	fmt.Println("Remaining:")
	fmt.Printf("  %s\n", last.Remaining)
	fmt.Println("Blocker:")
	fmt.Printf("  %s\n", last.Blocker)
	return nil
}
