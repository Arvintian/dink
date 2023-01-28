package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func dupStdio(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
}

func locationKubeConfig(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		return strings.Replace(path, "~", home, 1)
	}
	return path
}

func kubeConfigNamespace(dinkPath string, kubeConfig string) string {
	runArgs := []string{"kubectl", "--kubeconfig", kubeConfig}
	runArgs = append(runArgs, "get", "svc", "kubernetes", "-o", "json")
	kubectl := exec.Command(dinkPath, runArgs...)
	bts, err := kubectl.Output()
	if err != nil {
		panic(err)
	}
	var kubeService corev1.Service
	if err := json.Unmarshal(bts, &kubeService); err != nil {
		panic(err)
	}
	return kubeService.Namespace
}

func forwardServer(ctx context.Context, dinkPath, kubeConfig, namespace, service string, port int) string {
	listenPort, err := getAvailablePort()
	if err != nil {
		panic(err)
	}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d", listenPort)
	runArgs := []string{"kubectl", "--kubeconfig", kubeConfig}
	runArgs = append(runArgs, "port-forward", fmt.Sprintf("svc/%s", service), "--address", "0.0.0.0", fmt.Sprintf("%d:%d", listenPort, port), "-n", namespace)
	kubectl := exec.CommandContext(ctx, dinkPath, runArgs...)
	go func() {
		kubectl.Run()
	}()
	for i := 0; i < 100; i++ {
		res, err := http.Get(fmt.Sprintf("%s/health", endpoint))
		if err != nil {
			<-time.After(5 * time.Millisecond)
			continue
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			panic(fmt.Errorf("port forward to dink server error %v", err))
		}
		break
	}
	return endpoint
}

func getAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
