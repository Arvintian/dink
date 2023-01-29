package app

import (
	"dink/pkg/apis/dink/v1beta1/template"
	"dink/pkg/k8s"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogsCommand struct {
	KubeConfig string `name:"kube-config" usage:"kube config file path" default:"~/.kube/config"`
	Namespace  string `name:"namespace" short:"n" usage:"target namespace"`
	Tail       int    `name:"tail" usage:"lines of recent log file to display" default:"-1"`
	Follow     bool   `name:"follow" short:"f" usage:"specify if the logs should be streamed"`
}

func (r *LogsCommand) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("requires at least 1 argument")
	}

	self, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}

	kubeConfig := locationKubeConfig(r.KubeConfig)

	if r.Namespace == "" {
		r.Namespace = kubeConfigNamespace(self, kubeConfig)
	}

	client, err := k8s.GetClient(kubeConfig)
	if err != nil {
		return err
	}

	container, err := client.DinkV1beta1().Containers(r.Namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		fmt.Println("container not found")
		return nil
	}

	if dinkv1beta1.IsFinalState(container.Status.State) || container.Status.PodStatus == nil {
		fmt.Println("container not is running")
		return nil
	}

	thePod := template.GetPodName(container)

	runArgs := []string{"kubectl"}
	if r.KubeConfig != "" {
		runArgs = append(runArgs, "--kubeconfig", kubeConfig)
	}
	if r.Namespace != "" {
		runArgs = append(runArgs, "--namespace", r.Namespace)
	}
	runArgs = append(runArgs, "logs", "--tail", fmt.Sprintf("%d", r.Tail))
	if r.Follow {
		runArgs = append(runArgs, "-f")
	}
	runArgs = append(runArgs, "-c", "dink-agent", thePod)

	kubectl := exec.CommandContext(cmd.Context(), self, runArgs...)
	dupStdio(kubectl)
	if err := kubectl.Run(); err != nil {
		fmt.Printf("%v", err)
	}
	defer kubectl.Process.Kill()
	return nil
}
