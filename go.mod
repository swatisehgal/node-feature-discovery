module sigs.k8s.io/node-feature-discovery

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/googleapis/gnostic v0.4.1
	github.com/klauspost/cpuid v1.2.3
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.5.1
	github.com/swatisehgal/topologyapi v0.0.0-20201002094043-bc432ffbe41c
	github.com/vektra/errors v0.0.0-20140903201135-c64d83aba85a
	golang.org/x/net v0.0.0-20201109172640-a11eb1b685be
	golang.org/x/sys v0.0.0-20201109165425-215b40eba54c // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201109203340-2640f1f9cdfb // indirect
	google.golang.org/grpc v1.33.2
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.1 // indirect
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.18.6
	k8s.io/klog v1.0.0
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.18.6
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
	sigs.k8s.io/yaml v1.2.0
)

// The k8s "sub-"packages do not have 'semver' compatible versions. Thus, we
// need to override with commits (corresponding their kubernetes-* tags)
replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/swatisehgal/topologyapi => github.com/swatisehgal/topologyapi v0.0.0-20201002094043-bc432ffbe41c
	//force version of x/text due CVE-2020-14040
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/apiserver => k8s.io/apiserver v0.18.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.6
	k8s.io/code-generator => k8s.io/code-generator v0.18.6
	k8s.io/component-base => k8s.io/component-base v0.18.6
	k8s.io/cri-api => k8s.io/cri-api v0.18.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.6
	k8s.io/kubectl => k8s.io/kubectl v0.18.6
	k8s.io/kubelet => k8s.io/kubelet v0.18.6
	k8s.io/kubernetes => github.com/swatisehgal/kubernetes v1.19.0-beta.2.0.20200928123437-1c4f5ad40262
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.6
	k8s.io/metrics => k8s.io/metrics v0.18.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.6
)
