package converter

import (
	otlpmetriccol "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	otlptracecol "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	otlpresource "go.opentelemetry.io/proto/otlp/resource/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
	"github.com/tigrannajaryan/telemetry-schema/schema/otlp"
)

func convertResource(resource *otlpresource.Resource, schema *compiled.Schema, changes *compiled.ChangeLog) error {
	return schema.ConvertResourceToLatest("0.0.0", resource, changes)
}

func convertTraceRequest(
	request *otlptracecol.ExportTraceServiceRequest, schema *compiled.Schema, changes *compiled.ChangeLog,
) error {
	for _, rss := range request.ResourceSpans {
		if err := convertResource(rss.Resource, schema, changes); err != nil {
			return err
		}

		for _, ils := range rss.ScopeSpans {
			if err := schema.ConvertSpansToLatest("0.0.0", ils.Spans, changes); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertMetricRequest(
	request *otlpmetriccol.ExportMetricsServiceRequest, schema *compiled.Schema, changes *compiled.ChangeLog,
) error {
	for _, rss := range request.ResourceMetrics {
		convertResource(rss.Resource, schema, changes)
		for _, ils := range rss.ScopeMetrics {
			if err := schema.ConvertMetricsToLatest("0.0.0", &ils.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func ConvertRequest(request otlp.ExportRequest, schema *compiled.Schema, changes *compiled.ChangeLog) error {
	switch r := request.(type) {
	case *otlptracecol.ExportTraceServiceRequest:
		return convertTraceRequest(r, schema, changes)
	case *otlpmetriccol.ExportMetricsServiceRequest:
		return convertMetricRequest(r, schema, changes)
	}
	return nil
}
