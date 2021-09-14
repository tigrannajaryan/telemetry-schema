package schema

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/tigrannajaryan/telemetry-schema/schema/ast"
)

func Parse(schemaFile string) (*ast.Schema, error) {
	var ts ast.Schema
	schemaContent, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(schemaContent, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}
