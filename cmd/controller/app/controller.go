package app

import (
	"dink/pkg/controller"
	"dink/pkg/controller/controllers"
	"dink/pkg/k8s"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

type ControllerCommand struct {
	KubeConfig string `name:"kube-config" usage:"kube config file path"`
	Threads    int    `name:"threads" usage:"controller workers number" default:"2"`
	Root       string `name:"root" usage:"dink root path" default:"/var/lib/dink"`
	RunRoot    string `name:"run-root" usage:"dink runc root path" default:"/run/dink"`
	RuncRoot   string `name:"runc-root" usage:"dink runc root path" default:"/run/dink/runc"`
	DockerData string `name:"docker-data" usage:"docker data path" default:"/var/lib/dink/docker"`
	DockerHost string `name:"docker-host" usage:"docker daemon host" default:"tcp://127.0.0.1:2375"`
	AgentImage string `name:"agent-image" usage:"dink agent image"`
	NFSServer  string `name:"nfs-server" usage:"nfs server address"`
	NFSPath    string `name:"nfs-path" usage:"nfs mount path"`
	NFSOptions string `name:"nfs-options" usage:"nfs mount options" default:"vers=3,timeo=600,retrans=10,intr,nolock"`
}

func (r *ControllerCommand) Run(cmd *cobra.Command, args []string) error {
	controller.Config.Root = r.Root
	controller.Config.RunRoot = r.RunRoot
	controller.Config.RuncRoot = r.RuncRoot
	controller.Config.DockerData = r.DockerData
	controller.Config.DockerHost = r.DockerHost
	controller.Config.AgentImage = r.AgentImage
	controller.Config.NFSServer = r.NFSServer
	controller.Config.NFSPath = r.NFSPath
	controller.Config.NFSOptions = r.NFSOptions
	clientConfig, err := k8s.GetKubeConfig(r.KubeConfig)
	if err != nil {
		return err
	}

	dockerCli, err := client.NewClientWithOpts(client.WithHost(controller.Config.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	if _, err := dockerCli.Ping(cmd.Context()); err != nil {
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
