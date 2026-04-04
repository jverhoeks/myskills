package github

import (
	"fmt"
	"os"
	"os/exec"
)

// HasGH returns true if the gh CLI is available.
func HasGH() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// CreateBranch creates and checks out a new branch in the given repo directory.
func CreateBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = repoDir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating branch %s: %w", branch, err)
	}
	return nil
}

// CommitAll stages and commits all changes in the given repo directory.
func CommitAll(repoDir, message string) error {
	add := exec.Command("git", "add", ".")
	add.Dir = repoDir
	if err := add.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	commit := exec.Command("git", "commit", "-m", message)
	commit.Dir = repoDir
	commit.Stderr = os.Stderr
	if err := commit.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// PushBranch pushes the current branch to origin.
func PushBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

// OpenPRWithGH opens a PR using the gh CLI.
func OpenPRWithGH(repoDir, title, body string) (string, error) {
	args := []string{"pr", "create", "--title", title, "--body", body}
	cmd := exec.Command("gh", args...)
	cmd.Dir = repoDir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w", err)
	}
	return string(out), nil
}

// Submit handles the full submit workflow: branch, commit, push, PR.
func Submit(repoDir, name, ghMethod, token string) error {
	branch := "skill/" + name

	if err := CreateBranch(repoDir, branch); err != nil {
		return err
	}

	if err := CommitAll(repoDir, fmt.Sprintf("feat: add skill %s", name)); err != nil {
		return err
	}

	if err := PushBranch(repoDir, branch); err != nil {
		return err
	}

	title := fmt.Sprintf("Add skill: %s", name)
	body := fmt.Sprintf("Adds the `%s` skill to the org repository.", name)

	if ghMethod == "gh" && HasGH() {
		url, err := OpenPRWithGH(repoDir, title, body)
		if err != nil {
			return err
		}
		fmt.Printf("PR created: %s", url)
		return nil
	}

	if token != "" {
		fmt.Fprintf(os.Stderr, "Token-based PR creation is not yet implemented.\n")
		fmt.Fprintf(os.Stderr, "Branch %s has been pushed. Please create the PR manually.\n", branch)
		return nil
	}

	fmt.Printf("Branch %s pushed. Please create the PR manually on GitHub.\n", branch)
	return nil
}
