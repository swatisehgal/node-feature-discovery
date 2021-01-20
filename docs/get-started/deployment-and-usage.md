---
title: "Deployment and Usage"
layout: default
sort: 3
---

# Deployment and Usage
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Requirements

1. Linux (x86_64/Arm64/Arm)
1. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl)
   (properly set up and configured to work with your Kubernetes cluster)

## Deployment options

### Operator

Deployment using the
[Node Feature Discovery Operator][nfd-operator]
is recommended to be done via
[operatorhub.io](https://operatorhub.io/operator/nfd-operator).

1. You need to have
   [OLM][OLM]
   installed. If you don't, take a look at the
   [latest release](https://github.com/operator-framework/operator-lifecycle-manager/releases/latest)
   for detailed instructions.
1. Install the operator:
```bash
kubectl create -f https://operatorhub.io/install/nfd-operator.yaml
```
1. Create NodeFeatureDiscovery resource (in `nfd` namespace here):
```bash
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: nfd
---
apiVersion: nfd.kubernetes.io/v1alpha1
kind: NodeFeatureDiscovery
metadata:
  name: my-nfd-deployment
  namespace: nfd
EOF
```

### Deployment Templates

The template specs provided in the repo can be used directly:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-master.yaml.template
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-worker-daemonset.yaml.template
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-topology-updater-daemonset.yaml.template
```

This will required RBAC rules and deploy nfd-master (as a deployment) and
nfd-worker and nfd-topology-updater (as daemonset) in the `node-feature-discovery` namespace.

Alternatively you can download the templates and customize the deployment
manually.

#### Master-Worker Pod

You can also run nfd-master, nfd-worker and nfd-topology-updater inside the same pod

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-daemonset-combined.yaml.template
```

This creates a DaemonSet runs nfd-worker, nfd-master and nfd-topology-updater in the same Pod.
In this case no nfd-master is run on the master node(s), but, the worker nodes
are able to label themselves which may be desirable e.g. in single-node setups.
Also, NodeResourceTopology CRD instances are created corresponding to the worker nodes.

#### Worker One-shot

Feature discovery can alternatively be configured as a one-shot job.
The Job template may be used to achieve this:

```bash
NUM_NODES=$(kubectl get no -o jsonpath='{.items[*].metadata.name}' | wc -w)
curl -fs https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-worker-job.yaml.template | \
    sed s"/NUM_NODES/$NUM_NODES/" | \
    kubectl apply -f -
```

The example above launches as many jobs as there are non-master nodes. Note that
this approach does not guarantee running once on every node. For example,
tainted, non-ready nodes or some other reasons in Job scheduling may cause some
node(s) will run extra job instance(s) to satisfy the request.

#### Topology Updater One-shot

NFD Topology Updater can alternatively be configured as a one-shot job.
The Job template may be used to achieve this:

```bash
NUM_NODES=$(kubectl get no -o jsonpath='{.items[*].metadata.name}' | wc -w)
curl -fs https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-topology-updater-job.yaml.template | \
    sed s"/NUM_NODES/$NUM_NODES/" | \
    kubectl apply -f -
```

The example above launches as many jobs as there are non-master nodes. Note that
this approach does not guarantee running once on every node. For example,
tainted, non-ready nodes or some other reasons in Job scheduling may cause some
node(s) will run extra job instance(s) to satisfy the request.

### Build Your Own

If you want to use the latest development version (master branch) you need to
build your own custom image.
See the [Developer Guide](../advanced/developer-guide) for instructions how to
build images and deploy them on your cluster.

## Usage

### NFD-Master

NFD-Master runs as a deployment (with a replica count of 1), by default
it prefers running on the cluster's master nodes but will run on worker
nodes if no master nodes are found.

For High Availability, you should simply increase the replica count of
the deployment object. You should also look into adding
[inter-pod](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)
affinity to prevent masters from running on the same node.
However note that inter-pod affinity is costly and is not recommended
in bigger clusters.

NFD-Master listens for connections from nfd-worker(s) and connects to the
Kubernetes API server to add node labels advertised by them.

If you have RBAC authorization enabled (as is the default e.g. with clusters
initialized with kubeadm) you need to configure the appropriate ClusterRoles,
ClusterRoleBindings and a ServiceAccount in order for NFD to create node
labels. The provided template will configure these for you.

### NFD-Worker

NFD-Worker is preferably run as a Kubernetes DaemonSet. This assures
re-labeling on regular intervals capturing changes in the system configuration
and makes sure that new nodes are labeled as they are added to the cluster.
Worker connects to the nfd-master service to advertise hardware features.

When run as a daemonset, nodes are re-labeled at an interval specified using
the `--sleep-interval` option. In the
[template](https://github.com/kubernetes-sigs/node-feature-discovery/blob/{{ site.release }}/nfd-worker-daemonset.yaml.template#L26)
the default interval is set to 60s which is also the default when no
`--sleep-interval` is specified. Also, the configuration file is re-read on
each iteration providing a simple mechanism of run-time reconfiguration.

### NFD-Topology-Updater

NFD-Worker is preferably run as a Kubernetes DaemonSet. This assures
re-examination (and CRD updation) on regular intervals capturing changes in the
allocated resources and hence the allocatable resources on a per zone basis. It
makes sure that more CRD instances are created as new nodes get added to the
cluster. Topology-Updater connects to the nfd-master service to creates CRD
intances corresponding to nodes.

When run as a daemonset, nodes are re-examined for the allocated resources
(to determine the information of the allocatable resources on a per zone basis
where a zone can be a NUMA node) at an interval specified using the
`--sleep-interval` option. In the
[template](https://github.com/kubernetes-sigs/node-feature-discovery/blob/{{ site.release }}/nfd-topology-daemonset.yaml.template#L37) the interval is set to 3s.


### TLS authentication

NFD supports mutual TLS authentication between the nfd-master and nfd-worker
instances.  That is, nfd-worker and nfd-master both verify that the other end
presents a valid certificate.

TLS authentication is enabled by specifying `--ca-file`, `--key-file` and
`--cert-file` args, on both the nfd-master and nfd-worker instances.
The template specs provided with NFD contain (commented out) example
configuration for enabling TLS authentication.

The Common Name (CN) of the nfd-master certificate must match the DNS name of
the nfd-master Service of the cluster. By default, nfd-master only check that
the nfd-worker has been signed by the specified root certificate (--ca-file).
Additional hardening can be enabled by specifying --verify-node-name in
nfd-master args, in which case nfd-master verifies that the NodeName presented
by nfd-worker matches the Common Name (CN) of its certificate. This means that
each nfd-worker requires a individual node-specific TLS certificate.

## Configuration

NFD-Worker supports a configuration file. The default location is
`/etc/kubernetes/node-feature-discovery/nfd-worker.conf`, but,
this can be changed by specifying the`--config` command line flag.
Configuration file is re-read on each labeling pass (determined by
`--sleep-interval`) which makes run-time re-configuration of nfd-worker
possible.

Worker configuration file is read inside the container, and thus, Volumes and
VolumeMounts are needed to make your configuration available for NFD. The
preferred method is to use a ConfigMap which provides easy deployment and
re-configurability.  For example, create a config map using the example config
as a template:

```bash
cp nfd-worker.conf.example nfd-worker.conf
vim nfd-worker.conf  # edit the configuration
kubectl create configmap nfd-worker-config --from-file=nfd-worker.conf
```

Then, configure Volumes and VolumeMounts in the Pod spec (just the relevant
snippets shown below):

```yaml
...
  containers:
      volumeMounts:
        - name: nfd-worker-config
          mountPath: "/etc/kubernetes/node-feature-discovery/"
...
  volumes:
    - name: nfd-worker-config
      configMap:
        name: nfd-worker-config
...
```

You could also use other types of volumes, of course. That is, hostPath if
different config for different nodes would be required, for example.

The (empty-by-default)
[example config](https://github.com/kubernetes-sigs/node-feature-discovery/blob/{{ site.release }}/nfd-worker.conf.example)
is used as a config in the NFD Docker image. Thus, this can be used as a default
configuration in custom-built images.

Configuration options can also be specified via the `--options` command line
flag, in which case no mounts need to be used. The same format as in the config
file must be used, i.e. JSON (or YAML). For example:

```
--options='{"sources": { "pci": { "deviceClassWhitelist": ["12"] } } }'
```

Configuration options specified from the command line will override those read
from the config file.

## Using Node Labels

Nodes with specific features can be targeted using the `nodeSelector` field. The
following example shows how to target nodes with Intel TurboBoost enabled.

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    env: test
  name: golang-test
spec:
  containers:
    - image: golang
      name: go1
  nodeSelector:
    feature.node.kubernetes.io/cpu-pstate.turbo: 'true'
```

For more details on targeting nodes, see
[node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/).

## Uninstallation

### Operator Was Used for Deployment

If you followed the deployment instructions above you can simply do:

```bash
kubectl -n nfd delete NodeFeatureDiscovery my-nfd-deployment
```

Optionally, you can also remove the namespace:

```bash
kubectl delete ns nfd
```

See the [node-feature-discovery-operator][nfd-operator] and [OLM][OLM] project
documentation for instructions for uninstalling the operator and operator
lifecycle manager, respectively.

### Manual

```bash
NFD_NS=node-feature-discovery
kubectl -n $NFD_NS delete ds nfd-worker
kubectl -n $NFD_NS delete deploy nfd-master
kubectl -n $NFD_NS delete svc nfd-master
kubectl -n $NFD_NS delete sa nfd-master
kubectl delete clusterrole nfd-master
kubectl delete clusterrolebinding nfd-master
```

### Removing Feature Labels

NFD-Master has a special `--prune` command line flag for removing all
nfd-related node labels, annotations and extended resources from the cluster.

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/{{ site.release }}/nfd-prune.yaml.template
kubectl -n node-feature-discovery wait job.batch/nfd-prune --for=condition=complete && \
    kubectl -n node-feature-discovery delete job/nfd-prune
```

**NOTE:** You must run prune before removing the RBAC rules (serviceaccount,
clusterrole and clusterrolebinding).

<!-- Links -->
[nfd-operator]: https://github.com/kubernetes-sigs/node-feature-discovery-operator
[OLM]: https://github.com/operator-framework/operator-lifecycle-manager
