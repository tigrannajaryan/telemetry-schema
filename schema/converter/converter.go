package converter

import (
	otlpmetriccol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/metrics/v1"
	otlptracecol "github.com/open-telemetry/opentelemetry-proto/gen/go/collector/trace/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
	"github.com/tigrannajaryan/telemetry-schema/schema/otlp"
)

func convertResource(resource *otlpresource.Resource, schema *compiled.Schema, changes *compiled.ApplyResult) {
	schema.ConvertResourceToLatest("0.0.0", resource, changes)
}

func convertTraceRequest(
	request *otlptracecol.ExportTraceServiceRequest, schema *compiled.Schema, changes *compiled.ApplyResult,
) {
	for _, rss := range request.ResourceSpans {
		convertResource(rss.Resource, schema, changes)
		if changes.IsError() {
			return
		}

		for _, ils := range rss.InstrumentationLibrarySpans {
			schema.ConvertSpansToLatest("0.0.0", ils.Spans, changes)
			if changes.IsError() {
				return
			}
		}
	}
	return
}

func convertMetricRequest(
	request *otlpmetriccol.ExportMetricsServiceRequest, schema *compiled.Schema, changes *compiled.ApplyResult,
) {
	for _, rss := range request.ResourceMetrics {
		convertResource(rss.Resource, schema, changes)
		for _, ils := range rss.InstrumentationLibraryMetrics {
			if err := schema.ConvertMetricsToLatest("0.0.0", &ils.Metrics); err != nil {
				// logger.Debug("Conversion error", zap.Error(err))
			}
		}
	}
}

func ConvertRequest(request otlp.ExportRequest, schema *compiled.Schema, changes *compiled.ApplyResult) {
	switch r := request.(type) {
	case *otlptracecol.ExportTraceServiceRequest:
		convertTraceRequest(r, schema, changes)
	case *otlpmetriccol.ExportMetricsServiceRequest:
		convertMetricRequest(r, schema, changes)
	}
}
