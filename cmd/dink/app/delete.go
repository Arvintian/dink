package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	req "github.com/imroc/req/v3"
	"github.com/spf13/cobra"
)

type DeleteCommand struct {
	KubeConfig      string `name:"kube-config" usage:"kube config file path" default:"~/.kube/config"`
	Namespace       string `name:"namespace" short:"n" usage:"target namespace"`
	ServerNamespace string `name:"server-namespace" usage:"dink server namespace" default:"dink"`
	ServerService   string `name:"server-service" usage:"dink server service" default:"dink-server"`
	ServerPort      int    `name:"server-port" usage:"dink server port" default:"8000"`
}

func (r *DeleteCommand) Run(cmd *cobra.Command, args []string) error {
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

	rsp, err := req.Delete(fmt.Sprintf("%s/containers/%s/%s", serverEndpoint, r.Namespace, args[0]))
	if err != nil {
		return err
	}
	res, err := rsp.ToString()
	if err != nil {
		return err
	}
	if rsp.StatusCode == 200 {
		fmt.Printf("Deleting container %s\n", args[0])
	} else {
		fmt.Println(res)
		serverProxy.Process.Kill()
		serverProxy.Wait()
		os.Exit(-1)
	}
	return nil
}
