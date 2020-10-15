#!/usr/bin/env bash
kubectl label node cluster-node01 feature.node.kubernetes.io/cpu-cpuid.IBPB-
kubectl label node cluster-node01 feature.node.kubernetes.io/cpu-cpuid.STIBP-
kubectl label node cluster-node01 feature.node.kubernetes.io/cpu-cpuid.VMX-
kubectl label node cluster-node01 feature.node.kubernetes.io/iommu-enabled-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-config.NO_HZ-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-config.NO_HZ_FULL-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-version.full-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-version.major-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-version.minor-
kubectl label node cluster-node01 feature.node.kubernetes.io/kernel-version.revision-
kubectl label node cluster-node01 feature.node.kubernetes.io/memory-numa-
kubectl label node cluster-node01 feature.node.kubernetes.io/network-sriov.capable-
kubectl label node cluster-node01 feature.node.kubernetes.io/pci-0300_1039.present-
kubectl label node cluster-node01 feature.node.kubernetes.io/pci-0300_10de.present-
kubectl label node cluster-node01 feature.node.kubernetes.io/storage-nonrotationaldisk-
kubectl label node cluster-node01 feature.node.kubernetes.io/system-os_release.ID-
kubectl label node cluster-node01 feature.node.kubernetes.io/system-os_release.VERSION_ID-
kubectl label node cluster-node01 feature.node.kubernetes.io/system-os_release.VERSION_ID.major-
kubectl annotate node cluster-node01 nfd.node.kubernetes.io/extended-resources-
kubectl annotate node cluster-node01 nfd.node.kubernetes.io/feature-labels-
kubectl annotate node cluster-node01 nfd.node.kubernetes.io/master.version-
kubectl annotate node cluster-node01 nfd.node.kubernetes.io/worker.version-

