package otlp

import (
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	otlpmetriccol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/metrics/v1"
	otlptracecol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/trace/v1"
	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
	otlpmetric "github.com/open-telemetry/opentelemetry-proto/gen/go/metrics/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"
)

const attrsPerResource = 20
const metricLabelCount = 2

// Generator allows to generate a ExportRequest.
type Generator struct {
	random     *rand.Rand
	tracesSent uint64
	spansSent  uint64
}

// ExportRequest represents a telemetry data export request.
type ExportRequest interface {
	proto.Message
}

func NewGenerator() *Generator {
	return &Generator{
		random: rand.New(rand.NewSource(99)),
	}
}

func (g *Generator) genRandByteString(len int) string {
	b := make([]byte, len)
	for i := range b {
		b[i] = byte(g.random.Intn(10) + 33)
	}
	return string(b)
}

func (g *Generator) GenResource() *otlpresource.Resource {
	res := &otlpresource.Resource{
		Attributes: []*otlpcommon.KeyValue{
			{
				Key:   "StartTimeUnixnano",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{IntValue: 12345678}},
			},
			{Key: "Pid", Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{IntValue: 1234}}},
			{
				Key:   "HostName",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "fakehost"}},
			},
			{
				Key:   "service.name",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "generator"}},
			},
			{
				Key:   "service.namespace",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "MyTeam"}},
			},
			{
				Key:   "telemetry.auto.version",
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{"1.2.3"}},
			},
		},
	}

	m := map[string]bool{}
	for i := len(res.Attributes); i < attrsPerResource; {
		attrName := GenRandAttrName(g.random)
		if m[attrName] {
			continue
		}
		m[attrName] = true
		i++

		res.Attributes = append(
			res.Attributes,
			&otlpcommon.KeyValue{
				Key:   attrName,
				Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: g.genRandByteString(g.random.Intn(20) + 1)}},
			},
		)
	}

	return res
}

func (g *Generator) GenerateSpanBatch(
	spansPerBatch int, attrsPerSpan int, timedEventsPerSpan int,
) *otlptracecol.ExportTraceServiceRequest {
	traceID := atomic.AddUint64(&g.tracesSent, 1)

	il := &otlptrace.InstrumentationLibrarySpans{
		InstrumentationLibrary: &otlpcommon.InstrumentationLibrary{Name: "io.opentelemetry"},
	}
	batch := &otlptracecol.ExportTraceServiceRequest{
		ResourceSpans: []*otlptrace.ResourceSpans{
			{
				Resource:                    g.GenResource(),
				InstrumentationLibrarySpans: []*otlptrace.InstrumentationLibrarySpans{il},
			},
		},
	}

	for i := 0; i < spansPerBatch; i++ {
		startTime := time.Date(2019, 10, 31, 10, 11, 12, 13, time.UTC)

		spanID := atomic.AddUint64(&g.spansSent, 1)

		// Create a span.
		span := &otlptrace.Span{
			TraceId:           GenerateTraceID(traceID),
			SpanId:            GenerateSpanID(spanID),
			Name:              "load-generator-span",
			Kind:              otlptrace.Span_CLIENT,
			StartTimeUnixNano: TimeToTimestamp(startTime),
			EndTimeUnixNano:   TimeToTimestamp(startTime.Add(time.Duration(i) * time.Millisecond)),
		}

		if attrsPerSpan >= 0 {
			// Append attributes.
			span.Attributes = []*otlpcommon.KeyValue{}

			if attrsPerSpan >= 3 {
				span.Attributes = append(
					span.Attributes,
					&otlpcommon.KeyValue{
						Key:   "load_generator.span_seq_num",
						Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{IntValue: int64(spanID)}},
					},
				)
				span.Attributes = append(
					span.Attributes,
					&otlpcommon.KeyValue{
						Key:   "load_generator.trace_seq_num",
						Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{IntValue: int64(traceID)}},
					},
				)
				span.Attributes = append(
					span.Attributes,
					&otlpcommon.KeyValue{
						Key:   "k8s.pod.name",
						Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: "superpod-123"}},
					},
				)
			}

			m := map[string]bool{}
			for j := len(span.Attributes); j < attrsPerSpan; j++ {
				attrName := GenRandAttrName(g.random)
				if m[attrName] {
					continue
				}
				m[attrName] = true
				i++

				span.Attributes = append(
					span.Attributes,
					&otlpcommon.KeyValue{
						Key:   attrName,
						Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_StringValue{StringValue: g.genRandByteString(g.random.Intn(20) + 1)}},
					},
				)
			}
		}

		if timedEventsPerSpan > 0 {
			for i := 0; i < timedEventsPerSpan; i++ {
				span.Events = append(
					span.Events, &otlptrace.Span_Event{
						TimeUnixNano: TimeToTimestamp(startTime.Add(time.Duration(i) * time.Millisecond)),
						// TimeStartDeltaNano: (time.Duration(i) * time.Millisecond).Nanoseconds(),
						Attributes: []*otlpcommon.KeyValue{
							{
								Key:   "te",
								Value: &otlpcommon.AnyValue{Value: &otlpcommon.AnyValue_IntValue{IntValue: int64(spanID)}},
							},
						},
					},
				)
			}
		}

		il.Spans = append(il.Spans, span)
	}
	return batch
}

