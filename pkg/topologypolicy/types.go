/*
Copyright 2021 The Kubernetes Authors.
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

package topologypolicy

// TopologyManagerPolicy constants which represent the current configuration
// for Topology manager policy and Topology manager scope in Kubelet config
type TopologyManagerPolicy string

const (
	SingleNumaContainerScope TopologyManagerPolicy = "SingleNUMANodeContainerLevel"
	SingleNumaPodScope       TopologyManagerPolicy = "SingleNUMANodePodLevel"
	Restricted               TopologyManagerPolicy = "Restricted"
	BestEffort               TopologyManagerPolicy = "BestEffort"
	None                     TopologyManagerPolicy = "None"
)

// K8sTopologyPolicies are resource allocation policies constants
type K8sTopologyManagerPolicies string

const (
	singleNumaNode K8sTopologyManagerPolicies = "single-numa-node"
	restricted     K8sTopologyManagerPolicies = "restricted"
	bestEffort     K8sTopologyManagerPolicies = "best-effort"
	none           K8sTopologyManagerPolicies = "none"
)

// K8sTopologyScopes are constants which defines the granularity
// at which you would like resource alignment to be performed.
type K8sTopologyManagerScopes string

const (
	pod       K8sTopologyManagerScopes = "pod"
	container K8sTopologyManagerScopes = "container"
)
