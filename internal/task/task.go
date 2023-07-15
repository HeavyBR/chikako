package task

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

type Task struct {
	ID            uuid.UUID
	Name          string
	State         State
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	Cmd           []string
	Image         string
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

type Docker struct {
	Client      *client.Client
	Config      Config
	ContainerID string
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerID string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()

	// Pull image
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, reader)

	// Create container

	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}
	r := container.Resources{
		Memory: d.Config.Memory,
	}

	cc := container.Config{
		Image: d.Config.Image,
		Env:   d.Config.Env,
	}

	hc := container.HostConfig{
		Resources:       r,
		RestartPolicy:   rp,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container %s: %v\n", d.Config.Name, err)
		return DockerResult{Error: err}
	}

	d.ContainerID = resp.ID

	// Start container
	if err = d.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("Error starting container %s: %v\n", d.Config.Name, err)
		return DockerResult{Error: err}
	}

	// Copy logs
	out, err := d.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("Error getting logs from container %s: %v\n", d.Config.Name, err)
		return DockerResult{Error: err}
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{ContainerID: resp.ID, Action: "start", Result: "success", Error: nil}
}

func (d *Docker) Stop() DockerResult {
	ctx := context.Background()
	log.Printf("Stopping container %s\n", d.ContainerID)
	// Stop container
	if err := d.Client.ContainerStop(ctx, d.ContainerID, container.StopOptions{}); err != nil {
		log.Printf("Error stopping container %s: %v\n", d.Config.Name, err)
		return DockerResult{Error: err}
	}

	// Remove container
	if err := d.Client.ContainerRemove(ctx, d.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	}); err != nil {
		log.Printf("Error removing container %s: %v\n", d.Config.Name, err)
		return DockerResult{Error: err}
	}

	return DockerResult{ContainerID: d.ContainerID, Action: "stop", Result: "success", Error: nil}
}
