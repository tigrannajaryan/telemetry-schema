package compiled

import (
	otlpcommon "go.opentelemetry.io/proto/otlp/common/v1"
)

type keyVal struct {
	key string
	val *otlpcommon.AnyValue
}

// The maximum map size for which we use a slice. fastMap larger than this use
// a regular built-in Go map. fastMap smaller or equal to this use a slice since it
// faster for our use case. In benchmarks for our typical use cases we see the
// crossover point to be around 40. We are choosing half of that to be conservative
// since if we get much longer key strings the slice version will get worse than
// our benchmarks measure.
const maxSliceSize = 20

type fastMap struct {
	slice    [maxSliceSize]keyVal
	sliceLen int
	mp       map[string]*otlpcommon.AnyValue
}

func newFastMap(cap int) *fastMap {
	m := &fastMap{}
	if cap > maxSliceSize {
		m.mp = make(map[string]*otlpcommon.AnyValue, cap)
	}
	return m
}

func (m *fastMap) set(k string, v *otlpcommon.AnyValue) {
	if m.mp == nil {
		for i := 0; i < m.sliceLen; i++ {
			if m.slice[i].key == k {
				m.slice[i].val = v
				return
			}
		}
		m.slice[m.sliceLen] = keyVal{key: k, val: v}
		m.sliceLen++
	} else {
		m.mp[k] = v
	}
}

/*func (m *fastMap) get(k string) pdata.AttributeValue {
	for _, item := range m.slice {
		if item.key == k {
			return item.val
		}
	}
	return pdata.NewAttributeValueNull()
}*/

func (m *fastMap) exists(k string) bool {
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

func (m *fastMap) copyTo(attrs []*otlpcommon.KeyValue) {
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
