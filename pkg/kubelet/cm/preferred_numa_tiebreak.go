/*
Copyright 2026 The Kubernetes Authors.

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

package cm

import (
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/memorymanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager"
	tmbitmask "k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
)

type singleNUMATopologyComparer interface {
	ComparePreferredSingleNUMAForTopology(current, candidate tmbitmask.BitMask) (preferCandidate bool, ok bool)
}

type preferredNUMATieBreakerAggregator struct {
	cpu singleNUMATopologyComparer
	mem singleNUMATopologyComparer
}

// NewPreferredNUMATieBreakerAggregator combines CPU and memory static-manager signals for
// prefer-most-allocated-numa-node. If both managers decide but disagree which NUMA is more
// packed, the aggregator returns ok=false so Topology Manager falls back to Narrowest.
func NewPreferredNUMATieBreakerAggregator(cpu cpumanager.Manager, mem memorymanager.Manager) topologymanager.PreferredSingleNUMATieBreaker {
	return &preferredNUMATieBreakerAggregator{cpu: cpu, mem: mem}
}

func (a *preferredNUMATieBreakerAggregator) ComparePreferredSingleNUMAForTopology(current, candidate tmbitmask.BitMask) (preferCandidate bool, ok bool) {
	pickCPU, okCPU := a.cpu.ComparePreferredSingleNUMAForTopology(current, candidate)
	pickMem, okMem := a.mem.ComparePreferredSingleNUMAForTopology(current, candidate)
	if !okCPU && !okMem {
		return false, false
	}
	if okCPU && !okMem {
		return pickCPU, true
	}
	if !okCPU && okMem {
		return pickMem, true
	}
	if pickCPU == pickMem {
		return pickCPU, true
	}
	return false, false
}
