package app

import (
	"context"
	"fmt"
	"time"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func ensureCRDsCreated(cfg *rest.Config) {
	clientset, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Failed to ensuring CRDs created, %v", err)
	}

	defer func() {
		if err != nil {
			klog.Fatalf("Failed to ensuring CRDs created, %v", err)
			return
		}
	}()

	crdcli := clientset.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range dinkv1beta1.CRDs {
		if _, err = crdcli.Create(context.TODO(), crd.CRD, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			klog.Fatalf("Failed to ensuring CRDs created, %v", err)
		}
		// wait for ready
		if err = waitCRDEstablished(clientset, crd.CRD.GetName()); err != nil {
			klog.Fatalf("Failed to ensuring CRDs created, %v", err)
		}
		klog.Infof("CRD %s created", crd.Name)
	}
}

func waitCRDEstablished(clientset *apiextensionsclient.Clientset, name string) error {
	// wait for CRD being established
	return wait.Poll(100*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					return false, fmt.Errorf("name conflict: %v", cond.Reason)
				}
			}
		}
		return false, err
	})
}
