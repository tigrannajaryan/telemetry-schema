package compiled

import (
	"sort"

	otlpmetric "github.com/open-telemetry/opentelemetry-proto/gen/go/metrics/v1"
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

type Schema struct {
	Versions ActionsForVersions
}

type ActionsForVersions []*ActionsForVersion

// ActionsForVersion is the cumulative list of actions to apply in order
// to convert the schema from the specified Version to the latest Version.
type ActionsForVersion struct {
	VersionNum types.TelemetryVersion
	Resource   ResourceActions
	Spans      SpanActions
	Metrics    MetricActions
	//Logs     []LogRecordAction
}

type ResourceActions []ResourceAction

func (acts ResourceActions) Apply(resource *otlpresource.Resource) error {
	for _, a := range acts {
		err := a.Apply(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

type MetricActions struct {
	ByName       map[types.MetricName][]MetricAction
	OtherMetrics []MetricAction
}

func (acts MetricActions) Apply(metric *otlpmetric.Metric) error {
	metricName := metric.MetricDescriptor.Name
	actions, exists := acts.ByName[types.MetricName(metricName)]
	if !exists {
		actions = acts.OtherMetrics
	}

	for _, a := range actions {
		err := a.Apply(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

type ResourceAction interface {
	Apply(resource *otlpresource.Resource) error
}

type SpanAction interface {
	Apply(trace *otlptrace.Span) error
}

type SpanActions struct {
	ForAllSpans []SpanAction
}

func (acts SpanActions) Apply(span *otlptrace.Span) error {
	for _, a := range acts.ForAllSpans {
		err := a.Apply(span)
		if err != nil {
			return err
		}
	}
	return nil
}

type MetricAction interface {
	Apply(metric *otlpmetric.Metric) error
}

//type LogRecordAction interface {
//	Apply(log pdata.LogRecord) error
//}

func (afv ActionsForVersions) Len() int {
	return len(afv)
}

func (afv ActionsForVersions) Less(i, j int) bool {
	return afv[i].VersionNum < afv[j].VersionNum
}

func (afv ActionsForVersions) Swap(i, j int) {
	afv[i], afv[j] = afv[j], afv[i]
}

func (s *Schema) ConvertResourceToLatest(fromVersion types.TelemetryVersion, resource *otlpresource.Resource) error {
	startIndex := sort.Search(len(s.Versions), func(i int) bool {
		// TODO: use proper semver comparison.
		return s.Versions[i].VersionNum > fromVersion
	})
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		if err := s.Versions[i].Resource.Apply(resource); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) ConvertSpansToLatest(fromVersion types.TelemetryVersion, spans []*otlptrace.Span) error {
	startIndex := sort.Search(len(s.Versions), func(i int) bool {
		// TODO: use proper semver comparison.
		return s.Versions[i].VersionNum > fromVersion
	})
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		for j := 0; j < len(spans); j++ {
			span := spans[j]
			if err := s.Versions[i].Spans.Apply(span); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Schema) ConvertMetricsToLatest(fromVersion types.TelemetryVersion, metrics []*otlpmetric.Metric) error {
	startIndex := sort.Search(len(s.Versions), func(i int) bool {
		// TODO: use proper semver comparison.
		return s.Versions[i].VersionNum > fromVersion
	})
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		for j := 0; j < len(metrics); j++ {
			metric := metrics[j]
			if err := s.Versions[i].Metrics.Apply(metric); err != nil {
				return err
			}
		}
	}

	return nil
}
