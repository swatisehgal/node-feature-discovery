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
	"log"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"

	topov1alpha1 "github.com/swatisehgal/topologyapi/pkg/apis/topology/v1alpha1"
	v1 "k8s.io/kubelet/pkg/apis/podresources/v1"

	"github.com/fromanirh/topologyinfo/cpus"
	"github.com/fromanirh/topologyinfo/numa"
	"github.com/fromanirh/topologyinfo/numa/distances"
)

func TestResourcesAggregator(t *testing.T) {

	var resAggr ResourcesAggregator

	Convey("When I aggregate the node resources fake data and no pod allocation", t, func() {
		cpuInfo := &cpus.CPUs{
			Present: cpus.CPUIdList{0, 1},
			Online:  cpus.CPUIdList{0, 1},
			CoreCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1},
				1: cpus.CPUIdList{0, 1},
				2: cpus.CPUIdList{2, 3},
				3: cpus.CPUIdList{2, 3},
				4: cpus.CPUIdList{4, 5},
				5: cpus.CPUIdList{4, 5},
				6: cpus.CPUIdList{6, 7},
				7: cpus.CPUIdList{6, 7},
			},
			PackageCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1, 2, 3},
				1: cpus.CPUIdList{4, 5, 6, 7},
			},
			Packages:  cpus.CPUIdList{0, 1},
			NUMANodes: cpus.CPUIdList{0, 1},
			NUMANodeCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1, 2, 3},
				1: cpus.CPUIdList{4, 5, 6, 7},
			},
		}

		nodes := numa.Nodes{
			Online:           []int{0, 1},
			Possible:         []int{0, 1},
			WithCPU:          []int{0, 1},
			WithMemory:       []int{0, 1},
			WithNormalMemory: []int{0, 1},
		}

		dists, err := distances.NewDistancesFromData(map[string]string{
			"0": "10 30\n",
			"1": "30 10\n",
		})
		So(err, ShouldBeNil)

		availRes := &v1.AllocatableResourcesResponse{
			Devices: []*v1.ContainerDevices{
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netAAA-0"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 0,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netAAA-1"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 0,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netAAA-2"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 0,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netAAA-3"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 0,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netBBB-0"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 1,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netBBB-1"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 1,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/gpu",
					DeviceIds:    []string{"gpuAAA"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 1,
							},
						},
					},
				},
			},
			CpuIds: []int64{0, 1, 2, 3, 4, 5, 6, 7},
		}

		resAggr = NewResourcesAggregatorFromData(cpuInfo.Update(), nodes, dists, availRes)

		Convey("Creating a Resources Aggregator using a mock client", func() {
			So(err, ShouldBeNil)
		})

		Convey("When aggregating resources", func() {
			expected := map[string]*topov1alpha1.Zone{
				"node-0": &topov1alpha1.Zone{
					Type: "Node",
					Costs: map[string]int{
						"node-0": 10,
						"node-1": 30,
					},
					Resources: topov1alpha1.ResourceInfoMap{
						"fake.io/net": topov1alpha1.ResourceInfo{
							Allocatable: "4",
							Capacity:    "4",
						},
						"cpu": topov1alpha1.ResourceInfo{
							Allocatable: "4",
							Capacity:    "4",
						},
					},
				},
				"node-1": &topov1alpha1.Zone{
					Type: "Node",
					Costs: map[string]int{
						"node-0": 30,
						"node-1": 10,
					},
					Resources: topov1alpha1.ResourceInfoMap{
						"fake.io/gpu": topov1alpha1.ResourceInfo{
							Allocatable: "1",
							Capacity:    "1",
						},
						"fake.io/net": topov1alpha1.ResourceInfo{
							Allocatable: "2",
							Capacity:    "2",
						},
						"cpu": topov1alpha1.ResourceInfo{
							Allocatable: "4",
							Capacity:    "4",
						},
					},
				},
			}

			res := resAggr.Aggregate(nil) // XXX
			log.Printf("diff=%s", cmp.Diff(res, expected))
			So(cmp.Equal(res, expected), ShouldBeTrue)
		})
	})

	Convey("When I aggregate the node resources fake data and some pod allocation", t, func() {
		cpuInfo := &cpus.CPUs{
			Present: cpus.CPUIdList{0, 1},
			Online:  cpus.CPUIdList{0, 1},
			CoreCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1},
				1: cpus.CPUIdList{0, 1},
				2: cpus.CPUIdList{2, 3},
				3: cpus.CPUIdList{2, 3},
				4: cpus.CPUIdList{4, 5},
				5: cpus.CPUIdList{4, 5},
				6: cpus.CPUIdList{6, 7},
				7: cpus.CPUIdList{6, 7},
			},
			PackageCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1, 2, 3},
				1: cpus.CPUIdList{4, 5, 6, 7},
			},
			Packages:  cpus.CPUIdList{0, 1},
			NUMANodes: cpus.CPUIdList{0, 1},
			NUMANodeCPUs: map[int]cpus.CPUIdList{
				0: cpus.CPUIdList{0, 1, 2, 3},
				1: cpus.CPUIdList{4, 5, 6, 7},
			},
		}

		nodes := numa.Nodes{
			Online:           []int{0, 1},
			Possible:         []int{0, 1},
			WithCPU:          []int{0, 1},
			WithMemory:       []int{0, 1},
			WithNormalMemory: []int{0, 1},
		}

		dists, err := distances.NewDistancesFromData(map[string]string{
			"0": "10 30\n",
			"1": "30 10\n",
		})
		So(err, ShouldBeNil)

		availRes := &v1.AllocatableResourcesResponse{
			Devices: []*v1.ContainerDevices{
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netAAA"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 0,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/net",
					DeviceIds:    []string{"netBBB"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 1,
							},
						},
					},
				},
				&v1.ContainerDevices{
					ResourceName: "fake.io/gpu",
					DeviceIds:    []string{"gpuAAA"},
					Topology: &v1.TopologyInfo{
						Nodes: []*v1.NUMANode{
							&v1.NUMANode{
								ID: 1,
							},
						},
					},
				},
			},
			CpuIds: []int64{0, 1, 2, 3, 4, 5, 6, 7},
		}

		resAggr = NewResourcesAggregatorFromData(cpuInfo.Update(), nodes, dists, availRes)

		Convey("Creating a Resources Aggregator using a mock client", func() {
			So(err, ShouldBeNil)
		})
		Convey("When aggregating resources", func() {
			podRes := []PodResources{
				PodResources{
					Name:      "test-pod-0",
					Namespace: "default",
					Containers: []ContainerResources{
						ContainerResources{
							Name: "test-cnt-0",
							Resources: []ResourceInfo{
								ResourceInfo{
									Name: "cpu",
									Data: []string{"6", "7"},
								},
								ResourceInfo{
									Name: "fake.io/net",
									Data: []string{"netBBB"},
								},
							},
						},
					},
				},
			}

			expected := map[string]*topov1alpha1.Zone{
				"node-0": &topov1alpha1.Zone{
					Type: "Node",
					Costs: map[string]int{
						"node-0": 10,
						"node-1": 30,
					},
					Resources: topov1alpha1.ResourceInfoMap{
						"fake.io/net": topov1alpha1.ResourceInfo{
							Allocatable: "1",
							Capacity:    "1",
						},
						"cpu": topov1alpha1.ResourceInfo{
							Allocatable: "4",
							Capacity:    "4",
						},
					},
				},
				"node-1": &topov1alpha1.Zone{
					Type: "Node",
					Costs: map[string]int{
						"node-0": 30,
						"node-1": 10,
					},
					Resources: topov1alpha1.ResourceInfoMap{
						"fake.io/gpu": topov1alpha1.ResourceInfo{
							Allocatable: "1",
							Capacity:    "1",
						},
						"fake.io/net": topov1alpha1.ResourceInfo{
							Allocatable: "0",
							Capacity:    "1",
						},
						"cpu": topov1alpha1.ResourceInfo{
							Allocatable: "2",
							Capacity:    "4",
						},
					},
				},
			}

			res := resAggr.Aggregate(podRes)
			log.Printf("diff=%s", cmp.Diff(res, expected))
			So(cmp.Equal(res, expected), ShouldBeTrue)
		})
	})

}
