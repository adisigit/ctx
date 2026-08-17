package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(out.String()), nil
}

func CurrentBranch() (string, error) {
	return run("rev-parse", "--abbrev-ref", "HEAD")
}

func LastCommit() (string, error) {
	return run("rev-parse", "--short", "HEAD")
}

func RepoRoot() (string, error) {
	return run("rev-parse", "--show-toplevel")
}

func RepoName() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	parts := strings.Split(root, "/")
	return parts[len(parts)-1], nil
}

func RepoIdentity() (string, error) {
	url, err := run("remote", "get-url", "origin")
	if err == nil && url != "" {
		return normalizeRemote(url), nil
	}
	return RepoName()
}

func normalizeRemote(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimPrefix(url, "https://")
	url = strings.Replace(url, ":", "/", 1)
	return url
}

func ChangedFiles() ([]string, error) {
	output, err := run("diff", "--name-only", "HEAD-1", "HEAD")
	if err != nil {
		output, err = run("show", "--name-only", "--pretty=format:", "HEAD")
		if err != nil {
			return nil, err
		}
	}
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}

func IsDirty() (bool, error) {
	output, err := run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

func UncommittedFiles() ([]string, error) {
	output, err := run("status", "--porcelain")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	lines := strings.Split(output, "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		files = append(files, strings.TrimSpace(line))
	}
	return files, nil
}

func HasAnyCommit() bool {
	_, err := run("rev-parse", "HEAD")
	return err == nil
}

func CommitsSince(commitSHA string) (int, error) {
	if commitSHA == "" || commitSHA == "(no commits yet)" {
		return 0, nil
	}
	out, err := run("rev-list", commitSHA+"..HEAD", "--count")
	if err != nil {
		return 0, nil
	}
	var count int
	fmt.Sscanf(out, "%d", &count)
	return count, nil
}
