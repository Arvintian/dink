package main

import (
	"fmt"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

type AgentCommand struct {
	Version    bool   `name:"version" usage:"show version"`
	Data       string `name:"data" usage:"dink data path" default:"/var/lib/dink/data"`
	DockerData string `name:"docker-data" usage:"docker data path" default:"/var/lib/dink/docker"`
	DockerHost string `name:"docker-host" usage:"docker daemon host" default:"unix:///var/lib/dink/run/docker.sock"`
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
	root.AddCommand(cmdutil.Command(&CreateCommand{}, cobra.Command{
		Short: "Create a contianer",
		Long:  "Create a contianer",
	}))
	root.AddCommand(cmdutil.Command(&RemoveCommand{}, cobra.Command{
		Short: "Remove a contianer",
		Long:  "Remove a contianer",
	}))
	cmdutil.Main(root)
}
