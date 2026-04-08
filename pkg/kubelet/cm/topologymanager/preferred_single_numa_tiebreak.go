/*
Copyright The Kubernetes Authors.

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

package topologymanager

import "k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"

// PreferredSingleNUMATieBreaker compares two single-NUMA preferred topology hints when
// topologyManagerPolicy is single-numa-node and prefer-most-allocated-numa-node is enabled.
type PreferredSingleNUMATieBreaker interface {
	ComparePreferredSingleNUMAForTopology(current, candidate bitmask.BitMask) (preferCandidate bool, ok bool)
}
