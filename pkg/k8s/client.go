package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "dink/pkg/generated/clientset/versioned"
	dinkscheme "dink/pkg/generated/clientset/versioned/scheme"
	dinkv1beta1 "dink/pkg/generated/clientset/versioned/typed/dink/v1beta1"
)

var Scheme = runtime.NewScheme()

func init() {
	_ = dinkscheme.AddToScheme(Scheme)
	_ = scheme.AddToScheme(Scheme)
}

type Interface interface {
	kubernetes.Interface
	DinkV1beta1() dinkv1beta1.DinkV1beta1Interface
}

type Clientset struct {
	*kubernetes.Clientset
	dinkClient *clientset.Clientset
}

func (c *Clientset) DinkV1beta1() dinkv1beta1.DinkV1beta1Interface {
	return c.dinkClient.DinkV1beta1()
}

var _ Interface = (*Clientset)(nil)

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	dinkClient, err := clientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &Clientset{
		Clientset:  kubeClient,
		dinkClient: dinkClient,
	}, nil
}

// GetClient creates a client for k8s cluster
func GetClient(kubeConfigPath string) (Interface, error) {
	config, err := GetKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return NewForConfig(config)
}

// GetRateLimitClient creates a client for k8s cluster with custom defined qps and burst.
func GetRateLimitClient(kubeConfigPath string, qps float32, burst int) (Interface, error) {
	config, err := GetKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	if qps > 0.0 {
		config.QPS = qps
	}

	if burst > 0 {
		config.Burst = burst
	}

	return NewForConfig(config)
}

func GetKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
