package compiled

import (
	otlptrace "github.com/open-telemetry/opentelemetry-proto/gen/go/trace/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

type SpanRenameAction map[types.SpanName]types.SpanName

func (act SpanRenameAction) Apply(span *otlptrace.Span) error {
	newName, exists := act[types.SpanName(span.Name)]
	if exists {
		span.Name = string(newName)
	}
	return nil
}

type SpanAttributeRenameAction struct {
	AttributesRenameAction

	// ApplyOnlyToSpans limits which spans this action should apply to. If empty then
	// there is no limitation.
	ApplyOnlyToSpans map[types.SpanName]bool
}

func (act SpanAttributeRenameAction) Apply(span *otlptrace.Span) error {
	if len(act.ApplyOnlyToSpans) > 0 {
		if _, exists := act.ApplyOnlyToSpans[types.SpanName(span.Name)]; !exists {
			return nil
		}
	}

	return act.AttributesRenameAction.Apply(span.Attributes)
}
