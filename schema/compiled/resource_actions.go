package compiled

import (
	otlpresource "go.opentelemetry.io/proto/otlp/resource/v1"
)

type ResourceAttributesRenameAction AttributesRenameAction

func (rt ResourceAttributesRenameAction) Apply(resource *otlpresource.Resource, changes *ChangeLog) error {
	return AttributesRenameAction(rt).Apply(resource.Attributes, changes)
}
