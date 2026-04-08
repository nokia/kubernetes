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

package memorymanager

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/kubelet/cm/containermap"
	"k8s.io/kubernetes/pkg/kubelet/cm/memorymanager/state"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
	"k8s.io/kubernetes/test/utils/ktesting"
)

func memBlock(numa int, sz uint64) state.Block {
	return state.Block{
		NUMAAffinity: []int{numa},
		Type:         v1.ResourceMemory,
		Size:         sz,
	}
}

func TestComparePreferredSingleNUMAForTopology_MemoryScoring(t *testing.T) {
	logger, _ := ktesting.NewTestContext(t)
	mi := returnMachineInfo()
	reserved := systemReservedMemory{
		0: {v1.ResourceMemory: 1 * gb},
		1: {v1.ResourceMemory: 1 * gb},
	}
	staticPol, err := NewPolicyStatic(logger, &mi, reserved, topologymanager.NewFakeManager())
	if err != nil {
		t.Fatal(err)
	}

	m0, err := bitmask.NewBitMask(0)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := bitmask.NewBitMask(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("equal regular memory per NUMA is undecided", func(t *testing.T) {
		st := state.NewMemoryState(logger)
		st.SetMemoryAssignments(state.ContainerMemoryAssignments{
			"p1": {"c": {memBlock(0, 100)}},
			"p2": {"c": {memBlock(1, 100)}},
		})
		m := &manager{policy: staticPol, state: st, containerMap: containermap.NewContainerMap()}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=false", ok, pick)
		}
	})

	t.Run("strictly more regular memory on candidate NUMA prefers candidate", func(t *testing.T) {
		st := state.NewMemoryState(logger)
		st.SetMemoryAssignments(state.ContainerMemoryAssignments{
			"p1": {"c": {memBlock(0, 50)}},
			"p2": {"c": {memBlock(1, 200)}},
		})
		m := &manager{policy: staticPol, state: st, containerMap: containermap.NewContainerMap()}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if !ok || !pick {
			t.Fatalf("got ok=%v pick=%v want ok=true pick=true", ok, pick)
		}
	})

	t.Run("strictly more regular memory on current NUMA prefers current", func(t *testing.T) {
		st := state.NewMemoryState(logger)
		st.SetMemoryAssignments(state.ContainerMemoryAssignments{
			"p1": {"c": {memBlock(0, 400)}},
			"p2": {"c": {memBlock(1, 100)}},
		})
		m := &manager{policy: staticPol, state: st, containerMap: containermap.NewContainerMap()}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if !ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=true pick=false", ok, pick)
		}
	})

	t.Run("non-static policy is undecided", func(t *testing.T) {
		st := state.NewMemoryState(logger)
		st.SetMemoryAssignments(state.ContainerMemoryAssignments{
			"p1": {"c": {memBlock(0, 100)}},
		})
		m := &manager{policy: &mockPolicy{}, state: st, containerMap: containermap.NewContainerMap()}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=false", ok, pick)
		}
	})
}
