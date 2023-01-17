// +k8s:deepcopy-gen=package
// +groupName=dink.io

package v1beta1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var CRDs = []struct {
	Name string
	CRD  *apiextensionsv1beta1.CustomResourceDefinition
}{
	{
		ContainerPluralName,
		&apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: ContainerPluralName + "." + GroupName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   SchemeGroupVersion.Group,
				Version: SchemeGroupVersion.Version,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural:     ContainerPluralName,
					Kind:       ContainerKind,
					ShortNames: []string{},
				},
			},
		},
	},
}
