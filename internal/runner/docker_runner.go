package runner

import (
	"context"
)

type DockerRunner struct {
	Image string
}

func NewDockerRunner(image string) *DockerRunner {
	return &DockerRunner{
		Image: image,
	}
}

func (r *DockerRunner) Run(ctx context.Context, req CommandRequest) (*CommandResult, error) {
	image := r.Image
	if image == "" {
		image = "alpine"
	}

	dockerArgs := []string{"run", "--rm", image, req.Command}
	dockerArgs = append(dockerArgs, req.Args...)

	return executeCmd(ctx, "docker", dockerArgs...)
}
