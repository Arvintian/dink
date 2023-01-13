package main

import (
	"fmt"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Version = "0.0.0-dev"

type DinkCommand struct {
	Host    string `name:"host" usage:"bind host" default:"0.0.0.0"`
	Port    int    `name:"port" usage:"bind port" default:"8000"`
	Version bool   `name:"version" usage:"show version"`
}

func (r *DinkCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}
	klog.Infof("Run on %s:%d", r.Host, r.Port)
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
