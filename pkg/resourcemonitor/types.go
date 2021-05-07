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

package resourcemonitor

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	topologyv1alpha1 "github.com/swatisehgal/topologyapi/pkg/apis/topology/v1alpha1"
)

type Args struct {
	PodResourceSocketPath string
	SleepInterval         time.Duration
	Namespace             string
	SysfsRoot             string
	KubeletConfigFile     string
}

type ResourceInfo struct {
	Name corev1.ResourceName
	Data []string
}

type ContainerResources struct {
	Name      string
	Resources []ResourceInfo
}

type PodResources struct {
	Name       string
	Namespace  string
	Containers []ContainerResources
}

type ResourcesScanner interface {
	Scan() ([]PodResources, error)
}

type ResourcesAggregator interface {
	Aggregate(podResData []PodResources) map[string]*topologyv1alpha1.Zone
}
