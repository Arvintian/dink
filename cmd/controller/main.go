package main

import (
	"dink/pkg/controller"
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
	Threads    int    `name:"threads" usage:"controller workers number" default:"2"`
	Root       string `name:"root" usage:"dink root path" default:"/var/lib/dink"`
	RuncRoot   string `name:"runc-root" usage:"dink runc root path" default:"/run/dink/runc"`
	DockerData string `name:"docker-data" usage:"docker data path" default:"/var/lib/dink/docker"`
	DockerHost string `name:"docker-host" usage:"docker daemon host" default:"tcp://127.0.0.1:2375"`
}

var dink ControllerCommand

func (r *ControllerCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

	controller.Config.Root = dink.Root
	controller.Config.RuncRoot = dink.RuncRoot
	controller.Config.DockerData = dink.DockerData
	controller.Config.DockerHost = dink.DockerHost

	clientConfig, err := k8s.GetKubeConfig(r.KubeConfig)
	if err != nil {
		return err
	}

	ensureCRDsCreated(clientConfig)

	client, err := k8s.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	containerController := controllers.NewContainerController(cmd.Context(), client)
	go containerController.Run(r.Threads, cmd.Context().Done())

	podController := controllers.NewPodController(cmd.Context(), client)
	go podController.Run(r.Threads, cmd.Context().Done())

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
