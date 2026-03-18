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
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
)

func TestApplyTracingConfig(t *testing.T) {
	tests := []struct {
		name               string
		llmSvc             *v1alpha1.LLMInferenceService
		podSpec            *corev1.PodSpec
		expectedArgs       []string
		expectedEnvs       map[string]corev1.EnvVar
		unexpectedEnvNames []string
	}{
		{
			name: "tracing disabled explicitly",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{Enabled: false},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedArgs:       []string{"--tracing=false"},
			unexpectedEnvNames: []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_TRACES_EXPORTER"},
		},
		{
			name: "tracing nil (not specified)",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec:       v1alpha1.LLMInferenceServiceSpec{},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedArgs:       []string{"--tracing=false"},
			unexpectedEnvNames: []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_TRACES_EXPORTER"},
		},
		{
			name: "tracing enabled with defaults",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedArgs: []string{"--tracing=true"},
			expectedEnvs: map[string]corev1.EnvVar{
				"OTEL_EXPORTER_OTLP_ENDPOINT": {Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: "http://otel-collector:4317"},
				"OTEL_TRACES_EXPORTER":        {Name: "OTEL_TRACES_EXPORTER", Value: "otlp"},
				"OTEL_SERVICE_NAME":           {Name: "OTEL_SERVICE_NAME", Value: "gateway-api-inference-extension"},
				"OTEL_TRACES_SAMPLER":         {Name: "OTEL_TRACES_SAMPLER", Value: "parentbased_traceidratio"},
				"OTEL_TRACES_SAMPLER_ARG":     {Name: "OTEL_TRACES_SAMPLER_ARG", Value: "0.1"},
				"OTEL_RESOURCE_ATTRIBUTES":    {Name: "OTEL_RESOURCE_ATTRIBUTES", Value: "k8s.namespace.name=test-ns,k8s.node.name=$(OTEL_RESOURCE_ATTRIBUTES_NODE_NAME),k8s.pod.name=$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME)"},
			},
		},
		{
			name: "tracing enabled with custom sampling",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "prod-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://custom-collector:4317",
						Sampling: &v1alpha1.TracingSamplingSpec{
							Sampler:    "always_on",
							SamplerArg: "1.0",
						},
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedArgs: []string{"--tracing=true"},
			expectedEnvs: map[string]corev1.EnvVar{
				"OTEL_EXPORTER_OTLP_ENDPOINT": {Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: "http://custom-collector:4317"},
				"OTEL_TRACES_SAMPLER":         {Name: "OTEL_TRACES_SAMPLER", Value: "always_on"},
				"OTEL_TRACES_SAMPLER_ARG":     {Name: "OTEL_TRACES_SAMPLER_ARG", Value: "1.0"},
				"OTEL_RESOURCE_ATTRIBUTES":    {Name: "OTEL_RESOURCE_ATTRIBUTES", Value: "k8s.namespace.name=prod-ns,k8s.node.name=$(OTEL_RESOURCE_ATTRIBUTES_NODE_NAME),k8s.pod.name=$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME)"},
			},
		},
		{
			name: "tracing enabled with partial sampling (only sampler)",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
						Sampling: &v1alpha1.TracingSamplingSpec{
							Sampler: "always_on",
						},
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedEnvs: map[string]corev1.EnvVar{
				"OTEL_TRACES_SAMPLER":     {Name: "OTEL_TRACES_SAMPLER", Value: "always_on"},
				"OTEL_TRACES_SAMPLER_ARG": {Name: "OTEL_TRACES_SAMPLER_ARG", Value: "0.1"},
			},
		},
		{
			name: "no main container - no changes",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "sidecar"}},
			},
			expectedArgs:       nil,
			unexpectedEnvNames: []string{"OTEL_EXPORTER_OTLP_ENDPOINT"},
		},
		{
			name: "downward API env vars for node and pod name",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main"}},
			},
			expectedEnvs: map[string]corev1.EnvVar{
				"OTEL_RESOURCE_ATTRIBUTES_NODE_NAME": {
					Name: "OTEL_RESOURCE_ATTRIBUTES_NODE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "spec.nodeName",
						},
					},
				},
				"OTEL_RESOURCE_ATTRIBUTES_POD_NAME": {
					Name: "OTEL_RESOURCE_ATTRIBUTES_POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.name",
						},
					},
				},
			},
		},
		{
			name: "preserves existing env vars",
			llmSvc: &v1alpha1.LLMInferenceService{
				ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "test-ns"},
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
					},
				},
			},
			podSpec: &corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "main",
					Env: []corev1.EnvVar{
						{Name: "SSL_CERT_DIR", Value: "/certs"},
					},
				}},
			},
			expectedEnvs: map[string]corev1.EnvVar{
				"SSL_CERT_DIR":                {Name: "SSL_CERT_DIR", Value: "/certs"},
				"OTEL_EXPORTER_OTLP_ENDPOINT": {Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: "http://otel-collector:4317"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			applyTracingConfig(tt.llmSvc, tt.podSpec)

			// Find the main container
			var mainContainer *corev1.Container
			for i := range tt.podSpec.Containers {
				if tt.podSpec.Containers[i].Name == "main" {
					mainContainer = &tt.podSpec.Containers[i]
					break
				}
			}

			if tt.expectedArgs != nil {
				g.Expect(mainContainer).NotTo(BeNil(), "expected main container to exist")
				for _, arg := range tt.expectedArgs {
					g.Expect(mainContainer.Args).To(ContainElement(arg))
				}
			}

			if tt.expectedEnvs != nil {
				g.Expect(mainContainer).NotTo(BeNil(), "expected main container to exist")
				envMap := make(map[string]corev1.EnvVar)
				for _, env := range mainContainer.Env {
					envMap[env.Name] = env
				}
				for name, expected := range tt.expectedEnvs {
					g.Expect(envMap).To(HaveKey(name))
					g.Expect(envMap[name]).To(Equal(expected))
				}
			}

			for _, envName := range tt.unexpectedEnvNames {
				if mainContainer != nil {
					for _, env := range mainContainer.Env {
						g.Expect(env.Name).NotTo(Equal(envName), "env var %s should not be present", envName)
					}
				}
			}
		})
	}
}

func TestValidateTracing(t *testing.T) {
	tests := []struct {
		name      string
		llmSvc    *v1alpha1.LLMInferenceService
		expectErr bool
		errMsg    string
	}{
		{
			name: "tracing nil - no error",
			llmSvc: &v1alpha1.LLMInferenceService{
				Spec: v1alpha1.LLMInferenceServiceSpec{},
			},
			expectErr: false,
		},
		{
			name: "tracing disabled - no error",
			llmSvc: &v1alpha1.LLMInferenceService{
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{Enabled: false},
				},
			},
			expectErr: false,
		},
		{
			name: "tracing enabled without endpoint - error",
			llmSvc: &v1alpha1.LLMInferenceService{
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{Enabled: true},
				},
			},
			expectErr: true,
			errMsg:    "spec.tracing.otlpEndpoint is required when tracing is enabled",
		},
		{
			name: "tracing enabled with endpoint - no error",
			llmSvc: &v1alpha1.LLMInferenceService{
				Spec: v1alpha1.LLMInferenceServiceSpec{
					Tracing: &v1alpha1.TracingSpec{
						Enabled:      true,
						OTLPEndpoint: "http://otel-collector:4317",
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateTracing(tt.llmSvc)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.errMsg))
				g.Expect(IsValidationError(err)).To(BeTrue())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}
