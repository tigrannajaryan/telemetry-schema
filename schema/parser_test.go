package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSchema(t *testing.T) {
	ts, err := Parse("testdata/schema-example.yaml")
	require.NoError(t, err)
	require.NotNil(t, ts)
}
