package main

import (
	"dink/pkg/utils"
	"encoding/json"
	"fmt"
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
	if err := utils.CreateDir(dink.Root, 0755); err != nil {
		return err
	}
	nfs := exec.Command("mount", "-t", "nfs", "-o", dink.NFSOptions, fmt.Sprintf("%s:%s", dink.NFSServer, dink.NFSPath), dink.Root)
	dupStdio(nfs)
	if err := nfs.Run(); err != nil {
		return err
	}
	defer func() {
		if err := syscall.Unmount(dink.Root, 0); err != nil {
			klog.Errorf("umount %s error %v", dink.Root, err)
		}
		klog.Infof("unmount %s", dink.Root)
	}()

	containerHome := filepath.Join(dink.Root, "containers", r.ID)
	containerRunHome := filepath.Join(dink.RunRoot, r.ID)
	containerRunRootFS := filepath.Join(containerRunHome, "rootfs")
	if err := utils.CreateDir(containerRunRootFS, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(containerRunHome)

	var dockerConfig types.ContainerJSON
	bts, err := os.ReadFile(filepath.Join(containerHome, "docker.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bts, &dockerConfig); err != nil {
		return err
	}

	// runc config
	bts, err = os.ReadFile(filepath.Join(containerHome, "config.json"))
	if err != nil {
		return err
	}
	if err := utils.WriteBytesToFile(bts, filepath.Join(containerRunHome, "config.json")); err != nil {
		return err
	}

	// mount rootfs
	graph := map[string]string{}
	for k, v := range dockerConfig.GraphDriver.Data {
		graph[k] = strings.ReplaceAll(v, "/var/lib/docker", dink.DockerData)
		if k == "LowerDir" {
			graph[k] = strings.Join(strings.Split(graph[k], ":")[1:], ":")
		}
	}

	mount := exec.Command("fuse-overlayfs", "-o",
		strings.Join([]string{"lowerdir=" + graph["LowerDir"], "upperdir=" + graph["UpperDir"], "workdir=" + graph["WorkDir"]}, ","),
		containerRunRootFS)
	dupStdio(mount)
	if err := mount.Run(); err != nil {
		return err
	}

	klog.Infof("mount %s", containerRunRootFS)
	defer func() {
		if err := syscall.Unmount(containerRunRootFS, 0); err != nil {
			klog.Errorf("umount %s error %v", containerRunRootFS, err)
		}
		klog.Infof("unmount %s", containerRunRootFS)
	}()

	// start container
	runc := exec.Command("runc", "--root", dink.RuncRoot, "run", "--bundle", containerRunHome, r.ID)
	dupStdio(runc)
	klog.Infof("start container %s", r.ID)
	if err := runc.Run(); err != nil {
		return err
	}

	return nil
}

func dupStdio(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
}
