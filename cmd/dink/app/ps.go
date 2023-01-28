package app

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

type PsCommand struct {
	KubeConfig string `name:"kube-config" usage:"kube config file path" default:"~/.kube/config"`
	Namespace  string `name:"namespace" short:"n" usage:"target namespace"`
}

func (r *PsCommand) Run(cmd *cobra.Command, args []string) error {
	self, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	runArgs := []string{"kubectl"}
	if r.KubeConfig != "" {
		runArgs = append(runArgs, "--kubeconfig", locationKubeConfig(r.KubeConfig))
	}
	if r.Namespace != "" {
		runArgs = append(runArgs, "--namespace", r.Namespace)
	}
	runArgs = append(runArgs, "get", "container", "-o", "custom-columns=NAME:.metadata.name,STATE:.status.state,CREATED:.metadata.creationTimestamp")
	kubectl := exec.Command(self, runArgs...)
	dupStdio(kubectl)
	return kubectl.Run()
}
