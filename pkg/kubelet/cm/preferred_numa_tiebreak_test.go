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
	"testing"

	tmbitmask "k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
)

type stubNUMATopologyComparer struct {
	pick, ok bool
}

func (s stubNUMATopologyComparer) ComparePreferredSingleNUMAForTopology(tmbitmask.BitMask, tmbitmask.BitMask) (bool, bool) {
	return s.pick, s.ok
}

func TestPreferredNUMATieBreakerAggregator(t *testing.T) {
	m0, err := tmbitmask.NewBitMask(0)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := tmbitmask.NewBitMask(1)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		cpu        stubNUMATopologyComparer
		mem        stubNUMATopologyComparer
		wantPick   bool
		wantOK     bool
		current    tmbitmask.BitMask
		candidate  tmbitmask.BitMask
	}{
		{
			name:      "both undecided",
			cpu:       stubNUMATopologyComparer{ok: false},
			mem:       stubNUMATopologyComparer{ok: false},
			wantOK:    false,
			current:   m0,
			candidate: m1,
		},
		{
			name:      "cpu only",
			cpu:       stubNUMATopologyComparer{pick: true, ok: true},
			mem:       stubNUMATopologyComparer{ok: false},
			wantPick:  true,
			wantOK:    true,
			current:   m0,
			candidate: m1,
		},
		{
			name:      "memory only",
			cpu:       stubNUMATopologyComparer{ok: false},
			mem:       stubNUMATopologyComparer{pick: false, ok: true},
			wantPick:  false,
			wantOK:    true,
			current:   m0,
			candidate: m1,
		},
		{
			name:      "agree prefer candidate",
			cpu:       stubNUMATopologyComparer{pick: true, ok: true},
			mem:       stubNUMATopologyComparer{pick: true, ok: true},
			wantPick:  true,
			wantOK:    true,
			current:   m0,
			candidate: m1,
		},
		{
			name:      "disagree fallback",
			cpu:       stubNUMATopologyComparer{pick: true, ok: true},
			mem:       stubNUMATopologyComparer{pick: false, ok: true},
			wantOK:    false,
			current:   m0,
			candidate: m1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &preferredNUMATieBreakerAggregator{cpu: tc.cpu, mem: tc.mem}
			gotPick, gotOK := a.ComparePreferredSingleNUMAForTopology(tc.current, tc.candidate)
			if gotOK != tc.wantOK {
				t.Fatalf("ok: got %v want %v", gotOK, tc.wantOK)
			}
			if gotOK && gotPick != tc.wantPick {
				t.Fatalf("pick: got %v want %v", gotPick, tc.wantPick)
			}
		})
	}
}
