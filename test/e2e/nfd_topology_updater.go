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

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/test/e2e/framework"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	e2enetwork "k8s.io/kubernetes/test/e2e/framework/network"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"

	testutils "sigs.k8s.io/node-feature-discovery/test/e2e/utils"
)

var _ = framework.KubeDescribe("[NFD] Node topology updater", func() {
	f := framework.NewDefaultFramework("node-topology-updater")

	ginkgo.Context("with single nfd-master pod", func() {
		var (
			extClient                *extclient.Clientset
			crd                      *apiextensionsv1.CustomResourceDefinition
			masterPod                *v1.Pod
			masterService            *v1.Service
			topologyUpdaterDaemonSet *appsv1.DaemonSet
		)

		ginkgo.BeforeEach(func() {
			var err error

			if extClient == nil {
				extClient, err = extclient.NewForConfig(f.ClientConfig())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}

			ginkgo.By("Creating the node resource topologies CRD")
			crd = testutils.NewNodeResourceTopologies()
			_, err = extClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = testutils.ConfigureRBAC(f.ClientSet, f.Namespace.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			image := fmt.Sprintf("%s:%s", *dockerRepo, *dockerTag)
			masterPod = f.PodClient().CreateSync(testutils.NFDMasterPod(image, false))

			// Create nfd-master service
			masterService, err = testutils.CreateService(f.ClientSet, f.Namespace.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Waiting for the nfd-master service to be up")
			gomega.Expect(e2enetwork.WaitForService(f.ClientSet, f.Namespace.Name, masterService.Name, true, time.Second, 10*time.Second)).NotTo(gomega.HaveOccurred())

			ginkgo.By("Creating nfd-topology-updater daemonset")
			topologyUpdaterDaemonSet = testutils.NFDTopologyUpdaterDaemonSet(fmt.Sprintf("%s:%s", *dockerRepo, *dockerTag), []string{})
			topologyUpdaterDaemonSet, err = f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Create(context.TODO(), topologyUpdaterDaemonSet, metav1.CreateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Waiting for daemonset pods to be ready")
			gomega.Expect(e2epod.WaitForPodsReady(f.ClientSet, f.Namespace.Name, topologyUpdaterDaemonSet.Spec.Template.Labels["name"], 5)).NotTo(gomega.HaveOccurred())
		})

		ginkgo.FIt("should fill the node resource topologies CR with the data", func() {

		})

		ginkgo.AfterEach(func() {
			// Delete topology updater daemon set
			err := f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Delete(context.TODO(), topologyUpdaterDaemonSet.Name, metav1.DeleteOptions{})
			if err != nil {
				e2elog.Logf("failed to delete node topology updater daemon set: %v", err)
			}

			err = f.ClientSet.CoreV1().Services(f.Namespace.Name).Delete(context.TODO(), masterService.Name, metav1.DeleteOptions{})
			if err != nil {
				e2elog.Logf("failed to delete master service: %v", err)
			}

			err = f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(context.TODO(), masterPod.Name, metav1.DeleteOptions{})
			if err != nil {
				e2elog.Logf("failed to delete master pod: %v", err)
			}

			err = testutils.DeconfigureRBAC(f.ClientSet, f.Namespace.Name)
			if err != nil {
				e2elog.Logf("failed to delete RBAC resources: %v", err)
			}

			err = extClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crd.Name, metav1.DeleteOptions{})
			if err != nil {
				e2elog.Logf("failed to delete node resources topologies CRD: %v", err)
			}
		})
	})
})
