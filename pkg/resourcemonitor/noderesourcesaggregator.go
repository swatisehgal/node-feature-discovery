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
	"time"

	v1alpha1 "github.com/swatisehgal/topologyapi/pkg/apis/topology/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"github.com/fromanirh/topologyinfo/cpus"
	"github.com/fromanirh/topologyinfo/numa"
	"github.com/fromanirh/topologyinfo/numa/distances"
)

const (
	defaultPodResourcesTimeout = 10 * time.Second
	// obtained these values from node e2e tests : https://github.com/kubernetes/kubernetes/blob/82baa26905c94398a0d19e1b1ecf54eb8acb6029/test/e2e_node/util.go#L70
)

type nodeResources struct {
	perNUMACapacity map[int]map[v1.ResourceName]int64
	// mapping: resourceName -> resourceID -> nodeID
	resourceID2NUMAID map[string]map[string]int
	cpuInfo           *cpus.CPUs
	numaInfo          *numa.Nodes
	numaDists         *distances.Distances
}

type resourceData struct {
	allocatable int64
	capacity    int64
}

func NewResourcesAggregator(sysfsPath string, podResourceClient podresourcesapi.PodResourcesListerClient) (ResourcesAggregator, error) {
	var err error

	cpuInfo, err := cpus.NewCPUs(sysfsPath)
	if err != nil {
		return nil, err
	}

	nodes, err := numa.NewNodesFromSysFS(sysfsPath)
	if err != nil {
		return nil, err
	}

	dists, err := distances.NewDistancesFromSysfs(sysfsPath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPodResourcesTimeout)
	defer cancel()

	//Pod Resource API client
	resp, err := podResourceClient.GetAvailableResources(ctx, &podresourcesapi.AvailableResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Can't receive response: %v.Get(_) = _, %v", podResourceClient, err)
	}

	return NewResourcesAggregatorFromData(cpuInfo, nodes, dists, resp), nil
}

func NewResourcesAggregatorFromData(cpuInfo *cpus.CPUs, nodes numa.Nodes, dists *distances.Distances, resp *podresourcesapi.AvailableResourcesResponse) ResourcesAggregator {
	allDevs := GetContainerDevicesFromAllocatableResources(resp, cpuInfo)
	return &nodeResources{
		cpuInfo:           cpuInfo,
		numaInfo:          &nodes,
		numaDists:         dists,
		resourceID2NUMAID: makeResourceMap(len(nodes.Online), allDevs),
		perNUMACapacity:   MakeNodeCapacity(allDevs),
	}
}

// Aggregate provides the mapping (numa zone name) -> Zone from the given PodResources.
func (noderesourceData *nodeResources) Aggregate(podResData []PodResources) map[string]*v1alpha1.Zone {
	perNuma := make(map[int]map[v1.ResourceName]*resourceData)
	for nodeNum, nodeRes := range noderesourceData.perNUMACapacity {
		perNuma[nodeNum] = make(map[v1.ResourceName]*resourceData)
		for resName, resCap := range nodeRes {
			perNuma[nodeNum][resName] = &resourceData{
				capacity:    resCap,
				allocatable: resCap,
			}
		}
	}

	for _, podRes := range podResData {
		for _, contRes := range podRes.Containers {
			for _, res := range contRes.Resources {
				noderesourceData.updateAllocatable(perNuma, res)
			}
		}
	}

	zones := make(map[string]*v1alpha1.Zone)

	for nodeNum, resList := range perNuma {
		zone := &v1alpha1.Zone{
			Type:      "Node",
			Resources: make(v1alpha1.ResourceInfoMap),
		}

		costs, err := makeCostsPerNumaNode(noderesourceData.numaInfo, noderesourceData.numaDists, nodeNum)
		if err != nil {
			log.Printf("cannot find costs for NUMA node %d: %v", nodeNum, err)
		} else {
			zone.Costs = costs
		}

		for name, resData := range resList {
			allocatableQty := *resource.NewQuantity(resData.allocatable, resource.DecimalSI)
			capacityQty := *resource.NewQuantity(resData.capacity, resource.DecimalSI)
			zone.Resources[name.String()] = v1alpha1.ResourceInfo{
				Allocatable: allocatableQty.String(),
				Capacity:    capacityQty.String(),
			}
		}
		zones[makeZoneName(nodeNum)] = zone
	}
	return zones
}

