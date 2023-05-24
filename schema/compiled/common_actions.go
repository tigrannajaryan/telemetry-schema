package compiled

import (
	"fmt"

	otlpcommon "go.opentelemetry.io/proto/otlp/common/v1"
)

type AttributesRenameAction map[string]string

func (at AttributesRenameAction) Apply(attrs []*otlpcommon.KeyValue, changes *ChangeLog) error {
	var err error

	seenAttrs := newFastMap(len(attrs))
	var changeLog attrsModifyLog

	for _, attr := range attrs {
		if seenAttrs.exists(attr.Key) {
			err = fmt.Errorf("attribute %s conflicts", attr.Key)
			break
		}

		seenAttrs.set(attr.Key, attr.Value)

		if convertTo, exists := at[attr.Key]; exists {
			if seenAttrs.exists(convertTo) {
				err = fmt.Errorf("attribute %s conflicts", attr.Key)
				break
			}
			seenAttrs.set(convertTo, attr.Value)

			if changeLog.savedAttrs == nil {
				changeLog.origAttrs = attrs
				changeLog.savedAttrs = make([]otlpcommon.KeyValue, len(attrs))
				for j, attr := range attrs {
					changeLog.savedAttrs[j] = *attr
				}
			}

			attr.Key = convertTo
		}
	}

	if len(changeLog.savedAttrs) > 0 {
		changes.Append(&changeLog)
	}

	return err
}

type attrsModifyLog struct {
	origAttrs  []*otlpcommon.KeyValue
	savedAttrs []otlpcommon.KeyValue
}

func (r *attrsModifyLog) Rollback() {
	for i := 0; i < len(r.savedAttrs); i++ {
		*r.origAttrs[i] = r.savedAttrs[i]
	}
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
