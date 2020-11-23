/*
Copyright 2018 The Kubernetes Authors.

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
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	e2enetwork "k8s.io/kubernetes/test/e2e/framework/network"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"

	master "sigs.k8s.io/node-feature-discovery/pkg/nfd-master"
	testutils "sigs.k8s.io/node-feature-discovery/test/e2e/utils"
)

var (
	dockerRepo = flag.String("nfd.repo", "quay.io/kubernetes_incubator/node-feature-discovery", "Docker repository to fetch image from")
	dockerTag  = flag.String("nfd.tag", "e2e-test", "Docker tag to use")
)

// cleanupNode deletes all NFD-related metadata from the Node object, i.e.
// labels and annotations
func cleanupNode(cs clientset.Interface) {
	nodeList, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	for _, n := range nodeList.Items {
		var err error
		var node *v1.Node
		for retry := 0; retry < 5; retry++ {
			node, err = cs.CoreV1().Nodes().Get(context.TODO(), n.Name, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			update := false
			// Remove labels
			for key := range node.Labels {
				if strings.HasPrefix(key, master.LabelNs) {
					delete(node.Labels, key)
					update = true
				}
			}

			// Remove annotations
			for key := range node.Annotations {
				if strings.HasPrefix(key, master.AnnotationNs) {
					delete(node.Annotations, key)
					update = true
				}
			}

			if !update {
				break
			}

			ginkgo.By("Deleting NFD labels and annotations from node " + node.Name)
			_, err = cs.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
			if err != nil {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}

		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

// Actual test suite
var _ = framework.KubeDescribe("[NFD] Node Feature Discovery", func() {
	f := framework.NewDefaultFramework("node-feature-discovery")

	ginkgo.Context("when deploying a single nfd-master pod", func() {
		var masterPod *v1.Pod

		ginkgo.BeforeEach(func() {
			err := testutils.ConfigureRBAC(f.ClientSet, f.Namespace.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			image := fmt.Sprintf("%s:%s", *dockerRepo, *dockerTag)
			masterPod = f.PodClient().CreateSync(testutils.NFDMasterPod(image, false))

			// Create nfd-master service
			nfdSvc, err := testutils.CreateService(f.ClientSet, f.Namespace.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Waiting for the nfd-master service to be up")
			gomega.Expect(e2enetwork.WaitForService(f.ClientSet, f.Namespace.Name, nfdSvc.Name, true, time.Second, 10*time.Second)).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			err := testutils.DeconfigureRBAC(f.ClientSet, f.Namespace.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

		})

		//
		// Simple test with only the fake source enabled
		//
		ginkgo.Context("and a single worker pod with fake source enabled", func() {
			ginkgo.It("it should decorate the node with the fake feature labels", func() {

				fakeFeatureLabels := map[string]string{
					master.LabelNs + "fake-fakefeature1": "true",
					master.LabelNs + "fake-fakefeature2": "true",
					master.LabelNs + "fake-fakefeature3": "true",
				}

				// Remove pre-existing stale annotations and labels
				cleanupNode(f.ClientSet)

				// Launch nfd-worker
				ginkgo.By("Creating a nfd worker pod")
				image := fmt.Sprintf("%s:%s", *dockerRepo, *dockerTag)
				workerPod := testutils.NFDWorkerPod(image, []string{"--oneshot", "--sources=fake"})
				workerPod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), workerPod, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				ginkgo.By("Waiting for the nfd-worker pod to succeed")
				gomega.Expect(e2epod.WaitForPodSuccessInNamespace(f.ClientSet, workerPod.ObjectMeta.Name, f.Namespace.Name)).NotTo(gomega.HaveOccurred())
				workerPod, err = f.ClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), workerPod.ObjectMeta.Name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				ginkgo.By(fmt.Sprintf("Making sure '%s' was decorated with the fake feature labels", workerPod.Spec.NodeName))
				node, err := f.ClientSet.CoreV1().Nodes().Get(context.TODO(), workerPod.Spec.NodeName, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				for k, v := range fakeFeatureLabels {
					gomega.Expect(node.Labels[k]).To(gomega.Equal(v))
				}

				// Check that there are no unexpected NFD labels
				for k := range node.Labels {
					if strings.HasPrefix(k, master.LabelNs) {
						gomega.Expect(fakeFeatureLabels).Should(gomega.HaveKey(k))
					}
				}

				ginkgo.By("Deleting the node-feature-discovery worker pod")
				err = f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(context.TODO(), workerPod.ObjectMeta.Name, metav1.DeleteOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				cleanupNode(f.ClientSet)
			})
		})

		//
		// More comprehensive test when --e2e-node-config is enabled
		//
		ginkgo.Context("and nfd-workers as a daemonset with default sources enabled", func() {
			ginkgo.It("the node labels and annotations listed in the e2e config should be present", func() {
				err := testutils.ReadConfig()
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				if testutils.E2EConfigFile == nil {
					ginkgo.Skip("no e2e-config was specified")
				}
				if testutils.E2EConfigFile.DefaultFeatures == nil {
					ginkgo.Skip("no 'defaultFeatures' specified in e2e-config")
				}
				fConf := testutils.E2EConfigFile.DefaultFeatures

				// Remove pre-existing stale annotations and labels
				cleanupNode(f.ClientSet)

				ginkgo.By("Creating nfd-worker daemonset")
				workerDS := testutils.NFDWorkerDaemonSet(fmt.Sprintf("%s:%s", *dockerRepo, *dockerTag), []string{})
				workerDS, err = f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Create(context.TODO(), workerDS, metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				ginkgo.By("Waiting for daemonset pods to be ready")
				gomega.Expect(e2epod.WaitForPodsReady(f.ClientSet, f.Namespace.Name, workerDS.Spec.Template.Labels["name"], 5)).NotTo(gomega.HaveOccurred())

				ginkgo.By("Getting node objects")
				nodeList, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				for _, node := range nodeList.Items {
					if _, ok := fConf.Nodes[node.Name]; !ok {
						e2elog.Logf("node %q missing from e2e-config, skipping...", node.Name)
						continue
					}
					nodeConf := fConf.Nodes[node.Name]

					// Check labels
					e2elog.Logf("verifying labels of node %q...", node.Name)
					for k, v := range nodeConf.ExpectedLabelValues {
						gomega.Expect(node.Labels).To(gomega.HaveKeyWithValue(k, v))
					}
					for k := range nodeConf.ExpectedLabelKeys {
						gomega.Expect(node.Labels).To(gomega.HaveKey(k))
					}
					for k := range node.Labels {
						if strings.HasPrefix(k, master.LabelNs) {
							if _, ok := nodeConf.ExpectedLabelValues[k]; ok {
								continue
							}
							if _, ok := nodeConf.ExpectedLabelKeys[k]; ok {
								continue
							}
							// Ignore if the label key was not whitelisted
							gomega.Expect(fConf.LabelWhitelist).NotTo(gomega.HaveKey(k))
						}
					}

					// Check annotations
					e2elog.Logf("verifying annotations of node %q...", node.Name)
					for k, v := range nodeConf.ExpectedAnnotationValues {
						gomega.Expect(node.Annotations).To(gomega.HaveKeyWithValue(k, v))
					}
					for k := range nodeConf.ExpectedAnnotationKeys {
						gomega.Expect(node.Annotations).To(gomega.HaveKey(k))
					}
					for k := range node.Annotations {
						if strings.HasPrefix(k, master.AnnotationNs) {
							if _, ok := nodeConf.ExpectedAnnotationValues[k]; ok {
								continue
							}
							if _, ok := nodeConf.ExpectedAnnotationKeys[k]; ok {
								continue
							}
							// Ignore if the annotation was not whitelisted
							gomega.Expect(fConf.AnnotationWhitelist).NotTo(gomega.HaveKey(k))
						}
					}

					// Node running nfd-master should have master version annotation
					if node.Name == masterPod.Spec.NodeName {
						gomega.Expect(node.Annotations).To(gomega.HaveKey(master.AnnotationNs + "master.version"))
					}
				}

				ginkgo.By("Deleting nfd-worker daemonset")
				err = f.ClientSet.AppsV1().DaemonSets(f.Namespace.Name).Delete(context.TODO(), workerDS.ObjectMeta.Name, metav1.DeleteOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				cleanupNode(f.ClientSet)
			})
		})
	})
})
