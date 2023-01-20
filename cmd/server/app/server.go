package app

import (
	"context"
	"net/http"
	"time"

	"dink/pkg/k8s"
	"dink/pkg/server"
	"dink/pkg/server/handlers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

type ServerCommand struct {
	KubeConfig string `name:"kube-config" usage:"kube config file path"`
	Bind       string `name:"bind" usage:"bind address" default:"0.0.0.0:8000"`
	Root       string `name:"root" usage:"dink root path" default:"/var/lib/dink"`
	RunRoot    string `name:"run-root" usage:"dink runc root path" default:"/run/dink"`
	RuncRoot   string `name:"runc-root" usage:"dink runc root path" default:"/run/dink/runc"`
	DockerData string `name:"docker-data" usage:"docker data path" default:"/var/lib/dink/docker"`
	AgentImage string `name:"agent-image" usage:"dink agent image"`
	NFSServer  string `name:"nfs-server" usage:"nfs server address"`
	NFSPath    string `name:"nfs-path" usage:"nfs mount path"`
}

func (r *ServerCommand) Run(cmd *cobra.Command, args []string) error {
	handlers.Config.Root = r.Root
	handlers.Config.RunRoot = r.RunRoot
	handlers.Config.RuncRoot = r.RuncRoot
	handlers.Config.DockerData = r.DockerData
	handlers.Config.AgentImage = r.AgentImage
	handlers.Config.NFSServer = r.NFSServer
	handlers.Config.NFSPath = r.NFSPath
	client, err := k8s.GetClient(r.KubeConfig)
	if err != nil {
		return err
	}
	handlers.Config.Client = client
	gin.DisableConsoleColor()

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
