package dockerutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func (d *DockerCLI) RunCode(lang string, code string) (string, string, error) {
	ctx := context.Background()
	var imageName, fileName, cmd string

	// Pick image and command based on language
	switch lang {
	case "python":
		imageName = "python:3.12"
		fileName = "main.py"
		cmd = fmt.Sprintf("python %s", fileName)

	case "goLang":
		imageName = "golang:1.22"
		fileName = "main.go"
		cmd = fmt.Sprintf("go run %s", fileName)

	case "javascript":
		imageName = "node:20"
		fileName = "main.js"
		cmd = fmt.Sprintf("node %s", fileName)

	default:
		return "", "", fmt.Errorf("unsupported language: %s", lang)
	}

	// Create a temp directory for the code
	tmpDir, err := os.MkdirTemp("", "runner-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, fileName)
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return "", "", err
	}

	// Pull image if not present
	reader, err := d.Client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", "", err
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	// Create container
	resp, err := d.Client.ContainerCreate(ctx, &container.Config{
		Image:      imageName,
		Cmd:        []string{"bash", "-c", cmd},
		Tty:        false,
		WorkingDir: "/app",
	}, &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/app", tmpDir)},
	}, nil, nil, "")
	if err != nil {
		return "", "", err
	}

	// Start container
	if err := d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", err
	}

	// Wait until container exits
	statusCh, errCh := d.Client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", "", err
		}
	case <-statusCh:
	}

	// Get logs
	out, err := d.Client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", "", err
	}
	defer out.Close()

	var stdout, stderr bytes.Buffer
	stdcopy.StdCopy(&stdout, &stderr, out)

	return stdout.String(), stderr.String(), nil
}
