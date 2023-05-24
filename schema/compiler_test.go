package schema

import (
	"strconv"
	"testing"

	"github.com/gogo/protobuf/proto"
	otlptracecol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/trace/v1"
	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
	otlpmetric "github.com/open-telemetry/opentelemetry-proto/gen/go/metrics/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
	"github.com/tigrannajaryan/telemetry-schema/schema/converter"
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
	schema := compileTestSchema(t)

	resource := &otlpresource.Resource{}
	resource.Attributes = []*otlpcommon.KeyValue{
		{
			Key:   "unknown-attribute",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}},
		},
		{
			Key:   "k8s.cluster.name",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}},
		},
		{
			Key:   "telemetry.auto.version",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}},
		},
	}
	resource2 := proto.Clone(resource).(*otlpresource.Resource)
	changes := compiled.ApplyResult{}
	schema.ConvertResourceToLatest("0.0.0", resource2, &changes)
	assert.False(t, changes.IsError())

	assert.EqualValues(t, 3, len(resource2.Attributes))

	attrVal, exists := getAttr(resource2.Attributes, "unknown-attribute")
	assert.True(t, exists)
	assert.EqualValues(t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}}, attrVal)

	_, exists = getAttr(resource2.Attributes, "k8s.cluster.name")
	assert.False(t, exists)

	attrVal, exists = getAttr(resource2.Attributes, "kubernetes.cluster.name")
	assert.True(t, exists)
	assert.EqualValues(
		t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}}, attrVal,
	)

	_, exists = getAttr(resource2.Attributes, "telemetry.auto.version")
	assert.False(t, exists)

	attrVal, exists = getAttr(resource2.Attributes, "telemetry.auto_instr.version")
	assert.True(t, exists)
	assert.EqualValues(
		t, &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}}, attrVal,
	)
}

func TestResourceSchemaConversionConflict(t *testing.T) {
	schema := compileTestSchema(t)

	resource1 := &otlpresource.Resource{}
	resource1.Attributes = []*otlpcommon.KeyValue{
		{
			Key:   "k8s.cluster.name",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}},
		},
		{
			Key:   "telemetry.auto.version",
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}},
		},
	}
	resource2 := proto.Clone(resource1).(*otlpresource.Resource)
	resource2.Attributes = append(
		resource2.Attributes, &otlpcommon.KeyValue{
			Key:   "kubernetes.cluster.name", // This should conflict with conversion
			Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}},
		},
	)

	request := &otlptracecol.ExportTraceServiceRequest{
		ResourceSpans: []*otlptrace.ResourceSpans{
			{
				Resource: resource1,
			},
			{
				Resource: resource2,
			},
		},
	}

	requestCopy := proto.Clone(request)

	changes := &compiled.ApplyResult{}
	converter.ConvertRequest(request, schema, changes)
	assert.True(t, changes.IsError())
	assert.False(t, proto.Equal(request, requestCopy))

	changes.Rollback()
	assert.True(t, proto.Equal(request, requestCopy))
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

	metric3 := &otlpmetric.Metric{
		MetricDescriptor: &otlpmetric.MetricDescriptor{
			Name: "system.paging.operations",
			Type: otlpmetric.MetricDescriptor_INT64,
		},
	}

	dp3 := &otlpmetric.Int64DataPoint{
		Labels: []*otlpcommon.StringKeyValue{
			{Key: "direction", Value: "in"},
			{Key: "http.status_code", Value: "abc"},
		},
	}
	dp4 := &otlpmetric.Int64DataPoint{
		Labels: []*otlpcommon.StringKeyValue{
			{Key: "direction", Value: "out"},
			{Key: "http.status_code", Value: "abc"},
		},
	}

	metric3.Int64DataPoints = []*otlpmetric.Int64DataPoint{dp3, dp4}
	metrics = append(metrics, metric3)

	err := compiled.ConvertMetricsToLatest("0.0.0", &metrics)
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

	assert.EqualValues(t, "system.paging.operations.in", metrics[2].MetricDescriptor.Name)
	assert.Len(t, metrics[2].Int64DataPoints[0].Labels, 1)
	v, _ = getLabel(metrics[2].Int64DataPoints[0].Labels, "http.response_status_code")
	assert.EqualValues(t, "abc", v)

	assert.EqualValues(t, "system.paging.operations.out", metrics[3].MetricDescriptor.Name)
	assert.Len(t, metrics[3].Int64DataPoints[0].Labels, 1)
	v, _ = getLabel(metrics[3].Int64DataPoints[0].Labels, "http.response_status_code")
	assert.EqualValues(t, "abc", v)
}

func BenchmarkResourceSchemaConversion(b *testing.B) {
	schema := compileTestSchema(b)

	var resources []*otlpresource.Resource
	for i := 0; i < b.N; i++ {
		resource := &otlpresource.Resource{}
		resource.Attributes = []*otlpcommon.KeyValue{
			{
				Key:   "k8s.container.name",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{123}},
			},
			{
				Key:   "k8s.cluster.name",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"OnlineShop"}},
			},
			{
				Key:   "telemetry.auto.version",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}},
			},
		}

		for j := len(resource.Attributes); j < 15; j++ {
			resource.Attributes = append(
				resource.Attributes,
				&otlpcommon.KeyValue{
					Key:   "attribute" + strconv.Itoa(j),
					Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{int64(j)}},
				},
			)
		}

		resources = append(resources, resource)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		changes := compiled.ApplyResult{}
		schema.ConvertResourceToLatest("0.0.0", resources[i], &changes)
		assert.False(b, changes.IsError())
	}
}
