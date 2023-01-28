package main

import (
	"fmt"

	"dink/cmd/dink/app"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
	kubectl "k8s.io/kubectl/pkg/cmd"
)

var Version = "0.0.0-dev"

type DinkCommand struct {
	Version bool `name:"version" usage:"show version"`
}

func (r *DinkCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

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
	root.AddCommand(kubectl.NewDefaultKubectlCommand())
	root.AddCommand(cmdutil.Command(&app.PsCommand{}, cobra.Command{
		Short: "List contianers",
		Long:  "List contianers",
	}))
	root.AddCommand(cmdutil.Command(&app.StartCommand{}, cobra.Command{
		Short: "Start contianer",
		Long:  "Start contianer",
	}))
	root.AddCommand(cmdutil.Command(&app.StopCommand{}, cobra.Command{
		Short: "Stop contianer",
		Long:  "Stop contianer",
	}))
	root.AddCommand(cmdutil.Command(&app.ExecCommand{}, cobra.Command{
		Short: "Exec command in contianer",
		Long:  "Exec command in contianer",
	}))
	cmdutil.Main(root)
}
