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

package cpumanager

import (
	"testing"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
	"k8s.io/kubernetes/test/utils/ktesting"
	"k8s.io/utils/cpuset"
)

func TestComparePreferredSingleNUMAForTopology_CPUScoring(t *testing.T) {
	logger, _ := ktesting.NewTestContext(t)
	mi := returnMachineInfo()
	topo, err := topology.Discover(logger, &mi)
	if err != nil {
		t.Fatal(err)
	}
	staticPol, err := NewStaticPolicy(logger, topo, 0, cpuset.New(), topologymanager.NewFakeManager(), nil)
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

	t.Run("equal exclusive counts is undecided", func(t *testing.T) {
		st := &mockState{assignments: state.ContainerCPUAssignments{
			"p1": {"c": mustParseCPUSet(t, "0")},
			"p2": {"c": mustParseCPUSet(t, "3")},
		}}
		m := &manager{policy: staticPol, state: st}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=false", ok, pick)
		}
	})

	t.Run("strictly more exclusive on candidate NUMA prefers candidate", func(t *testing.T) {
		st := &mockState{assignments: state.ContainerCPUAssignments{
			"p1": {"c": mustParseCPUSet(t, "0")},
			"p2": {"c": mustParseCPUSet(t, "3")},
			"p3": {"c": mustParseCPUSet(t, "4")},
			"p4": {"c": mustParseCPUSet(t, "5")},
		}}
		m := &manager{policy: staticPol, state: st}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if !ok || !pick {
			t.Fatalf("got ok=%v pick=%v want ok=true pick=true", ok, pick)
		}
	})

	t.Run("strictly more exclusive on current NUMA prefers current", func(t *testing.T) {
		st := &mockState{assignments: state.ContainerCPUAssignments{
			"p1": {"c": mustParseCPUSet(t, "0")},
			"p2": {"c": mustParseCPUSet(t, "1")},
			"p3": {"c": mustParseCPUSet(t, "2")},
			"p4": {"c": mustParseCPUSet(t, "3")},
		}}
		m := &manager{policy: staticPol, state: st}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if !ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=true pick=false", ok, pick)
		}
	})

	t.Run("non-static policy is undecided", func(t *testing.T) {
		st := &mockState{assignments: state.ContainerCPUAssignments{
			"p1": {"c": mustParseCPUSet(t, "0")},
		}}
		m := &manager{policy: &mockPolicy{}, state: st}
		pick, ok := m.ComparePreferredSingleNUMAForTopology(m0, m1)
		if ok || pick {
			t.Fatalf("got ok=%v pick=%v want ok=false", ok, pick)
		}
	})
}
