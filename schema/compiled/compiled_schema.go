package compiled

import (
	"sort"

	otlpmetric "go.opentelemetry.io/proto/otlp/metrics/v1"
	otlpresource "go.opentelemetry.io/proto/otlp/resource/v1"
	otlptrace "go.opentelemetry.io/proto/otlp/trace/v1"

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

func (acts ResourceActions) Apply(resource *otlpresource.Resource, changes *ChangeLog) error {
	for _, a := range acts {
		if err := a.Apply(resource, changes); err != nil {
			return err
		}
	}
	return nil
}

type MetricActions struct {
	Actions []MetricAction
}

func (acts MetricActions) Apply(metrics []*otlpmetric.Metric) ([]*otlpmetric.Metric, error) {
	for _, a := range acts.Actions {
		var err error
		metrics, err = a.Apply(metrics)
		if err != nil {
			return metrics, err
		}
	}
	return metrics, nil
}

type ResourceAction interface {
	Apply(resource *otlpresource.Resource, changes *ChangeLog) error
}

type ChangeLog struct {
	log []Change
}

type Change interface {
	Rollback()
}

func (ar *ChangeLog) Merge(other ChangeLog) {
	ar.log = append(ar.log, other.log...)
}

func (ar *ChangeLog) Rollback() {
	for i := len(ar.log) - 1; i >= 0; i-- {
		ar.log[i].Rollback()
	}
}

func (ar *ChangeLog) Append(f Change) {
	ar.log = append(ar.log, f)
}

type SpanAction interface {
	Apply(trace *otlptrace.Span, changes *ChangeLog) error
}

type SpanActions struct {
	ForAllSpans []SpanAction
}

func (acts SpanActions) Apply(span *otlptrace.Span, changes *ChangeLog) error {
	for _, a := range acts.ForAllSpans {
		if err := a.Apply(span, changes); err != nil {
			return err
		}
	}
	return nil
}

type MetricAction interface {
	Apply(metrics []*otlpmetric.Metric) ([]*otlpmetric.Metric, error)
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

func (s *Schema) ConvertResourceToLatest(
	fromVersion types.TelemetryVersion, resource *otlpresource.Resource, changes *ChangeLog,
) error {
	startIndex := sort.Search(
		len(s.Versions), func(i int) bool {
			// TODO: use proper semver comparison.
			return s.Versions[i].VersionNum > fromVersion
		},
	)
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		if err := s.Versions[i].Resource.Apply(resource, changes); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) ConvertSpansToLatest(
	fromVersion types.TelemetryVersion, spans []*otlptrace.Span, changes *ChangeLog,
) error {
	startIndex := sort.Search(
		len(s.Versions), func(i int) bool {
			// TODO: use proper semver comparison.
			return s.Versions[i].VersionNum > fromVersion
		},
	)
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		for j := 0; j < len(spans); j++ {
			span := spans[j]
			if err := s.Versions[i].Spans.Apply(span, changes); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Schema) ConvertMetricsToLatest(
	fromVersion types.TelemetryVersion, metrics *[]*otlpmetric.Metric,
) error {
	startIndex := sort.Search(
		len(s.Versions), func(i int) bool {
			// TODO: use proper semver comparison.
			return s.Versions[i].VersionNum > fromVersion
		},
	)
	if startIndex > len(s.Versions) {
		// Nothing to do
		return nil
	}

	for i := startIndex; i < len(s.Versions); i++ {
		var err error
		*metrics, err = s.Versions[i].Metrics.Apply(*metrics)
		if err != nil {
			return err
		}
	}

	return nil
}
