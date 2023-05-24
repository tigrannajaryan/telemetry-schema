package compiled

import (
	otlpmetric "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

type MetricRenameAction map[types.MetricName]types.MetricName

func (act MetricRenameAction) Apply(metrics []*otlpmetric.Metric) ([]*otlpmetric.Metric, error) {
	for _, metric := range metrics {
		newName, exists := act[types.MetricName(metric.Name)]
		if exists {
			metric.Name = string(newName)
		}
	}
	return metrics, nil
}

type MetricLabelRenameAction struct {
	// ApplyOnlyToMetrics limits which metrics this action should apply to. If empty then
	// there is no limitation.
	ApplyOnlyToMetrics map[types.MetricName]bool
	LabelMap           map[string]string
}

func (act MetricLabelRenameAction) Apply(metrics []*otlpmetric.Metric) (
	[]*otlpmetric.Metric, error,
) {
	var retErr error
	/*	for _, metric := range metrics {

		if len(act.ApplyOnlyToMetrics) > 0 {
			if _, exists := act.ApplyOnlyToMetrics[types.MetricName(metric.Name)]; !exists {
				continue
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
					retErr = err
				}
			}
		}

	}*/

	return metrics, retErr
}

/*
func renameLabels(labels []*otlpcommon.StringKeyValue, renameRules map[string]string) error {
	var err error
	newLabels := newFastMapStr(len(labels))
	converted := false
	for _, label := range labels {
		k := label.Key
		if convertTo, exists := renameRules[string(k)]; exists {
			k = string(convertTo)
			converted = true
		}
		if exists := newLabels.exists(k); exists {
			err = fmt.Errorf("label %s conflicts", k)
		}
		newLabels.set(k, label.Value)
	}
	if converted {
		newLabels.copyTo(labels)
	}
	return err
}
*/

type MetricSplitAction struct {
	// ApplyOnlyToMetrics limits which metrics this action should apply to. If empty then
	// there is no limitation.
	MetricName    types.MetricName
	AttributeName types.AttributeName
	SplitMap      map[types.AttributeValue]types.MetricName
}

func (act MetricSplitAction) Apply(metrics []*otlpmetric.Metric) ([]*otlpmetric.Metric, error) {
	/*
		for i := 0; i < len(metrics); i++ {
			metric := metrics[i]
			if act.MetricName != types.MetricName(metric.Name) {
				continue
			}

			var outputMetrics []*otlpmetric.Metric
			dt := metric.MetricDescriptor.Type
			switch dt {
			case otlpmetric.MetricDescriptor_INT64:
				dps := metric.Int64DataPoints
				for j := 0; j < len(dps); j++ {
					dp := dps[j]
					outputMetric := splitMetric(act.AttributeName, act.SplitMap, metric, dp)
					outputMetrics = append(outputMetrics, outputMetric)
				}
			}

			metrics = append(append(metrics[0:i], outputMetrics...), metrics[i+1:]...)
		}
	*/
	return metrics, nil
}

/*
func splitMetric(
	splitByAttr types.AttributeName,
	splitRules map[types.AttributeValue]types.MetricName,
	input *otlpmetric.Metric,
	inputDp *otlpmetric.Int64DataPoint,
) *otlpmetric.Metric {
	output := &otlpmetric.Metric{}
	descr := *input.MetricDescriptor
	output.MetricDescriptor = &descr

	outputDp := *inputDp
	outputDp.Labels = nil

	for _, label := range inputDp.Labels {
		if label.Key == string(splitByAttr) {
			if convertTo, exists := splitRules[types.AttributeValue(label.Value)]; exists {
				newMetricName := string(convertTo)
				output.MetricDescriptor.Name = newMetricName
			}
			continue
		}
		outputDp.Labels = append(outputDp.Labels, label)
	}
	output.Int64DataPoints = []*otlpmetric.Int64DataPoint{&outputDp}
	return output
}
*/
