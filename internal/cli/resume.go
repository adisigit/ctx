package cli

import (
	"ctx/internal/git"
	"ctx/internal/store"
	"fmt"
	"strings"

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

	commitsSince, _ := git.CommitsSince(last.CommitSHA)

	divider := strings.Repeat("─", 32)

	fmt.Println(repo)
	fmt.Println(divider)
	fmt.Printf("Welcome back.\n\n")

	fmt.Println("Last session:")
	fmt.Printf("  %s\n\n", last.Session)

	fmt.Println("Since then:")
	fmt.Printf("  %d commit(s)\n\n", commitsSince)

	fmt.Println("Completed:")
	fmt.Printf("  %s\n\n", last.Completed)

	fmt.Println("Unfinished work:")
	fmt.Printf("  ⚠ %s\n\n", last.Remaining)

	if last.Blocker != "" && last.Blocker != "None" {
		fmt.Println("Blocker:")
		fmt.Printf("  🚫 %s\n\n", last.Blocker)
	}

	fmt.Println("Files touched:")
	if len(last.Files) == 0 {
		fmt.Println("  (none recorded)")
	} else {
		for _, f := range last.Files {
			fmt.Printf("  • %s\n", f)
		}
	}

	fmt.Printf("\nGit state: commit %s, branch %s\n", last.CommitSHA, last.Branch)
	return nil
}
