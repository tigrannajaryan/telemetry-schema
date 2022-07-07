package ast

import "github.com/tigrannajaryan/telemetry-schema/schema/types"

type VersionOfMetrics struct {
	Changes []MetricTranslationAction
	Current []MetricSchema `yaml:"current_metric_schema"`
}

type MetricTranslationAction struct {
	RenameMetrics       map[types.MetricName]types.MetricName `yaml:"rename_metrics"`
	RenameLabels        *AttributeMapForMetrics               `yaml:"rename_attributes"`
	AddAttributes       *AttributeMapForMetrics               `yaml:"add_attributes"`
	DuplicateAttributes *AttributeMapForMetrics               `yaml:"duplicate_attributes"`
	Split               *SplitMetric                          `yaml:"split"`
	Merge               *MergeMetric                          `yaml:"merge"`
	ToDelta             []types.MetricName                    `yaml:"to_delta"`
}

type AttributeMapForMetrics struct {
	ApplyToMetrics []types.MetricName `yaml:"apply_to_metrics"`
	AttributeMap   map[string]string  `yaml:"label_map"`
}

type SplitMetric struct {
	ApplyToMetric       types.MetricName                          `yaml:"apply_to_metric"`
	ByAttribute         types.AttributeName                       `yaml:"by_attribute"`
	AttributesToMetrics map[types.MetricName]types.AttributeValue `yaml:"metrics_from_attributes"`
}

type MergeMetric struct {
	CreateMetric         types.MetricName                          `yaml:"create_metric"`
	ByAttribute          string                                    `yaml:"by_attribute"`
	AttributesForMetrics map[types.MetricName]types.AttributeValue `yaml:"attributes_for_metrics"`
}

type MetricSchema struct {
	MetricNames []string `yaml:"metric_names"`
	Unit        string
	ValueType   string `yaml:"value_type"`
	Temporality string
	Monotonic   bool
	Attributes  map[string]AttributesSchema
}

type AttributesSchema struct {
	Values      []string
	Description string
	Required    string
	Example     string
}
