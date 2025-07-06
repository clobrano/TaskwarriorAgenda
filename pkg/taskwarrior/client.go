package taskwarrior

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetTasks(filter []string) ([]Task, error) {
	args := append(filter, "export")
	cmd := exec.Command("task", args...)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("taskwarrior command failed: exit code %d, %s, stderr: %s",
				exitErr.ExitCode(), err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("taskwarrior command failed: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(output, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal taskwarrior output: %w", err)
	}
	return tasks, nil
}
