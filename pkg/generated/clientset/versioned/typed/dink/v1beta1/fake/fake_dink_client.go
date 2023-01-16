/*
Copyright The Dink Authors.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "dink/pkg/generated/clientset/versioned/typed/dink/v1beta1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeDinkV1beta1 struct {
	*testing.Fake
}

func (c *FakeDinkV1beta1) Containers(namespace string) v1beta1.ContainerInterface {
	return &FakeContainers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeDinkV1beta1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}