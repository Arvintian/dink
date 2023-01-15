package main

import (
	"dink/pkg/utils"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type StartCommand struct {
	ID         string   `name:"id" usage:"container's id"`
	Env        []string `name:"env" usage:"container's env"`
	WorkingDir string   `name:"workingdir" usage:"container's workingdir"`
	Cmd        []string `name:"cmd" usage:"container's cmd"`
	TTY        bool     `name:"tty" usage:"container's tty open"`
}

func (r *StartCommand) Run(cmd *cobra.Command, args []string) error {
	containerHome := filepath.Join(dink.Root, "containers", r.ID)
	containerRunHome := filepath.Join(dink.RuncRoot, r.ID)
	containerRunRootFS := filepath.Join(containerRunHome, "rootfs")
	if err := utils.CreateDir(containerRunRootFS, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(containerRunHome)

	// runc config
	var runcConfig specs.Spec
	bts, err := os.ReadFile(filepath.Join(containerHome, "config.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bts, &runcConfig); err != nil {
		return err
	}
	if len(r.Cmd) > 0 {
		runcConfig.Process.Args = r.Cmd
	}
	runcConfig.Process.Env = append(runcConfig.Process.Env, r.Env...)
	if r.WorkingDir != "" {
		runcConfig.Process.Cwd = r.WorkingDir
	}
	if r.TTY {
		runcConfig.Process.Terminal = true
	}
	bts, err = json.Marshal(runcConfig)
	if err != nil {
		return err
	}
	if err := utils.WriteBytesToFile(bts, filepath.Join(containerRunHome, "config.json")); err != nil {
		return err
	}

	// mount rootfs
	var dockerConfig types.ContainerJSON
	bts, err = os.ReadFile(filepath.Join(containerHome, "docker.json"))
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
	runc := exec.Command("runc", "--root", filepath.Join(dink.RuncRoot, "runc"), "run", "--bundle", containerRunHome, r.ID)
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
