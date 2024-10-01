// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifestutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
)

func TestDefaultAnnotations(t *testing.T) {
	// prepare
	otelcol := v1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "my-ns",
		},
		Spec: v1beta1.OpenTelemetryCollectorSpec{
			Config: v1beta1.Config{
				Service: v1beta1.Service{
					Extensions: []string{"test"},
				},
			},
		},
	}

	// test
	podAnnotations, err := PodAnnotations(otelcol, []string{})
	require.NoError(t, err)

	//verify propagation from metadata.annotations to spec.template.spec.metadata.annotations
	assert.Equal(t, "true", podAnnotations["prometheus.io/scrape"])
	assert.Equal(t, "8888", podAnnotations["prometheus.io/port"])
	assert.Equal(t, "/metrics", podAnnotations["prometheus.io/path"])
	assert.Equal(t, "da678471a1bebed4dc643526ec1ea20572b787f5b37da985adc6e50de30ddd3e", podAnnotations["opentelemetry-operator-config/sha256"])
}

func TestNonDefaultPodAnnotation(t *testing.T) {
	// prepare
	otelcol := v1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "my-ns",
		},
		Spec: v1beta1.OpenTelemetryCollectorSpec{
			Observability: v1beta1.ObservabilitySpec{
				Metrics: v1beta1.MetricsConfigSpec{
					DisablePrometheusAnnotations: true,
				},
			},
		},
	}

	// test
	annotations, err := Annotations(otelcol, []string{})
	require.NoError(t, err)
	podAnnotations, err := PodAnnotations(otelcol, []string{})
	require.NoError(t, err)

	//verify
	assert.NotContains(t, annotations, "prometheus.io/scrape", "Prometheus scrape annotation should not exist")
	assert.NotContains(t, annotations, "prometheus.io/port", "Prometheus port annotation should not exist")
	assert.NotContains(t, annotations, "prometheus.io/path", "Prometheus path annotation should not exist")
	//verify propagation from metadata.annotations to spec.template.spec.metadata.annotations
	assert.NotContains(t, podAnnotations, "prometheus.io/scrape", "Prometheus scrape annotation should not exist in pod annotations")
	assert.NotContains(t, podAnnotations, "prometheus.io/port", "Prometheus port annotation should not exist in pod annotations")
	assert.NotContains(t, podAnnotations, "prometheus.io/path", "Prometheus path annotation should not exist in pod annotations")
	assert.Equal(t, "6207c0917232f2b3514cc175299e1639adcec9fa4e14f54e6e95ba035d3a0e4f", podAnnotations["opentelemetry-operator-config/sha256"])
}

func TestUserAnnotations(t *testing.T) {
	// prepare
	otelcol := v1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "my-ns",
			Annotations: map[string]string{"prometheus.io/scrape": "false",
				"prometheus.io/port": "1234",
				"prometheus.io/path": "/test",
			},
		},
		Spec: v1beta1.OpenTelemetryCollectorSpec{
			Config: v1beta1.Config{
				Service: v1beta1.Service{
					Extensions: []string{"test2"},
				},
			},
		},
	}

	// test
	annotations, err := Annotations(otelcol, []string{})
	require.NoError(t, err)
	podAnnotations, err := PodAnnotations(otelcol, []string{})
	require.NoError(t, err)

	//verify
	assert.Equal(t, "false", annotations["prometheus.io/scrape"])
	assert.Equal(t, "1234", annotations["prometheus.io/port"])
	assert.Equal(t, "/test", annotations["prometheus.io/path"])
	assert.Equal(t, "aa48219365373b529e2b506136fe65214b821b6616b9fa84bc9ea1c0e6f85f75", podAnnotations["opentelemetry-operator-config/sha256"])
}

func TestAnnotationsPropagateDown(t *testing.T) {
	// prepare
	otelcol := v1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"myapp": "mycomponent"},
		},
		Spec: v1beta1.OpenTelemetryCollectorSpec{
			OpenTelemetryCommonFields: v1beta1.OpenTelemetryCommonFields{
				PodAnnotations: map[string]string{"pod_annotation": "pod_annotation_value"},
			},
		},
	}

	// test
	annotations, err := Annotations(otelcol, []string{})
	require.NoError(t, err)
	podAnnotations, err := PodAnnotations(otelcol, []string{})
	require.NoError(t, err)

	// verify
	assert.Len(t, annotations, 1)
	assert.Equal(t, "mycomponent", annotations["myapp"])
	assert.Equal(t, "mycomponent", podAnnotations["myapp"])
	assert.Equal(t, "pod_annotation_value", podAnnotations["pod_annotation"])
}

func TestAnnotationsFilter(t *testing.T) {
	otelcol := v1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"test.bar.io":  "foo",
				"test.io/port": "1234",
				"test.io/path": "/test",
			},
		},
		Spec: v1beta1.OpenTelemetryCollectorSpec{
			Mode: "deployment",
		},
	}

	// This requires the filter to be in regex match form and not the other simpler wildcard one.
	annotations, err := Annotations(otelcol, []string{".*\\.bar\\.io"})

	// verify
	require.NoError(t, err)
	assert.Len(t, annotations, 2)
	assert.NotContains(t, annotations, "test.bar.io")
	assert.Equal(t, "1234", annotations["test.io/port"])
}
