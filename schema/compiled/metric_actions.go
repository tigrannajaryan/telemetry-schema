package compiled

import (
	"fmt"

	otlpcommon "github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"
	otlpmetric "github.com/open-telemetry/opentelemetry-proto/gen/go/metrics/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

type MetricRenameAction map[types.MetricName]types.MetricName

func (act MetricRenameAction) Apply(metric *otlpmetric.Metric) error {
	newName, exists := act[types.MetricName(metric.MetricDescriptor.Name)]
	if exists {
		metric.MetricDescriptor.Name = string(newName)
	}
	return nil
}

type MetricLabelRenameAction struct {
	// ApplyOnlyToMetrics limits which metrics this action should apply to. If empty then
	// there is no limitation.
	ApplyOnlyToMetrics map[types.MetricName]bool
	LabelMap           map[string]string
}

func (act MetricLabelRenameAction) Apply(metric *otlpmetric.Metric) error {
	if len(act.ApplyOnlyToMetrics) > 0 {
		if _, exists := act.ApplyOnlyToMetrics[types.MetricName(metric.MetricDescriptor.Name)]; !exists {
			return nil
		}
	}

	dt := metric.MetricDescriptor.Type
	switch dt {
	case otlpmetric.MetricDescriptor_INT64:
		dps := metric.Int64DataPoints
		for i := 0; i < len(dps); i++ {
			dp := dps[i]
			err := renameLabels(dp.Labels, act.LabelMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func renameLabels(labels []*otlpcommon.StringKeyValue, renameRules map[string]string) error {
	var err error
	newLabels := map[string]string{}
	converted := false
	for _, label := range labels {
		k := label.Key
		if convertTo, exists := renameRules[string(k)]; exists {
			k = string(convertTo)
			converted = true
		}
		if _, exists := newLabels[k]; exists {
			err = fmt.Errorf("label %s conflicts", k)
		}
		newLabels[k] = label.Value
	}
	if converted {
		i := 0
		for k, v := range newLabels {
			labels[i].Key = k
			labels[i].Value = v
			i++
		}
	}
	return err
}
