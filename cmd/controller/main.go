package main

import (
	"fmt"

	"dink/cmd/controller/app"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"

type DinkCommand struct {
	Version bool `name:"version" usage:"show version"`
	app.ControllerCommand
}

func (r *DinkCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

	return r.ControllerCommand.Run(cmd, args)
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
