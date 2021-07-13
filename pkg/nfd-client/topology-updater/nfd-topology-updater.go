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

package topologyupdater

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	v1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	"golang.org/x/net/context"
	nfdclient "sigs.k8s.io/node-feature-discovery/pkg/nfd-client"
	"sigs.k8s.io/node-feature-discovery/pkg/podres"
	"sigs.k8s.io/node-feature-discovery/pkg/resourcemonitor"
	pb "sigs.k8s.io/node-feature-discovery/pkg/topologyupdater"
	"sigs.k8s.io/node-feature-discovery/pkg/utils"
	"sigs.k8s.io/node-feature-discovery/pkg/version"
)

// Command line arguments
type Args struct {
	nfdclient.Args
	NoPublish bool
	Oneshot   bool
}

type NfdTopologyUpdater interface {
	nfdclient.NfdClient
	Update(v1alpha1.ZoneList) error
}

type staticNodeInfo struct {
	args                Args
	resourcemonitorArgs resourcemonitor.Args
	tmPolicy            string
}

type nfdTopologyUpdater struct {
	nfdclient.NfdBaseClient
	nodeInfo  *staticNodeInfo
	certWatch *utils.FsWatcher
	client    pb.NodeTopologyClient
	stop      chan struct{} // channel for signaling stop
}

// Create new NewTopologyUpdater instance.
func NewTopologyUpdater(args *Args, resourcemonitorArgs *resourcemonitor.Args, policy string) (NfdTopologyUpdater, error) {
	base, err := nfdclient.NewNfdBaseClient(&args.Args)
	if err != nil {
		return nil, err
	}

	nfd := &nfdTopologyUpdater{
		NfdBaseClient: base,
		nodeInfo: &staticNodeInfo{
			args:                *args,
			resourcemonitorArgs: *resourcemonitorArgs,
			tmPolicy:            policy,
		},
		stop: make(chan struct{}, 1),
	}
	return nfd, nil
}

// Run nfdTopologyUpdater client. Returns if a fatal error is encountered, or, after
// one request if OneShot is set to 'true' in the updater args.
func (w *nfdTopologyUpdater) Run() error {
	klog.Infof("Node Feature Discovery Topology Updater %s", version.Get())
	klog.Infof("NodeName: '%s'", nfdclient.NodeName)

	podResClient, err := podres.GetPodResClient(w.nodeInfo.resourcemonitorArgs.PodResourceSocketPath)
	if err != nil {
		klog.Fatalf("Failed to get PodResource Client: %v", err)
	}
	var resScan resourcemonitor.ResourcesScanner

	resScan, err = resourcemonitor.NewPodResourcesScanner(w.nodeInfo.resourcemonitorArgs.Namespace, podResClient)
	if err != nil {
		klog.Fatalf("Failed to initialize ResourceMonitor instance: %v", err)
	}

	// CAUTION: these resources are expected to change rarely - if ever.
	// So we are intentionally do this once during the process lifecycle.
	// TODO: Obtain node resources dynamically from the podresource API
	// zonesChannel := make(chan v1alpha1.ZoneList)
	var zones v1alpha1.ZoneList

	resAggr, err := resourcemonitor.NewResourcesAggregator(w.nodeInfo.resourcemonitorArgs.FsRoot, podResClient)
	if err != nil {
		klog.Fatalf("Failed to obtain node resource information: %v", err)
	}

	klog.Infof("resAggr is: %v\n", resAggr)

	// Create watcher for TLS certificates
	w.certWatch, err = utils.CreateFsWatcher(time.Second, w.nodeInfo.args.CaFile, w.nodeInfo.args.CertFile, w.nodeInfo.args.KeyFile)
	if err != nil {
		return err
	}

	crTrigger := time.After(0)
	for {
		select {
		case <-crTrigger:
			klog.Infof("Scanning\n")
			podResources, err := resScan.Scan()
			utils.KlogDump(1, "podResources are", "  ", podResources)
			if err != nil {
				klog.Warningf("Scan failed: %v\n", err)
				continue
			}
			zones = resAggr.Aggregate(podResources)
			utils.KlogDump(1, "After aggregating resources identified zones are", "  ", zones)
			if err = w.Update(zones); err != nil {
				klog.Exit(err)
			}

			if w.nodeInfo.args.Oneshot {
				return nil
			}

			if w.nodeInfo.resourcemonitorArgs.SleepInterval > 0 {
				crTrigger = time.After(w.nodeInfo.resourcemonitorArgs.SleepInterval)
			}

		case <-w.certWatch.Events:
			klog.Infof("TLS certificate update, renewing connection to nfd-master")
			w.Disconnect()
			if err := w.Connect(); err != nil {
				return err
			}

		case <-w.stop:
			klog.Infof("shutting down nfd-worker")
			w.certWatch.Close()
			return nil
		}
	}

}

func (w *nfdTopologyUpdater) Update(zones v1alpha1.ZoneList) error {
	// Connect to NFD master
	err := w.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer w.Disconnect()
	// Update the node with the feature labels.
	if w.client == nil {
		return nil
	}

	err = advertiseNodeTopology(w.client, zones, w.nodeInfo.tmPolicy, nfdclient.NodeName())
	if err != nil {
		return fmt.Errorf("failed to advertise node topology: %s", err.Error())
	}

	return nil
}

// Stop NfdWorker
func (w *nfdTopologyUpdater) Stop() {
	select {
	case w.stop <- struct{}{}:
	default:
	}
}

// connect creates a client connection to the NFD master
func (w *nfdTopologyUpdater) Connect() error {
	// Return a dummy connection in case of dry-run
	if w.nodeInfo.args.NoPublish {
		return nil
	}

	if err := w.NfdBaseClient.Connect(); err != nil {
		return err
	}
	w.client = pb.NewNodeTopologyClient(w.ClientConn())

	return nil
}

// disconnect closes the connection to NFD master
func (w *nfdTopologyUpdater) Disconnect() {
	w.NfdBaseClient.Disconnect()
}

// advertiseNodeTopology advertises the topology CR to a Kubernetes node
// via the NFD server.
func advertiseNodeTopology(client pb.NodeTopologyClient, zoneInfo v1alpha1.ZoneList, tmPolicy string, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	zones := make([]*pb.Zone, len(zoneInfo))
	for i, zone := range zoneInfo {
		resInfo := make([]*pb.ResourceInfo, len(zone.Resources))
		for j, info := range zone.Resources {
			resInfo[j] = &pb.ResourceInfo{
				Name:        info.Name,
				Allocatable: info.Allocatable.String(),
				Capacity:    info.Capacity.String(),
			}
		}

		zones[i] = &pb.Zone{
			Name:      zone.Name,
			Type:      zone.Type,
			Parent:    zone.Parent,
			Resources: resInfo,
			Costs:     updateMap(zone.Costs),
		}
	}

	topologyReq := &pb.NodeTopologyRequest{
		Zones:            zones,
		NfdVersion:       version.Get(),
		NodeName:         nodeName,
		TopologyPolicies: []string{tmPolicy},
	}

	utils.KlogDump(1, "Sending NodeTopologyRequest to nfd-master:", "  ", topologyReq)

	_, err := client.UpdateNodeTopology(ctx, topologyReq)
	if err != nil {
		klog.Warningf("failed to set node topology CR: %v", err)
		return err
	}

	return nil
}
func updateMap(data []v1alpha1.CostInfo) []*pb.CostInfo {
	ret := make([]*pb.CostInfo, len(data))
	for i, cost := range data {
		ret[i] = &pb.CostInfo{
			Name:  cost.Name,
			Value: int32(cost.Value),
		}
	}
	return ret
}
