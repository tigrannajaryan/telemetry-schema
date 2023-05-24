package compiled

import (
	"fmt"

	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
)

type AttributesRenameAction map[string]string

func (at AttributesRenameAction) Apply(attrs []*otlpcommon.KeyValue, changes *ApplyResult) {
	var err error

	seenAttrs := newFastMap(len(attrs))
	var changeLog keyRenameLog

	for i, attr := range attrs {
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
				changeLog.savedAttrs = changeLog.fixedBuf[:]
				changeLog.savedAttrs = changeLog.savedAttrs[:0]
				changeLog.origAttrs = attrs
			} else if len(changeLog.savedAttrs) == len(changeLog.fixedBuf) {
				changeLog.savedAttrs = make([]savedAttrKey, len(changeLog.savedAttrs), len(changeLog.savedAttrs)+1)
				copy(changeLog.savedAttrs, changeLog.fixedBuf[:])
			}

			changeLog.savedAttrs = append(
				changeLog.savedAttrs, savedAttrKey{
					at:  i,
					key: attr.Key,
				},
			)

			attr.Key = convertTo
		}
	}

	if err != nil {
		changes.AppendError(err)
	}

	if len(changeLog.savedAttrs) > 0 {
		changes.Append(&changeLog)
	}
}

type savedAttrKey struct {
	at  int
	key string
}

type keyRenameLog struct {
	origAttrs  []*otlpcommon.KeyValue
	fixedBuf   [8]savedAttrKey
	savedAttrs []savedAttrKey
}

func (r *keyRenameLog) Rollback() {
	for i := 0; i < len(r.savedAttrs); i++ {
		r.origAttrs[r.savedAttrs[i].at].Key = r.savedAttrs[i].key
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
