module sigs.k8s.io/node-feature-discovery

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.3
	github.com/jaypipes/ghw v0.6.2-0.20210115144335-efbe6fd4efca
	github.com/klauspost/cpuid v1.2.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.6.1
	github.com/swatisehgal/topologyapi v0.0.0-20201002094043-bc432ffbe41c
	github.com/vektra/errors v0.0.0-20140903201135-c64d83aba85a
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/sys v0.0.0-20201116194326-cc9327a14d48 // indirect
	google.golang.org/genproto v0.0.0-20201116205149-79184cff4dfe // indirect
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.20.0-beta.2
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.20.0-beta.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.19.4
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/yaml v1.2.0
)

// The k8s "sub-"packages do not have 'semver' compatible versions. Thus, we
// need to override with commits (corresponding their kubernetes-* tags)
replace (
	//force version of x/text due CVE-2020-14040
	golang.org/x/text => golang.org/x/text v0.3.3
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	k8s.io/api => k8s.io/api v0.20.0-beta.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0-beta.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.0-beta.2
	k8s.io/apiserver => k8s.io/apiserver v0.20.0-beta.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.0-beta.2
	k8s.io/client-go => k8s.io/client-go v0.20.0-beta.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.0-beta.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.0-beta.2
	k8s.io/code-generator => k8s.io/code-generator v0.20.0-beta.2
	k8s.io/component-base => k8s.io/component-base v0.20.0-beta.2
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.0-beta.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.0-beta.2
	k8s.io/cri-api => k8s.io/cri-api v0.20.0-beta.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.0-beta.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.0-beta.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.0-beta.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.0-beta.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.0-beta.2
	k8s.io/kubectl => k8s.io/kubectl v0.20.0-beta.2
	k8s.io/kubelet => github.com/fromanirh/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20201127133213-6f0bc0d851ab
	k8s.io/kubernetes => github.com/fromanirh/kubernetes v1.18.0-alpha.1.0.20201127133213-6f0bc0d851ab
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.0-beta.2
	k8s.io/metrics => k8s.io/metrics v0.20.0-beta.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.0-beta.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.0-beta.2
)
