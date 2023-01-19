package main

import (
	"fmt"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

type AgentCommand struct {
	Version    bool   `name:"version" usage:"show version"`
	Root       string `name:"root" usage:"dink root path" default:"/var/lib/dink"`
	RunRoot    string `name:"run-root" usage:"dink root path" default:"/run/dink"`
	RuncRoot   string `name:"runc-root" usage:"dink runc root path" default:"/run/dink/runc"`
	DockerData string `name:"docker-data" usage:"docker data path" default:"/var/lib/dink/docker"`
	NFSServer  string `name:"nfs-server" usage:"nfs server address"`
	NFSPath    string `name:"nfs-path" usage:"nfs mount path"`
}

var dink AgentCommand

func (r *AgentCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}
	return nil
}

func (r *AgentCommand) ShowVersion() error {
	fmt.Println(Version)
	return nil
}

func main() {
	root := cmdutil.Command(&dink, cobra.Command{
		Long: "Run docker like container in kubernetes",
	})
	root.AddCommand(cmdutil.Command(&StartCommand{}, cobra.Command{
		Short: "Start a contianer",
		Long:  "Start a contianer",
	}))
	cmdutil.Main(root)
}
