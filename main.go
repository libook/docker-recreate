package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var (
	shortUpdateImgFlag = flag.Bool("u", false, "Update the image before recreating the container")
	updateImgFlag      = flag.Bool("update", false, "Update the image before recreating the container")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	containerID := flag.Arg(0)

	// Combine results from full flag and short flag
	shouldUpdateImageFlag := *shortUpdateImgFlag || *updateImgFlag

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	container, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}

	networks := container.NetworkSettings.Networks

	networkConfig := network.NetworkingConfig{
		EndpointsConfig: networks,
	}

	// network, err := cli.NetworkInspect(ctx, networkID)

	isRunning := container.State.Running || container.State.Paused
	imageName := container.Config.Image

	if isRunning {
		fmt.Println("Stopping container ...")

		err = cli.ContainerStop(ctx, containerID, nil)
		if err != nil {
			panic(err)
		}
	}

	if !container.HostConfig.AutoRemove {
		fmt.Println("Removing container ...")

		err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
		if err != nil {
			panic(err)
		}
	}

	if shouldUpdateImageFlag {
		fmt.Println("Updating container image ...")
		out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		defer out.Close()
	}

	fmt.Println("Recreating container ...")
	createdContainer, err := cli.ContainerCreate(ctx, container.Config, container.HostConfig, &networkConfig, container.Name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting container ...")
	err = cli.ContainerStart(ctx, createdContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

}
