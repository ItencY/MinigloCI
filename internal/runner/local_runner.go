package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

var (
	ErrUnauthorizedCommand = errors.New("unauthorized command")
	ErrCommandTimeout      = errors.New("command timed out")
	ErrExecutionFailed     = errors.New("command execution failed")
)

// Allowlist of permitted utilities (RCE protection)
var allowedBinaries = map[string]string{
	"echo": "/usr/bin/echo",
}

const (
	maxOutputSize  = 1024 * 1024     // 1 MB for memory protection
	commandTimeout = 5 * time.Second // Command execution timeout
)

type LocalRunner struct{}

func NewLocalRunner() *LocalRunner {
	return &LocalRunner{}
}

func (lr *LocalRunner) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	binaryPath, exists := allowedBinaries[req.Command]
	if !exists {
		return nil, ErrUnauthorizedCommand
	}
	runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	// #nosec G204
	cmd := exec.CommandContext(runCtx, binaryPath, req.Args...)
	stdoutBuf, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe failed: %w", err)
	}
	stderrBuf, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe failed: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("command start failed: %w", err)
	}
	stdoutBytes, _ := io.ReadAll(io.LimitReader(stdoutBuf, maxOutputSize))
	stderrBytes, _ := io.ReadAll(io.LimitReader(stderrBuf, maxOutputSize))
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrCommandTimeout
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, ErrExecutionFailed
		}
	}
	return &RunResponse{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
	}, nil
}
