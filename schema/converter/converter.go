package converter

import (
	otlpmetriccol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/metrics/v1"
	otlptracecol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/trace/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
	"github.com/tigrannajaryan/telemetry-schema/schema/otlp"
)

func convertResource(resource *otlpresource.Resource, schema *compiled.Schema) compiled.ApplyResult {
	return schema.ConvertResourceToLatest("0.0.0", resource)
}

func convertTraceRequest(
	request *otlptracecol.ExportTraceServiceRequest, schema *compiled.Schema,
) (changes compiled.ApplyResult) {
	for _, rss := range request.ResourceSpans {
		change := convertResource(rss.Resource, schema)
		changes.Merge(change)
		if change.IsError() {
			return
		}

		for _, ils := range rss.InstrumentationLibrarySpans {
			r := schema.ConvertSpansToLatest("0.0.0", ils.Spans)
			changes.Merge(r)
			if r.IsError() {
				return
			}
		}
	}
	return
}

func convertMetricRequest(
	request *otlpmetriccol.ExportMetricsServiceRequest, schema *compiled.Schema,
) {
	for _, rss := range request.ResourceMetrics {
		convertResource(rss.Resource, schema)
		for _, ils := range rss.InstrumentationLibraryMetrics {
			if err := schema.ConvertMetricsToLatest("0.0.0", &ils.Metrics); err != nil {
				// logger.Debug("Conversion error", zap.Error(err))
			}
		}
	}
}

func ConvertRequest(request otlp.ExportRequest, schema *compiled.Schema) compiled.ApplyResult {
	switch r := request.(type) {
	case *otlptracecol.ExportTraceServiceRequest:
		return convertTraceRequest(r, schema)
	case *otlpmetriccol.ExportMetricsServiceRequest:
		convertMetricRequest(r, schema)
	}
	return compiled.ApplyResult{}
}
