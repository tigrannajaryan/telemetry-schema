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

func (acts ResourceActions) Apply(resource *otlpresource.Resource) (changes ApplyResult) {
	for _, a := range acts {
		change := a.Apply(resource)
		changes.Merge(change)
		if change.IsError() {
			break
		}
	}
	return changes
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
	Apply(resource *otlpresource.Resource) ApplyResult
}

type ApplyResult struct {
	errs     []error
	rollback []Rollbacker
}

type Rollbacker interface {
	Rollback()
}

func (ar *ApplyResult) IsError() bool {
	return len(ar.errs) > 0
}

func (ar *ApplyResult) Merge(next ApplyResult) {
	ar.rollback = append(ar.rollback, next.rollback...)
	ar.errs = append(ar.errs, next.errs...)
}

func (ar *ApplyResult) Rollback() {
	for i := len(ar.rollback) - 1; i >= 0; i-- {
		ar.rollback[i].Rollback()
	}
}

func (ar *ApplyResult) Append(f Rollbacker) {
	ar.rollback = append(ar.rollback, f)
}

func (ar *ApplyResult) AppendError(err error) {
	if err != nil {
		ar.errs = append(ar.errs, err)
	}
}

type SpanAction interface {
	Apply(trace *otlptrace.Span) ApplyResult
}

type SpanActions struct {
	ForAllSpans []SpanAction
}

func (acts SpanActions) Apply(span *otlptrace.Span) (changes ApplyResult) {
	for _, a := range acts.ForAllSpans {
		change := a.Apply(span)
		changes.Merge(change)
		if change.IsError() {
			break
		}
	}
	return changes
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
	fromVersion types.TelemetryVersion, resource *otlpresource.Resource,
) (changes ApplyResult) {
	startIndex := sort.Search(
		len(s.Versions), func(i int) bool {
			// TODO: use proper semver comparison.
			return s.Versions[i].VersionNum > fromVersion
		},
	)
	if startIndex > len(s.Versions) {
		// Nothing to do
		return
	}

	for i := startIndex; i < len(s.Versions); i++ {
		change := s.Versions[i].Resource.Apply(resource)
		changes.Merge(change)
		if change.IsError() {
			break
		}
	}

	return
}

func (s *Schema) ConvertSpansToLatest(
	fromVersion types.TelemetryVersion, spans []*otlptrace.Span,
) (ret ApplyResult) {
	startIndex := sort.Search(
		len(s.Versions), func(i int) bool {
			// TODO: use proper semver comparison.
			return s.Versions[i].VersionNum > fromVersion
		},
	)
	if startIndex > len(s.Versions) {
		// Nothing to do
		return ret
	}

	for i := startIndex; i < len(s.Versions); i++ {
		for j := 0; j < len(spans); j++ {
			span := spans[j]
			r := s.Versions[i].Spans.Apply(span)
			ret.Merge(r)
			if r.IsError() {
				return ret
			}
		}
	}

	return ret
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
