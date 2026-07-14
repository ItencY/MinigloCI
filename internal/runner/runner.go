package runner

import "context"

type RunRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type RunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

type Runner interface {
	Run(ctx context.Context, req RunRequest) (*RunResponse, error)
}
