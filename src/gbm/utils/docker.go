package utils

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os/exec"
	"reflect"
)

type Container types.Container

func GetContainerByName(containerName string) Container {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		for _, n := range c.Names {
			if containerName == n[1:] {
				return Container(c)
			}
		}
	}

	return Container{}
}

func (c Container) IsEmpty() bool {
	return reflect.DeepEqual(c, Container{})
}

// TODO improve: execute command in docker (using docker sdk)
func Exec(ctx context.Context, containerId string, command []string) ([]byte, error) {
	cmd := append([]string {"docker", "exec", containerId}, command...)
	executed := exec.Command(cmd[0], cmd[1:]...)
	out, err := executed.Output()
	return out, err
}
