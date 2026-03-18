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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
)

const (
	defaultTracingSampler    = "parentbased_traceidratio"
	defaultTracingSamplerArg = "0.1"
	defaultOTELServiceName   = "gateway-api-inference-extension"
)

// applyTracingConfig injects the --tracing flag and OpenTelemetry environment variables
// into the "main" container of the scheduler pod spec based on the tracing configuration.
func applyTracingConfig(llmSvc *v1alpha1.LLMInferenceService, podSpec *corev1.PodSpec) {
	tracingEnabled := llmSvc.Spec.Tracing != nil && llmSvc.Spec.Tracing.Enabled

	for i := range podSpec.Containers {
		if podSpec.Containers[i].Name != "main" {
			continue
		}

		podSpec.Containers[i].Args = append(podSpec.Containers[i].Args,
			fmt.Sprintf("--tracing=%t", tracingEnabled),
		)

		if !tracingEnabled {
			return
		}

		tracing := llmSvc.Spec.Tracing

		sampler := defaultTracingSampler
		samplerArg := defaultTracingSamplerArg
		if tracing.Sampling != nil {
			if tracing.Sampling.Sampler != "" {
				sampler = tracing.Sampling.Sampler
			}
			if tracing.Sampling.SamplerArg != "" {
				samplerArg = tracing.Sampling.SamplerArg
			}
		}

		tracingEnvVars := []corev1.EnvVar{
			{
				Name:  "OTEL_EXPORTER_OTLP_ENDPOINT",
				Value: tracing.OTLPEndpoint,
			},
			{
				Name:  "OTEL_TRACES_EXPORTER",
				Value: "otlp",
			},
			{
				Name: "OTEL_RESOURCE_ATTRIBUTES_NODE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "spec.nodeName",
					},
				},
			},
			{
				Name: "OTEL_RESOURCE_ATTRIBUTES_POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			{
				Name:  "OTEL_RESOURCE_ATTRIBUTES",
				Value: fmt.Sprintf("k8s.namespace.name=%s,k8s.node.name=$(OTEL_RESOURCE_ATTRIBUTES_NODE_NAME),k8s.pod.name=$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME)", llmSvc.GetNamespace()),
			},
			{
				Name:  "OTEL_SERVICE_NAME",
				Value: defaultOTELServiceName,
			},
			{
				Name:  "OTEL_TRACES_SAMPLER",
				Value: sampler,
			},
			{
				Name:  "OTEL_TRACES_SAMPLER_ARG",
				Value: samplerArg,
			},
		}

		podSpec.Containers[i].Env = append(podSpec.Containers[i].Env, tracingEnvVars...)
		return
	}
}
