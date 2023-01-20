package main

import (
	"fmt"

	controllercmd "dink/cmd/controller/app"
	servercmd "dink/cmd/server/app"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Version = "0.0.0-dev"

type DinkCommand struct {
	Version    bool   `name:"version" usage:"show version"`
	KubeConfig string `name:"kube-config" usage:"kube config file path"`
	Bind       string `name:"bind" usage:"server bind address" default:"0.0.0.0:8000"`
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

func (r *DinkCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

	controller := controllercmd.ControllerCommand{
		KubeConfig: r.KubeConfig,
		Threads:    r.Threads,
		Root:       r.Root,
		RunRoot:    r.RunRoot,
		RuncRoot:   r.RuncRoot,
		DockerData: r.DockerData,
		DockerHost: r.DockerHost,
		AgentImage: r.AgentImage,
		NFSServer:  r.NFSServer,
		NFSPath:    r.NFSPath,
		NFSOptions: r.NFSOptions,
	}
	go func() {
		if err := controller.Run(cmd, args); err != nil {
			klog.Fatal(err)
		}
	}()

	server := servercmd.ServerCommand{
		KubeConfig: r.KubeConfig,
		Bind:       r.Bind,
		Root:       r.Root,
		RunRoot:    r.RunRoot,
		RuncRoot:   r.RuncRoot,
		DockerData: r.DockerData,
		AgentImage: r.AgentImage,
		NFSServer:  r.NFSServer,
		NFSPath:    r.NFSPath,
		NFSOptions: r.NFSOptions,
	}
	go func() {
		if err := server.Run(cmd, args); err != nil {
			klog.Fatal(err)
		}
	}()

	<-cmd.Context().Done()
	return nil
}

func (r *DinkCommand) ShowVersion() error {
	fmt.Println(Version)
	return nil
}

func main() {
	root := cmdutil.Command(&DinkCommand{}, cobra.Command{
		Long: "Run docker like container in kubernetes",
	})
	cmdutil.Main(root)
}
