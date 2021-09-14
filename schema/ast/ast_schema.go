package ast

import "github.com/tigrannajaryan/telemetry-schema/schema/types"

type Schema struct {
	FileFormat string `yaml:"file_format"`
	SchemaURL  string `yaml:"schema_url"`
	Versions   map[types.TelemetryVersion]VersionDef
}

type VersionDef struct {
	All        VersionOfAttributes
	Resources  VersionOfAttributes
	Spans      VersionOfSpans
	SpanEvents VersionOfSpanEvents `yaml:"span_events"`
	Logs       VersionOfLogs
	Metrics    VersionOfMetrics
}

type VersionOfAttributes struct {
	Changes []AttributeTranslationAction
}

type AttributeTranslationAction struct {
	RenameAttributes *MappingOfAttributes `yaml:"rename_attributes"`
}

type MappingOfAttributes map[string]string
