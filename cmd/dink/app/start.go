package app

import (
	"fmt"
	"os"
	"path/filepath"

	req "github.com/imroc/req/v3"
	"github.com/spf13/cobra"
)

type StartCommand struct {
	KubeConfig      string `name:"kube-config" usage:"kube config file path" default:"~/.kube/config"`
	Namespace       string `name:"namespace" short:"n" usage:"target namespace"`
	ServerNamespace string `name:"server-namespace" usage:"dink server namespace" default:"dink"`
	ServerService   string `name:"server-service" usage:"dink server service" default:"dink-server"`
	ServerPort      int    `name:"server-port" usage:"dink server port" default:"8000"`
}

func (r *StartCommand) Run(cmd *cobra.Command, args []string) error {
	self, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}

	kubeConfig := locationKubeConfig(r.KubeConfig)

	if r.Namespace == "" {
		r.Namespace = kubeConfigNamespace(self, kubeConfig)
	}

	serverEndpoint := forwardServer(cmd.Context(), self, kubeConfig, r.ServerNamespace, r.ServerService, r.ServerPort)

	rsp, err := req.Put(fmt.Sprintf("%s/containers/%s/%s/start", serverEndpoint, r.Namespace, args[0]))
	if err != nil {
		return err
	}
	res, err := rsp.ToString()
	if err != nil {
		return err
	}
	fmt.Println(rsp.StatusCode, res)
	return nil
}
