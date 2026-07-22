package runner

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUnauthorizedCommand = errors.New("unauthorized command")
	ErrCommandTimeout      = errors.New("command timed out")
	ErrExecutionFailed     = errors.New("command execution failed")
)

const (
	maxOutputSize  = 1024 * 1024     // 1 MB for memory protection
	commandTimeout = 5 * time.Second // Command execution timeout
)

// Allowlist of permitted utilities (RCE protection)
var allowedBinaries = map[string]string{
	"echo": "/usr/bin/echo",
}

type CommandRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

type Runner interface {
	Run(ctx context.Context, req CommandRequest) (*CommandResult, error)
}
