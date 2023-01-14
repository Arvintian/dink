package main

import (
	"dink/pkg/utils"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type StartCommand struct {
	ID string `name:"id" usage:"container's id"`
}

func (r *StartCommand) Run(cmd *cobra.Command, args []string) error {
	containerHome := filepath.Join(dink.Root, "containers", r.ID)
	containerRunHome := filepath.Join(dink.RuncRoot, r.ID)
	containerRunRootFS := filepath.Join(containerRunHome, "rootfs")
	if err := utils.CreateDir(containerRunRootFS, 0755); err != nil {
		return err
	}
	if err := utils.CopyFile(filepath.Join(containerHome, "config.json"), filepath.Join(containerRunHome, "config.json")); err != nil {
		return err
	}
	defer os.RemoveAll(containerRunHome)

	// mount rootfs info
	var dockerConfig types.ContainerJSON
	bts, err := os.ReadFile(filepath.Join(containerHome, "docker.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bts, &dockerConfig); err != nil {
		return err
	}
	graph := map[string]string{}
	for k, v := range dockerConfig.GraphDriver.Data {
		graph[k] = strings.ReplaceAll(v, "/var/lib/docker", dink.DockerData)
		if k == "LowerDir" {
			graph[k] = strings.Join(strings.Split(graph[k], ":")[1:], ":")
		}
	}

	mount := exec.Command("fuse-overlayfs", "-o",
		strings.Join([]string{"lowerdir=" + graph["LowerDir"], "upperDir=" + graph["UpperDir"], "workDir=" + graph["WorkDir"]}, ","),
		containerRunRootFS)
	ioOut, ioIn, err := os.Pipe()
	if err != nil {
		return err
	}
	mount.Stderr = ioIn
	mount.Stdout = ioIn
	go func() {
		io.Copy(os.Stdout, ioOut)
	}()

	if err := mount.Run(); err != nil {
		return err
	}
	klog.Infof("mount %s", mount.String())

	// shutdown
	<-cmd.Context().Done()

	if err := syscall.Unmount(containerRunRootFS, 0); err != nil {
		klog.Errorf("umount %s error %v", containerRunRootFS, err)
	}
	klog.Infof("unmount %s", containerRunRootFS)

	return nil
}
