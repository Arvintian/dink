package app

import (
	"dink/pkg/server/handlers"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	req "github.com/imroc/req/v3"
	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

type CreateCommand struct {
	KubeConfig      string   `name:"kube-config" usage:"kube config file path" default:"~/.kube/config"`
	Namespace       string   `name:"namespace" short:"n" usage:"target namespace"`
	ServerNamespace string   `name:"server-namespace" usage:"dink server namespace" default:"dink"`
	ServerService   string   `name:"server-service" usage:"dink server service" default:"dink-server"`
	ServerPort      int      `name:"server-port" usage:"dink server port" default:"8000"`
	Name            string   `name:"name" usage:"assign a name to the container"`
	HostName        string   `name:"hostname" usage:"container host name"`
	RestartPolicy   string   `name:"restart" usage:"restart policy to apply when a container exits" default:"Never"`
	Env             []string `name:"env" short:"e" usage:"set container environment variables"`
	Workdir         string   `name:"workdir" short:"w" usage:"working directory inside the container"`
	Entrypoint      string   `name:"entrypoint" usage:"overwrite the default ENTRYPOINT of the image"`
	Interactive     bool     `name:"interactive" short:"i" usage:"keep container's STDIN open"`
	TTY             bool     `name:"tty" short:"t" usage:"allocate a pseudo-TTY"`
}

func (r *CreateCommand) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("requires at least 1 argument")
	}

	selfExecPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	self, err := filepath.Abs(selfExecPath)
	if err != nil {
		return err
	}

	kubeConfig := locationKubeConfig(r.KubeConfig)

	if r.Namespace == "" {
		r.Namespace = kubeConfigNamespace(self, kubeConfig)
	}

	serverEndpoint, serverProxy := forwardServer(cmd.Context(), self, kubeConfig, r.ServerNamespace, r.ServerService, r.ServerPort)

	defer func() {
		serverProxy.Process.Kill()
		serverProxy.Wait()
	}()

	name := namesgenerator.GetRandomName(0)
	name = strings.ReplaceAll(name, "_", "-")
	if r.Name != "" {
		name = r.Name
	}
	command := []string{}
	if len(args) > 1 {
		command = args[1:]
	}
	entrypoint := []string{}
	if r.Entrypoint != "" {
		entrypoint = append(entrypoint, r.Entrypoint)
	}

	envs := []corev1.EnvVar{}
	for _, s := range r.Env {
		if !strings.Contains(s, "=") {
			continue
		}
		sp := strings.SplitN(s, "=", 2)
		envs = append(envs, corev1.EnvVar{
			Name:  sp[0],
			Value: sp[1],
		})
	}

	payload := handlers.ContainerConfig{
		Image:         args[0],
		HostName:      r.HostName,
		RestartPolicy: r.RestartPolicy,
		Env:           envs,
		WorkingDir:    r.Workdir,
		Entrypoint:    entrypoint,
		Cmd:           command,
		Stdin:         r.Interactive,
		TTY:           r.TTY,
	}
	rsp, err := req.C().R().SetBodyJsonMarshal(payload).Post(fmt.Sprintf("%s/containers/%s/%s", serverEndpoint, r.Namespace, name))
	if err != nil {
		return err
	}

	res, err := rsp.ToString()
	if err != nil {
		return err
	}
	if rsp.StatusCode == 200 {
		fmt.Printf("Creating container %s\n", name)
	} else {
		fmt.Println(res)
		serverProxy.Process.Kill()
		serverProxy.Wait()
		os.Exit(-1)
	}
	return nil
}
