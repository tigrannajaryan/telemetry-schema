package ast

import "github.com/tigrannajaryan/telemetry-schema/schema/types"

type VersionOfMetrics struct {
	Changes []MetricTranslationAction
	Current []MetricSchema `yaml:"current_metric_schema"`
}

type MetricTranslationAction struct {
	RenameMetrics   map[types.MetricName]types.MetricName `yaml:"rename_metrics"`
	RenameLabels    *LabelMapForMetrics                   `yaml:"rename_labels"`
	AddLabels       *LabelMapForMetrics                   `yaml:"add_labels"`
	DuplicateLabels *LabelMapForMetrics                   `yaml:"duplicate_labels"`
	Split           *SplitMetric                          `yaml:"split"`
	Merge           *MergeMetric                          `yaml:"merge"`
	ToDelta         []types.MetricName                    `yaml:"to_delta"`
}

type LabelMapForMetrics struct {
	ApplyToMetrics []types.MetricName `yaml:"apply_to_metrics"`
	LabelMap       map[string]string  `yaml:"label_map"`
}

type SplitMetric struct {
	ApplyToMetric   types.MetricName                      `yaml:"apply_to_metric"`
	ByLabel         string                                `yaml:"by_label"`
	LabelsToMetrics map[types.LabelValue]types.MetricName `yaml:"labels_to_metrics"`
}

type MergeMetric struct {
	CreateMetric     types.MetricName                      `yaml:"create_metric"`
	ByLabel          string                                `yaml:"by_label"`
	LabelsForMetrics map[types.LabelValue]types.MetricName `yaml:"labels_for_metrics"`
}

type MetricSchema struct {
	MetricNames []string `yaml:"metric_names"`
	Unit        string
	ValueType   string `yaml:"value_type"`
	Temporality string
	Monotonic   bool
	Labels      map[string]LabelSchema
}

type LabelSchema struct {
	Values      []string
	Description string
	Required    string
	Example     string
}
