package dockerpool

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Container struct {
	ID    string
	InUse bool
}

type WarmPool struct {
	mu         sync.Mutex
	containers map[string][]*Container
	client     *client.Client
}

func NewWarmPool(cli *client.Client) *WarmPool {
	wp := &WarmPool{
		client:     cli,
		containers: make(map[string][]*Container),
	}
	wp.CreateWarmpool("python", "python:3.12")
	wp.CreateWarmpool("golang", "golang:1.22")
	wp.CreateWarmpool("javascript", "node:20")
	return wp
}

func ReturnImage(lang string) string {
	var imagename string
	switch lang {
	case "golang":
		imagename = "golang:1.22"
	case "python":
		imagename = "python:3.12"
	case "javascript":
		imagename = "node:20"
	}
	return imagename
}

func (wp *WarmPool) CreateWarmpool(language string, image string) {
	for i := 0; i < 3; i++ {
		resp, err := wp.client.ContainerCreate(context.Background(), &container.Config{
			Image: image,
			Cmd:   []string{"sleep", "infinity"},
			Tty:   false,
		}, nil, nil, nil, "")

		if err != nil {
			fmt.Print("error creating container", err)
		}

		wp.client.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
		wp.containers[language] = append(wp.containers[language], &Container{
			ID:    resp.ID,
			InUse: false,
		})

	}
}

func (wp *WarmPool) AcquireFromWarmpool(language string) *Container {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	for _, c := range wp.containers[language] {
		if !c.InUse {
			c.InUse = true
			return c
		}
	}

	//spin up a new container
	resp, err := wp.client.ContainerCreate(context.Background(), &container.Config{
		Image: ReturnImage(language),
		Cmd:   []string{"sleep", "infinity"},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		fmt.Print("Error creating new containers while acquiring from warm pool")

	}
	NewC := &Container{
		ID:    resp.ID,
		InUse: false,
	}
	wp.client.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	wp.containers[language] = append(wp.containers[language], NewC)
	return NewC

}

func (wp *WarmPool) CopyCodeTocontainer(code string, containerID string, destination string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	Header := tar.Header{
		Size: int64(len(code)),
		Mode: 0644,
		Name: "mytarfile",
	}

	if err := tw.WriteHeader(&Header); err != nil {
		return fmt.Errorf("error while writing header:%v", err)
	}

	if _, err := tw.Write([]byte(code)); err != nil {
		return fmt.Errorf("error while writing file contents:%v", err)
	}

	if err := wp.client.CopyToContainer(context.Background(), containerID, destination, &buf, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}); err != nil {
		return fmt.Errorf("error copying code into container:%v", err)
	}

	return nil
}

func getFilextension(lang string) string {
	switch lang {
	case "python":
		return ".py"
	case "golang":
		return ".go"
	case "javascript":
		return ".js"
	default:
		return ""
	}
}

func getRunCommand(lang, file string) []string {
	switch lang {
	case "python":
		return []string{"python3", file}
	case "golang":
		return []string{"go", "run", file}
	case "javascript":
		return []string{"node", file}
	default:
		return []string{"echo", "unsupported language"}
	}
}

func (wp *WarmPool) RunCode(language, code string) (string, string, error) {
	c := wp.AcquireFromWarmpool(language)
	defer wp.ReleaseContainer(language, c.ID)
	if c == nil {
		return "", "", fmt.Errorf("no available containers")
	}

	ctx := context.Background()
	tmpDir := "/app"
	tmpFile := tmpDir + "/main" + getFilextension(language)

	execResp, err := wp.client.ContainerExecCreate(ctx, c.ID, container.ExecOptions{
		Cmd:          []string{"mkdir", "-p", tmpDir},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create mkdir exec: %v", err)
	}

	attachResp, err := wp.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to attach mkdir exec: %v", err)
	}
	defer attachResp.Close()

	if err := wp.client.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return "", "", fmt.Errorf("failed to start mkdir exec: %v", err)
	}

	_, _ = io.Copy(io.Discard, attachResp.Reader)

	err = wp.CopyCodeTocontainer(code, c.ID, tmpFile)
	if err != nil {
		return "", "", fmt.Errorf("error while copying code into the container %v", err)
	}

	cmd := getRunCommand(language, tmpFile)
	execResp, err = wp.client.ContainerExecCreate(ctx, c.ID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create exec command: %v", err)
	}

	attach, err := wp.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to attach exec: %v", err)
	}
	defer attach.Close()

	var stdout, stderr bytes.Buffer
	stdcopy.StdCopy(&stdout, &stderr, attach.Reader)
	return stdout.String(), stderr.String(), nil
}

func (wp *WarmPool) ReleaseContainer(language, ContainerID string) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	for _, container := range wp.containers[language] {
		if container.ID == ContainerID {
			container.InUse = false
			break
		}
	}

}
