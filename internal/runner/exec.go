package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

func executeCmd(ctx context.Context, name string, args ...string) (*CommandResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, name, args...)
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
	return &CommandResult{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
	}, nil
}
