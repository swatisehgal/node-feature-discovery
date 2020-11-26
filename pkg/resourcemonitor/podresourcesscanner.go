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
	"context"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"

	"k8s.io/api/core/v1"

	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
)

type PodResourcesScanner struct {
	args              Args
	podResourceClient podresourcesapi.PodResourcesListerClient
}

func NewPodResourcesScanner(args Args, podResourceClient podresourcesapi.PodResourcesListerClient) (ResourcesScanner, error) {
	resourcemonitorInstance := &PodResourcesScanner{
		args:              args,
		podResourceClient: podResourceClient,
	}
	if resourcemonitorInstance.args.Namespace != "" {
		log.Printf("watching namespace %q", resourcemonitorInstance.args.Namespace)
	} else {
		log.Printf("watching all namespaces")
	}

	return resourcemonitorInstance, nil
}

// isWatchable tells if the the given namespace should be watched.
func (resMon *PodResourcesScanner) isWatchable(podNamespace string) bool {
	if resMon.args.Namespace == "" {
		return true
	}
	//TODO:  add an explicit check for guaranteed pods
	return resMon.args.Namespace == podNamespace
}

// Scan gathers all the PodResources from the system, using the podresources API client.
func (resMon *PodResourcesScanner) Scan() ([]PodResources, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPodResourcesTimeout)
	defer cancel()

	//Pod Resource API client
	resp, err := resMon.podResourceClient.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Can't receive response: %v.Get(_) = _, %v", resMon.podResourceClient, err)
	}

	var podResData []PodResources

	for _, podResource := range resp.GetPodResources() {
		if !resMon.isWatchable(podResource.GetNamespace()) {
			log.Printf("SKIP pod %q\n", podResource.Name)
			continue
		}

		podRes := PodResources{
			Name:      podResource.GetName(),
			Namespace: podResource.GetNamespace(),
		}

		for _, container := range podResource.GetContainers() {
			var resCPUs []string
			for _, cpuID := range container.GetCpuIds() {
				resCPUs = append(resCPUs, fmt.Sprintf("%d", cpuID))
			}

			contRes := ContainerResources{
				Name: container.Name,
				Resources: []ResourceInfo{
					{
						Name: v1.ResourceCPU,
						Data: resCPUs,
					},
				},
			}

			for _, device := range container.GetDevices() {
				contRes.Resources = append(contRes.Resources, ResourceInfo{
					Name: v1.ResourceName(device.ResourceName),
					Data: device.DeviceIds,
				})
			}

			log.Printf("pod %q container %q contData=%s\n", podResource.GetName(), container.Name, spew.Sdump(contRes))
			podRes.Containers = append(podRes.Containers, contRes)
		}

		podResData = append(podResData, podRes)

	}

	return podResData, nil
}
