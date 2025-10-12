package dockerutil

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerCLI struct {
	Client *client.Client
}

func NewDockerCLI() (*DockerCLI, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerCLI{Client: cli}, nil
}

func (d *DockerCLI) RunPython() (string, string, error) {
	ctx := context.Background()

	// Pull Python image
	reader, err := d.Client.ImagePull(ctx, "docker.io/library/python", image.PullOptions{})
	if err != nil {
		return "", "", err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader) // consume pull output

	// Create container
	resp, err := d.Client.ContainerCreate(ctx, &container.Config{
		Image: "python",
		Cmd:   []string{"python", "-c", "print('hello world')"},
	}, nil, nil, nil, "")
	if err != nil {
		return "", "", err
	}

	// Start container
	if err := d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", err
	}

	// Wait for container to finish
	statusCh, errCh := d.Client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", "", err
		}
	case <-statusCh:
	}

	// Capture logs
	out, err := d.Client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", "", err
	}
	defer out.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, out); err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
