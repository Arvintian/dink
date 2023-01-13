package main

import (
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type RemoveCommand struct {
	ID string `name:"id" usage:"container's id"`
}

func (r *RemoveCommand) Run(cmd *cobra.Command, args []string) error {
	dockerCli, err := client.NewClientWithOpts(client.WithHost(dink.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if err := dockerCli.ContainerRemove(cmd.Context(), r.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(dink.Data, "containers", r.ID)); err != nil {
		return err
	}

	klog.Infof("success removed container %s", r.ID)
	return nil
}
