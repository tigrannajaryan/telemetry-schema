package compiled

import (
	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
)

type keyValStr struct {
	key string
	val string
}

type fastMapStr struct {
	slice    [maxSliceSize]keyValStr
	sliceLen int
	mp       map[string]string
}

func newFastMapStr(cap int) *fastMapStr {
	m := &fastMapStr{}
	if cap > maxSliceSize {
		m.mp = make(map[string]string, cap)
	}
	return m
}

func (m *fastMapStr) set(k string, v string) {
	if m.mp == nil {
		for i := 0; i < m.sliceLen; i++ {
			if m.slice[i].key == k {
				m.slice[i].val = v
				return
			}
		}
		m.slice[m.sliceLen] = keyValStr{key: k, val: v}
		m.sliceLen++
	} else {
		m.mp[k] = v
	}
}

/*func (m *fastMapStr) get(k string) pdata.AttributeValue {
	for _, item := range m.slice {
		if item.key == k {
			return item.val
		}
	}
	return pdata.NewAttributeValueNull()
}*/

func (m *fastMapStr) exists(k string) bool {
	if m.mp == nil {
		for i := 0; i < m.sliceLen; i++ {
			if m.slice[i].key == k {
				return true
			}
		}
		return false
	} else {
		_, exists := m.mp[k]
		return exists
	}
}

func (m *fastMapStr) copyTo(attrs []*otlpcommon.StringKeyValue) {
	if m.mp == nil {
		for i := 0; i < m.sliceLen; i++ {
			attrs[i].Key = m.slice[i].key
			attrs[i].Value = m.slice[i].val
		}
	} else {
		i := 0
		for k, v := range m.mp {
			attrs[i].Key = k
			attrs[i].Value = v
			i++
		}
	}
}