// GetContainerDevicesFromAllocatableResources normalize all compute resources to ContainerDevices.
// This is helpful because cpuIDs are not represented as ContainerDevices, but with a different format;
// Having a consistent representation of all the resources as ContainerDevices makes it simpler for
// the code to consume them.
func GetContainerDevicesFromAllocatableResources(availRes *podresourcesapi.AvailableResourcesResponse, cpus *cpus.CPUs) []*podresourcesapi.ContainerDevices {
	var contDevs []*podresourcesapi.ContainerDevices
	for _, dev := range availRes.GetDevices() {
		contDevs = append(contDevs, dev)
	}

	cpusPerNuma := make(map[int][]string)
	for _, cpuID := range availRes.GetCpuIds() {
		nodeID, ok := cpus.GetNodeIDForCPU(int(cpuID))
		if !ok {
			log.Printf("cannot find the NUMA node for CPU %d", cpuID)
			continue
		}

		cpuIDList := cpusPerNuma[nodeID]
		cpuIDList = append(cpuIDList, fmt.Sprintf("%d", cpuID))
		cpusPerNuma[nodeID] = cpuIDList
	}

	for nodeID, cpuList := range cpusPerNuma {
		contDevs = append(contDevs, &podresourcesapi.ContainerDevices{
			ResourceName: string(v1.ResourceCPU),
			DeviceIds:    cpuList,
			Topology: &podresourcesapi.TopologyInfo{
				Nodes: []*podresourcesapi.NUMANode{
					{ID: int64(nodeID)},
				},
			},
		})
	}

	return contDevs
}

// updateAllocatable computes the actually alloctable resources.
// This function assumes the allocatable resources are initialized to be equal to the capacity.
func (noderesourceData *nodeResources) updateAllocatable(numaData map[int]map[v1.ResourceName]*resourceData, ri ResourceInfo) {
	for _, resID := range ri.Data {
		resName := string(ri.Name)
		resMap, ok := noderesourceData.resourceID2NUMAID[resName]
		if !ok {
			log.Printf("unknown resource %q", ri.Name)
			continue
		}
		nodeNum, ok := resMap[resID]
		if !ok {
			log.Printf("unknown resource %q: %q", resName, resID)
			continue
		}
		numaData[nodeNum][ri.Name].allocatable--
	}
}

// makeZoneName returns the canonical name of a NUMA zone from its ID.
func makeZoneName(numaID int) string {
	return fmt.Sprintf("node-%d", numaID)
}

// MakeNodeCapacity computes the node capacity as mapping (NUMA node ID) -> Resource -> Capacity (amount, int).
// The computation is done assuming all the resources to represent the capacity for are represented on a slice
// of ContainerDevices. No special treatment is done for CPU IDs. See GetContainerDevicesFromAllocatableResources.
func MakeNodeCapacity(devices []*podresourcesapi.ContainerDevices) map[int]map[v1.ResourceName]int64 {
	perNUMACapacity := make(map[int]map[v1.ResourceName]int64)
	// initialize with the capacities
	for _, device := range devices {
		resourceName := device.GetResourceName()
		for _, node := range device.GetTopology().GetNodes() {
			nodeNum := int(node.GetID())
			nodeRes, ok := perNUMACapacity[nodeNum]
			if !ok {
				nodeRes = make(map[v1.ResourceName]int64)
			}
			nodeRes[v1.ResourceName(resourceName)] = int64(len(device.GetDeviceIds()))
			perNUMACapacity[nodeNum] = nodeRes
		}
	}
	return perNUMACapacity

}

// makeResourceMap creates the mapping (resource name) -> (device ID) -> (NUMA node ID) from the given slice of ContainerDevices.
// this is useful to quickly learn the NUMA ID of a given (resource, device).
func makeResourceMap(numaNodes int, devices []*podresourcesapi.ContainerDevices) map[string]map[string]int {
	resourceMap := make(map[string]map[string]int)

	for _, device := range devices {
		deviceID2NUMAID := make(map[string]int)
		for _, node := range device.GetTopology().GetNodes() {
			nodeNuma := int(node.GetID())
			for _, deviceID := range device.GetDeviceIds() {
				deviceID2NUMAID[deviceID] = nodeNuma
			}
		}
		resourceMap[device.GetResourceName()] = deviceID2NUMAID
	}
	return resourceMap
}

// makeCostsPerNumaNode builds the cost map to reach all the known NUMA zones (mapping (numa zone) -> cost) starting from the given NUMA zone.
func makeCostsPerNumaNode(numaInfo *numa.Nodes, numaDists *distances.Distances, numaIDSrc int) (map[string]int, error) {
	nodeCosts := make(map[string]int)
	for _, numaIDDst := range numaInfo.Online {
		dist, err := numaDists.BetweenNodes(numaIDSrc, numaIDDst)
		if err != nil {
			return nil, err
		}

		nodeCosts[makeZoneName(numaIDDst)] = dist
	}
	return nodeCosts, nil
}
