package dockerpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
