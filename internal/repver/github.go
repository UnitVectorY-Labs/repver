package repver

import (
	"fmt"
	"os/exec"
)

func CreateGitHubPullRequest() error {
	cmd := exec.Command("gh", "pr", "create", "--fill")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error creating GitHub pull request: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
