package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"dink/pkg/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type CreateCommand struct {
	Name       string   `name:"name" usage:"container's name"`
	Image      string   `name:"image" usage:"container's image"`
	Hostname   string   `name:"hostname" usage:"container's hostname"`
	Env        []string `name:"env" usage:"container's env"`
	WorkingDir string   `name:"workingdir" usage:"container's workingdir"`
	Entrypoint []string `name:"entrypoint" usage:"container's entrypoint"`
	Cmd        []string `name:"cmd" usage:"container's cmd"`
	TTY        bool     `name:"tty" usage:"container's tty open"`
}

func (r *CreateCommand) Run(cmd *cobra.Command, args []string) error {
	dockerCli, err := client.NewClientWithOpts(client.WithHost(dink.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	config := &container.Config{
		Image:      r.Image,
		Hostname:   r.Hostname,
		Env:        r.Env,
		WorkingDir: r.WorkingDir,
		Entrypoint: r.Entrypoint,
		Cmd:        r.Cmd,
		Tty:        r.TTY,
	}

	// image pull
	out, err := dockerCli.ImagePull(cmd.Context(), r.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// create container
	createRsp, err := dockerCli.ContainerCreate(cmd.Context(), config, nil, nil, nil, r.Name)
	if err != nil {
		return err
	}
	inspectRsp, err := dockerCli.ContainerInspect(cmd.Context(), createRsp.ID)
	if err != nil {
		return err
	}

	// create runc bundle
	containerHome := filepath.Join(dink.Root, "containers", createRsp.ID)
	if err := utils.CreateDir(containerHome, 0755); err != nil {
		return err
	}
	containerRunHome := filepath.Join(dink.RuncRoot, createRsp.ID)

	bts, err := json.Marshal(inspectRsp)
	if err != nil {
		return err
	}
	if err := utils.WriteBytesToFile(bts, filepath.Join(containerHome, "docker.json")); err != nil {
		return err
	}

	runcConfig := createRuntimeConfig()
	runcConfig.Process.Args = append(runcConfig.Process.Args, inspectRsp.Config.Entrypoint...)
	runcConfig.Process.Args = append(runcConfig.Process.Args, inspectRsp.Config.Cmd...)
	runcConfig.Process.Env = append(runcConfig.Process.Env, inspectRsp.Config.Env...)
	if inspectRsp.Config.WorkingDir != "" {
		runcConfig.Process.Cwd = inspectRsp.Config.WorkingDir
	}
	runcConfig.Hostname = inspectRsp.Config.Hostname
	runcConfig.Root.Path = filepath.Join(containerRunHome, "rootfs")

	bts, err = json.Marshal(runcConfig)
	if err != nil {
		return err
	}
	if err := utils.WriteBytesToFile(bts, filepath.Join(containerHome, "config.json")); err != nil {
		return err
	}

	klog.Infof("success created container %s", createRsp.ID)
	return nil
}
