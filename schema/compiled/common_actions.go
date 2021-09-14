package compiled

import (
	"fmt"

	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
)

type AttributesRenameAction map[string]string

func (at AttributesRenameAction) Apply(attrs []*otlpcommon.KeyValue) error {
	var err error
	newAttrs := newFastMap(len(attrs))
	converted := false

	for _, attr := range attrs {
		k := attr.Key
		if convertTo, exists := at[k]; exists {
			k = convertTo
			converted = true
		}
		if newAttrs.exists(k) {
			err = fmt.Errorf("label %s conflicts", k)
		}
		newAttrs.set(k, attr.Value)
	}
	if converted && err == nil {
		newAttrs.copyTo(attrs)
	}
	return err
}

/*func (at AttributesRenameAction) Apply(attrs pdata.AttributeMap) error {
	var err error
	newAttrs := pdata.NewAttributeMap()
	newAttrs.InitEmptyWithCapacity(attrs.Len())
	converted := false

	attrs.ForEach(func(k string, v pdata.AttributeValue) {
		if convertTo, exists := at[k]; exists {
			k = convertTo
			converted = true
		}
		if _, exists := newAttrs.Get(k); exists {
			err = fmt.Errorf("label %s conflicts", k)
		}
		newAttrs.Insert(k, v)
	})
	if converted && err == nil {
		newAttrs.CopyTo(attrs)
	}
	return err
}
*/
