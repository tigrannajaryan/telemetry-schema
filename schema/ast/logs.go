package ast

type VersionOfLogs struct {
	Changes []LogTranslationAction
}

type LogTranslationAction struct {
	RenameAttributes *RenameLogAttributes `yaml:"rename_attributes"`
}

type RenameLogAttributes struct {
	AttributeMap map[string]string `yaml:"attribute_map"`
}
