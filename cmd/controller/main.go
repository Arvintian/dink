package main

import (
	"dink/pkg/controller/controllers"
	"dink/pkg/k8s"
	"fmt"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

type ControllerCommand struct {
	Version    bool   `name:"version" usage:"show version"`
	KubeConfig string `name:"kube-config" usage:"kube config file path"`
	Threads    int    `name:"threads" usage:"controller workers number" default:"5"`
}

var dink ControllerCommand

func (r *ControllerCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

	client, err := k8s.GetClient(r.KubeConfig)
	if err != nil {
		return err
	}

	podController := controllers.NewPodController(client, client)

	podController.Run(r.Threads, cmd.Context().Done())

	<-cmd.Context().Done()
	return nil
}

func (r *ControllerCommand) ShowVersion() error {
	fmt.Println(Version)
	return nil
}

func main() {
	root := cmdutil.Command(&dink, cobra.Command{
		Long: "Run docker like container in kubernetes",
	})
	cmdutil.Main(root)
}