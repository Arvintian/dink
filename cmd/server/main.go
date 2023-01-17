package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"dink/pkg/k8s"
	"dink/pkg/server"

	"github.com/Arvintian/go-utils/cmdutil"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Version = "0.0.0-dev"

type ServerCommand struct {
	Version    bool   `name:"version" usage:"show version"`
	KubeConfig string `name:"kube-config" usage:"kube config file path"`
	Bind       string `name:"bind" usage:"bind address" default:"0.0.0.0:8000"`
}

var dink ServerCommand

func (r *ServerCommand) Run(cmd *cobra.Command, args []string) error {
	if r.Version {
		return r.ShowVersion()
	}

	gin.DisableConsoleColor()

	client, err := k8s.GetClient(r.KubeConfig)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    r.Bind,
		Handler: server.CreateHTTPRouter(cmd.Context(), client),
	}

	go func() {
		<-cmd.Context().Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}()

	klog.Infof("server listen and serve on %s", r.Bind)
	err = srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (r *ServerCommand) ShowVersion() error {
	fmt.Println(Version)
	return nil
}

func main() {
	root := cmdutil.Command(&dink, cobra.Command{
		Long: "Run docker like container in kubernetes",
	})
	cmdutil.Main(root)
}
