/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const nodeResourceTopologiesName = "noderesourcetopologies.topology.node.k8s.io"

func newNodeResourceTopologies() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeResourceTopologiesName,
			Annotations: map[string]string{
				"api-approved.kubernetes.io": "https://github.com/kubernetes/enhancements/pull/1870",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "topology.node.k8s.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "noderesourcetopologies",
				Singular: "noderesourcetopology",
				ShortNames: []string{
					"node-res-topo",
				},
				Kind: "NodeResourceTopology",
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha1",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"topologyPolicies": {
									Type: "array",
									Items: &apiextensionsv1.JSONSchemaPropsOrArray{
										Schema: &apiextensionsv1.JSONSchemaProps{
											Type: "string",
										},
									},
								},
							},
						},
					},
					Served:  true,
					Storage: true,
				},
			},
		},
	}
}

func CreateNodeResourceTopologies(extClient extclient.Interface) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), nodeResourceTopologiesName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if err == nil {
		return crd, nil
	}

	crd, err = extClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), newNodeResourceTopologies(), metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return crd, nil
}
