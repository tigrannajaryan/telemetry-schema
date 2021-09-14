package ast

import "github.com/tigrannajaryan/telemetry-schema/schema/types"

type VersionOfSpans struct {
	Changes []SpanTranslationAction
}

type VersionOfSpanEvents struct {
	Changes []SpanEventTranslationAction
}

type SpanTranslationAction struct {
	RenameAttributes *RenameSpanAttributes `yaml:"rename_attributes"`
}

type SpanEventTranslationAction struct {
	RenameEvents     *RenameSpanEvents          `yaml:"rename_events"`
	RenameAttributes *RenameSpanEventAttributes `yaml:"rename_attributes"`
}

type RenameSpanAttributes struct {
	AttributeMap map[string]string `yaml:"attribute_map"`
}

type RenameSpanEvents struct {
	EventNameMap map[string]string `yaml:"name_map"`
}

type RenameSpanEventAttributes struct {
	ApplyToSpans  []types.SpanName  `yaml:"apply_to_spans"`
	ApplyToEvents []types.EventName `yaml:"apply_to_events"`
	AttributeMap  map[string]string `yaml:"attribute_map"`
}
