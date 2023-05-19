package compiled

import (
	otlpresource "github.com/open-telemetry/opentelemetry-proto/gen/go/resource/v1"
)

type ResourceAttributesRenameAction AttributesRenameAction

func (rt ResourceAttributesRenameAction) Apply(resource *otlpresource.Resource) ApplyResult {
	return AttributesRenameAction(rt).Apply(resource.Attributes)
}
