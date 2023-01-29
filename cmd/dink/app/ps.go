package app

import (
	"fmt"
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
	selfExecPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	self, err := filepath.Abs(selfExecPath)
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
	columns := "NAME:.metadata.name,IMAGE:.spec.template.image,IP:.status.podStatus.podIP,HOST:.status.podStatus.hostIP,CREATED:.metadata.creationTimestamp,STATE:.status.state"
	runArgs = append(runArgs, "get", "container", "-o", fmt.Sprintf("custom-columns=%s", columns))
	kubectl := exec.Command(self, runArgs...)
	dupStdio(kubectl)
	return kubectl.Run()
}
