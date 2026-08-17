package cli

import (
	"bufio"
	"ctx/internal/git"
	"ctx/internal/model"
	"ctx/internal/store"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	flagSession   string
	flagCompleted string
	flagRemaining string
	flagBlocker   string
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "write context work session",
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
	doneCmd.Flags().StringVar(&flagSession, "session", "", "session name, e.g. \"Webhook retry\"")
	doneCmd.Flags().StringVar(&flagCompleted, "completed", "", "what got finished")
	doneCmd.Flags().StringVar(&flagRemaining, "remaining", "", "what's left")
	doneCmd.Flags().StringVar(&flagBlocker, "blocker", "", "current blocker, if any")
}

func runDone(cmd *cobra.Command, args []string) error {
	if flagSession == "" && flagCompleted == "" && flagRemaining == "" && flagBlocker == "" {
		if err := promptDoneFields(); err != nil {
			return err
		}
	}
	branch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	repo, err := git.RepoIdentity()
	if err != nil {
		return err
	}
	var commit string
	var files []string
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
	if dirty {
		files, err = git.UncommittedFiles()
		if err != nil {
			return err
		}
	} else if git.HasAnyCommit() {
		files, err = git.ChangedFiles()
		if err != nil {
			return err
		}
	}
	s, err := store.NewSQLiteStore()
	if err != nil {
		return err
	}
	entery := model.ContextEntry{
		ID:        uuid.NewString(),
		Repo:      repo,
		Branch:    branch,
		CommitSHA: commit,
		Files:     files,
		Session:   flagSession,
		Completed: flagCompleted,
		Remaining: flagRemaining,
		Blocker:   flagBlocker,
		Type:      model.Progress,
		CreatedAt: time.Now(),
	}
	if err := s.Save(entery); err != nil {
		return err
	}
	fmt.Println("done")
	return nil
}

func promptDoneFields() error {
	reader := bufio.NewReader(os.Stdin)

	flagSession = ask(reader, "Session name")
	flagCompleted = ask(reader, "What did you complete")
	flagRemaining = ask(reader, "What's remaining")
	flagBlocker = ask(reader, "Any blocker (leave empty for none)")

	if flagBlocker == "" {
		flagBlocker = "None"
	}
	return nil
}

func ask(reader *bufio.Reader, question string) string {
	fmt.Printf("%s: ", question)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
