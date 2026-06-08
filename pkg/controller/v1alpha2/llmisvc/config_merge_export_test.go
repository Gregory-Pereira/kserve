/*
Copyright 2025 The KServe Authors.

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

package llmisvc

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// SetUseVersionedConfigForTest overrides the useVersionedConfig flag for testing
// and returns a cleanup function that restores the original value.
func SetUseVersionedConfigForTest(enabled bool) func() {
	original := useVersionedConfig
	useVersionedConfig = enabled
	return func() {
		useVersionedConfig = original
	}
}

// DetectGPUResourceTypes exposes detectGPUResourceTypes for testing.
func DetectGPUResourceTypes(podSpecs ...*corev1.PodSpec) sets.Set[string] {
	return detectGPUResourceTypes(podSpecs...)
}
