package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	cli *client.Client
)

func InitializeDocker() {
	var err error

	cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(context.Background(), "joshjms/minijudge-agent:latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
}

func StartContainer(port int) (container.CreateResponse, error) {
	fmt.Println("Starting container")

	containerConfig := &container.Config{
		Image: "joshjms/minijudge-agent:latest",
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"3000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(port),
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return container.CreateResponse{}, err
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return container.CreateResponse{}, err
	}

	fmt.Println("Container started")

	return resp, nil
}

func StopContainer(resp container.CreateResponse, port int) error {
	fmt.Println("Stopping container")

	if err := cli.ContainerStop(context.Background(), resp.ID, container.StopOptions{}); err != nil {
		return err
	}

	if err := cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	portManager.ReleasePort(port)

	fmt.Println("Container stopped")
	return nil
}