func (g *Generator) GenerateLogBatch(logsPerBatch int, attrsPerLog int) ExportRequest {
	/*
		traceID := atomic.AddUint64(&g.tracesSent, 1)

		batch := &ExportLogsServiceRequest{ResourceLogs: []*ResourceLogs{{Resource: GenResource()}}}
		for i := 0; i < logsPerBatch; i++ {
			startTime := time.Date(2019, 10, 31, 10, 11, 12, 13, time.UTC)

			spanID := atomic.AddUint64(&g.spansSent, 1)

			// Create a log.
			log := &Log{
				TraceId:      GenerateTraceID(traceID),
				SpanId:       GenerateSpanID(spanID),
				TimeUnixnano: TimeToTimestamp(startTime.Add(time.Duration(i) * time.Millisecond)),
				EventType:    "auto_generated_event",
				Body: &AttributeValue{
					Type:        AttributeValueType_STRING,
					StringValue: fmt.Sprintf("Log message %d of %d, traceid=%q, spanid=%q", i, logsPerBatch, traceID, spanID),
				},
			}

			if attrsPerLog >= 0 {
				// Append attributes.
				log.Attributes = []*KeyValue{}

				if attrsPerLog >= 2 {
					log.Attributes = append(log.Attributes,
						&KeyValue{Key: "load_generator.span_seq_num", Type: AttributeKeyValue_INT, IntValue: int64(spanID)})
					log.Attributes = append(log.Attributes,
						&KeyValue{Key: "load_generator.trace_seq_num", Type: AttributeKeyValue_INT, IntValue: int64(traceID)})
				}

				for j := len(log.Attributes); j < attrsPerLog; j++ {
					attrName := GenRandAttrName(g.random)
					log.Attributes = append(log.Attributes,
						&KeyValue{Key: attrName, Type: AttributeKeyValue_STRING, StringValue: g.genRandByteString(g.random.Intn(20) + 1)})
				}
			}

			batch.ResourceLogs[0].Logs = append(batch.ResourceLogs[0].Logs, log)
		}
		return batch
	*/
	return nil
}

func (g *Generator) genInt64Timeseries(
	startTime time.Time, offset int, valuesPerTimeseries int,
) []*otlpmetric.Int64DataPoint {
	var timeseries []*otlpmetric.Int64DataPoint
	for j := 0; j < 1; j++ {
		var points []*otlpmetric.Int64DataPoint

		for k := 0; k < valuesPerTimeseries; k++ {
			pointTs := TimeToTimestamp(startTime.Add(time.Duration(j*k) * time.Millisecond))

			m := map[string]bool{}
			m["http.status_code"] = true

			point := otlpmetric.Int64DataPoint{
				TimeUnixNano: pointTs,
				Value:        int64(offset * j * k),
				Labels: []*otlpcommon.StringKeyValue{
					{
						Key:   "http.status_code",
						Value: strconv.Itoa(k + 200),
					},
				},
			}

			if k == 0 {
				point.StartTimeUnixNano = pointTs
			}

			for l := len(point.Labels); l < metricLabelCount; {
				attrName := GenRandAttrName(g.random)
				if m[attrName] {
					continue
				}
				m[attrName] = true
				l++

				point.Labels = append(
					point.Labels, &otlpcommon.StringKeyValue{
						Key:   attrName,
						Value: strconv.Itoa(j),
					},
				)
			}

			points = append(points, &point)
		}

		timeseries = append(timeseries, points...)
	}

	return timeseries
}

func (g *Generator) genInt64Gauge(
	startTime time.Time, i int, labelKeys []string, valuesPerTimeseries int,
) *otlpmetric.Metric {
	descr := GenMetricDescriptor(i)

	metric1 := &otlpmetric.Metric{
		MetricDescriptor: descr,
		Int64DataPoints:  g.genInt64Timeseries(startTime, i, valuesPerTimeseries),
	}

	return metric1
}

func GenMetricDescriptor(i int) *otlpmetric.MetricDescriptor {
	descr := &otlpmetric.MetricDescriptor{
		Name:        "metric" + strconv.Itoa(i),
		Description: "some description: " + strconv.Itoa(i),
		Type:        otlpmetric.MetricDescriptor_INT64,
		//Labels: []*otlpcommon.StringKeyValue{
		//	{
		//		Key:   "label1",
		//		Value: "val1",
		//	},
		//	{
		//		Key:   "label2",
		//		Value: "val2",
		//	},
		//},
	}
	return descr
}

func (g *Generator) GenerateMetricBatch(
	metricsPerBatch int,
	valuesPerTimeseries int,
	int64 bool,
) *otlpmetriccol.ExportMetricsServiceRequest {

	il := &otlpmetric.InstrumentationLibraryMetrics{}
	batch := &otlpmetriccol.ExportMetricsServiceRequest{
		ResourceMetrics: []*otlpmetric.ResourceMetrics{
			{
				Resource:                      g.GenResource(),
				InstrumentationLibraryMetrics: []*otlpmetric.InstrumentationLibraryMetrics{il},
			},
		},
	}

	for i := 0; i < metricsPerBatch; i++ {
		startTime := time.Date(2019, 10, 31, 10, 11, 12, 13, time.UTC)

		labelKeys := []string{
			"label1",
			"label2",
		}

		if int64 {
			il.Metrics = append(il.Metrics, g.genInt64Gauge(startTime, i, labelKeys, valuesPerTimeseries))
		}
	}
	return batch
}

type SpanTranslator struct {
}

func (st *SpanTranslator) TranslateSpans(batch *otlptracecol.ExportTraceServiceRequest) *otlptracecol.ExportTraceServiceRequest {
	return batch
}
