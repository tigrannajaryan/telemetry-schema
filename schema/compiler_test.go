package schema

import (
	"strconv"
	"testing"

	"github.com/gogo/protobuf/proto"
	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
	otlpmetric "github.com/open-telemetry/opentelemetry-proto/gen/go/metrics/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
)

func compileTestSchema(t testing.TB) *compiled.Schema {
	ts, err := Parse("testdata/schema-example.yaml")
	require.NoError(t, err)
	require.NotNil(t, ts)

	//l1 := ts.Metrics["1.1.0"].Current[5].Labels
	//l2 := ts.Metrics["1.1.0"].Current[6].Labels
	//fmt.Printf("%p %p\n", &l1, &l2)

	compiled := Compile(ts)
	require.NotNil(t, compiled)

	return compiled
}

func TestCompileSchema(t *testing.T) {
	compileTestSchema(t)
}

func getAttr(attrs []*otlpcommon.KeyValue, key string) (*otlpcommon.AnyValue, bool) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value, true
		}
	}
	return nil, false
}

func TestResourceSchemaConversion(t *testing.T) {
	compiled := compileTestSchema(t)

	resource := &otlpresource.Resource{}
	resource.Attributes = []*otlpcommon.KeyValue{
		{Key: "unknown-attribute", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}}},
		{Key: "k8s.cluster.name", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}}},
		{Key: "telemetry.auto.version", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}}},
	}
	resource2 := proto.Clone(resource).(*otlpresource.Resource)
	err := compiled.ConvertResourceToLatest("0.0.0", resource2)
	assert.NoError(t, err)

	assert.EqualValues(t, 3, len(resource2.Attributes))

	attrVal, exists := getAttr(resource2.Attributes, "unknown-attribute")
	assert.True(t, exists)
	assert.EqualValues(t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}}, attrVal)

	_, exists = getAttr(resource2.Attributes, "k8s.cluster.name")
	assert.False(t, exists)

	attrVal, exists = getAttr(resource2.Attributes, "kubernetes.cluster.name")
	assert.True(t, exists)
	assert.EqualValues(t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}}, attrVal)

	_, exists = getAttr(resource2.Attributes, "telemetry.auto.version")
	assert.False(t, exists)

	attrVal, exists = getAttr(resource2.Attributes, "telemetry.auto_instr.version")
	assert.True(t, exists)
	assert.EqualValues(t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}}, attrVal)
}

func getLabel(attrs []*otlpcommon.StringKeyValue, key string) (string, bool) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value, true
		}
	}
	return "", false
}

func TestMetricsSchemaConversion(t *testing.T) {
	compiled := compileTestSchema(t)

	var metrics []*otlpmetric.Metric

	metric1 := &otlpmetric.Metric{
		MetricDescriptor: &otlpmetric.MetricDescriptor{
			Name: "container.cpu.usage.total",
			Type: otlpmetric.MetricDescriptor_INT64,
		},
	}
	dp1 := &otlpmetric.Int64DataPoint{
		Labels: []*otlpcommon.StringKeyValue{
			{Key: "a", Value: "b"},
			{Key: "http.status_code", Value: "abc"},
			{Key: "status", Value: "123"},
		},
	}

	metric1.Int64DataPoints = []*otlpmetric.Int64DataPoint{dp1}
	metrics = []*otlpmetric.Metric{metric1}

	metric2 := &otlpmetric.Metric{
		MetricDescriptor: &otlpmetric.MetricDescriptor{
			Name: "unknown-metric",
			Type: otlpmetric.MetricDescriptor_INT64,
		},
	}

	dp2 := &otlpmetric.Int64DataPoint{
		Labels: []*otlpcommon.StringKeyValue{
			{Key: "c", Value: "d"},
			{Key: "http.status_code", Value: "abc"},
			{Key: "status", Value: "234"},
		},
	}

	metric2.Int64DataPoints = []*otlpmetric.Int64DataPoint{dp2}
	metrics = append(metrics, metric2)

	err := compiled.ConvertMetricsToLatest("0.0.0", metrics)
	assert.NoError(t, err)

	assert.EqualValues(t, "cpu.usage.total", metric1.MetricDescriptor.Name)
	v, _ := getLabel(dp1.Labels, "a")
	assert.EqualValues(t, "b", v)
	v, _ = getLabel(dp1.Labels, "http.response_status_code")
	assert.EqualValues(t, "abc", v)
	v, _ = getLabel(dp1.Labels, "status")
	assert.EqualValues(t, "123", v)

	assert.EqualValues(t, "unknown-metric", metric2.MetricDescriptor.Name)
	v, _ = getLabel(dp2.Labels, "c")
	assert.EqualValues(t, "d", v)
	v, _ = getLabel(dp2.Labels, "http.response_status_code")
	assert.EqualValues(t, "abc", v)
	v, _ = getLabel(dp2.Labels, "status")
	assert.EqualValues(t, "234", v)
}

func BenchmarkResourceSchemaConversion(b *testing.B) {
	compiled := compileTestSchema(b)

	var resources []*otlpresource.Resource
	for i := 0; i < b.N; i++ {
		resource := &otlpresource.Resource{}
		resource.Attributes = []*otlpcommon.KeyValue{
			{Key: "k8s.container.name", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}}},
			{Key: "k8s.cluster.name", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}}},
			{Key: "telemetry.auto.version", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}}},
		}

		for j := len(resource.Attributes); j < 20; j++ {
			resource.Attributes = append(resource.Attributes,
				&otlpcommon.KeyValue{Key: "attribute" + strconv.Itoa(j), Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{int64(j)}}})
		}

		resources = append(resources, resource)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := compiled.ConvertResourceToLatest("0.0.0", resources[i])
		assert.NoError(b, err)
	}
}
