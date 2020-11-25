/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package distances

import (
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/node-feature-discovery/pkg/topologyinfo/numa"
	"sigs.k8s.io/node-feature-discovery/pkg/topologyinfo/sysfs"
)

const (
	distanceSeparator string = " "
)

func nodeDistancesFromString(numaNodes int, data string) ([]int, error) {
	dists := strings.Split(strings.TrimSpace(data), distanceSeparator)
	if len(dists) != numaNodes {
		return nil, fmt.Errorf("found %d distance values, expected %d", len(dists), numaNodes)
	}
	ret := make([]int, numaNodes, numaNodes)
	for idx, dist := range dists {
		val, err := strconv.Atoi(dist)
		if err != nil {
			return ret, err
		}
		ret[idx] = val
	}
	return ret, nil
}

type Distances struct {
	onlineNodes map[int]bool
	// sysfs provides a dense square distance cost matrix
	byNode [][]int
}

func (d *Distances) BetweenNodes(from, to int) (int, error) {
	if _, ok := d.onlineNodes[from]; !ok {
		return -1, fmt.Errorf("unknown NUMA node: %d", from)
	}
	if _, ok := d.onlineNodes[to]; !ok {
		return -1, fmt.Errorf("unknown NUMA node: %d", to)
	}
	return d.byNode[from][to], nil
}

// NewDistancesFromData takes a map in the format "0": "10 21 30\n"
func NewDistancesFromData(data map[string]string) (*Distances, error) {
	dist := NewDistancesEmpty()

	numNodes := len(data)
	for nodeData, distData := range data {
		nodeID, err := strconv.Atoi(nodeData)
		if err != nil {
			return dist, err
		}
		dist.onlineNodes[nodeID] = true
		nodeDist, err := nodeDistancesFromString(numNodes, distData)
		if err != nil {
			return dist, err
		}
		dist.byNode = append(dist.byNode, nodeDist)
	}
	return dist, nil
}

func NewDistancesFromSysfs(sysfsPath string) (*Distances, error) {
	nodes, err := numa.NewNodesFromSysFS(sysfsPath)
	if err != nil {
		return nil, err
	}

	dist := NewDistancesEmpty()

	sys := sysfs.New(sysfsPath)
	for _, nodeID := range nodes.Online {
		// here we are iterating over src nodes: reading the distances to other nodes
		// from each of the online nodes. But if we know how to reach all other nodes
		// from node X, then we know node X is online, so is safe to assume node X will
		// be included in the distances vectors of other nodes.
		// TL;DR: no need to explicitely iterate over destination nodes.
		dist.onlineNodes[nodeID] = true

		distData, err := sys.ForNode(nodeID).ReadFile("distance")
		if err != nil {
			return nil, err
		}

		nodeDist, err := nodeDistancesFromString(len(nodes.Online), distData)
		if err != nil {
			return nil, err
		}

		dist.byNode = append(dist.byNode, nodeDist)
	}

	return dist, nil
}

func NewDistancesEmpty() *Distances {
	dist := Distances{
		onlineNodes: make(map[int]bool),
	}
	return &dist
}
