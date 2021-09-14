package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/tigrannajaryan/telemetry-schema/schema"
	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "i", "", "schema file to check")
	flag.Parse()
	if inputFile == "" {
		fmt.Print("Must specify a schema file to check.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ts, err := schema.Parse(inputFile)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	if ts.FileFormat != "1.0.0" {
		fmt.Printf("file_format must be 1.0.0.\n")
		os.Exit(1)
	}

	if ts == nil {
		panic("Schema is empty")
	}

	schemaVerInFileName := path.Base(inputFile)
	fmt.Printf("Schema version according to file name is %s.\n", schemaVerInFileName)

	if _, ok := ts.Versions[types.TelemetryVersion(schemaVerInFileName)]; !ok {
		fmt.Printf("Schema version %s is not found in the file.\n", schemaVerInFileName)
		os.Exit(1)
	}

	if ts.SchemaURL == "" {
		fmt.Printf("schema_url is missing.\n")
		os.Exit(1)
	}

	surl, err := url.Parse(ts.SchemaURL)
	if err != nil {
		fmt.Printf("schema_url cannot be parsed: %v.\n", err)
		os.Exit(1)
	}

	paths := strings.Split(surl.Path, "/")
	if len(paths) == 0 {
		fmt.Printf("schema_url path should not be empty.\n")
		os.Exit(1)
	}

	schemaVerInPath := paths[len(paths)-1]
	if schemaVerInPath != schemaVerInFileName {
		fmt.Printf("The last part of schema_url path is %s but expected %s\n", schemaVerInPath, schemaVerInFileName)
		os.Exit(5)
	}

	fmt.Printf("%s schema file checks are successful.\n", inputFile)
}
