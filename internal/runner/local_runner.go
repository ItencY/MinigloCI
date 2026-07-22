package runner

import (
	"context"
)

type LocalRunner struct{}

func NewLocalRunner() *LocalRunner {
	return &LocalRunner{}
}

func (lr *LocalRunner) Run(ctx context.Context, req CommandRequest) (*CommandResult, error) {
	binaryPath, exists := allowedBinaries[req.Command]
	if !exists {
		return nil, ErrUnauthorizedCommand
	}
	return executeCmd(ctx, binaryPath, req.Args...)
}
