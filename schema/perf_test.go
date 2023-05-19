package schema

import (
	"log"
	"runtime"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/tigrannajaryan/telemetry-schema/schema/converter"
	"github.com/tigrannajaryan/telemetry-schema/schema/otlp"
)

const spansPerBatch = 100
const metricsPerBatch = spansPerBatch
const logsPerBatch = spansPerBatch

const attrsPerSpans = 10
const eventsPerSpan = 3
const attrsPerLog = attrsPerSpans

var batchTypes = []struct {
	name     string
	batchGen func(gen *otlp.Generator) []otlp.ExportRequest
}{
	//{name: "Logs", batchGen: generateLogBatches},
	{name: "Trace/Attribs", batchGen: generateAttrBatches},
	//{name: "Trace/Events", batchGen: generateTimedEventBatches},
	{name: "Metric/Int64", batchGen: generateMetricInt64Batches},
}

const BatchCount = 1

func BenchmarkGenerate(b *testing.B) {
	b.SkipNow()

	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				gen := otlp.NewGenerator()
				for i := 0; i < b.N; i++ {
					batches := batchType.batchGen(gen)
					if batches == nil {
						// Unsupported test type and batch type combination.
						b.SkipNow()
						return
					}
				}
			},
		)
	}
}

func BenchmarkEncode(b *testing.B) {

	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				b.StopTimer()
				gen := otlp.NewGenerator()
				batches := batchType.batchGen(gen)
				if batches == nil {
					// Unsupported test type and batch type combination.
					b.SkipNow()
					return
				}

				runtime.GC()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					for _, batch := range batches {
						encode(batch)
					}
				}
			},
		)
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				batches := batchType.batchGen(otlp.NewGenerator())
				if batches == nil {
					// Unsupported test type and batch type combination.
					b.SkipNow()
					return
				}

				var encodedBytes [][]byte
				for _, batch := range batches {
					encodedBytes = append(encodedBytes, encode(batch))
				}

				runtime.GC()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j, bytes := range encodedBytes {
						decode(bytes, batches[j].(proto.Message))
					}
				}
			},
		)
	}
}

func BenchmarkDecodeAndConvertSchema(b *testing.B) {
	ast, err := Parse("testdata/schema-example.yaml")
	require.NoError(b, err)

	schema := Compile(ast)

	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				batches := batchType.batchGen(otlp.NewGenerator())
				if batches == nil {
					// Unsupported test type and batch type combination.
					b.SkipNow()
					return
				}

				var encodedBytes [][]byte
				for _, batch := range batches {
					encodedBytes = append(encodedBytes, encode(batch))
				}

				runtime.GC()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j, bytes := range encodedBytes {
						msg := batches[j].(proto.Message)
						decode(bytes, msg)
						converter.ConvertRequest(msg.(otlp.ExportRequest), schema)
					}
				}
			},
		)
	}
}

func generateAttrBatches(gen *otlp.Generator) []otlp.ExportRequest {
	var batches []otlp.ExportRequest
	for i := 0; i < BatchCount; i++ {
		batches = append(batches, gen.GenerateSpanBatch(spansPerBatch, attrsPerSpans, 0))
	}
	return batches
}

func generateTimedEventBatches(gen *otlp.Generator) []otlp.ExportRequest {
	var batches []otlp.ExportRequest
	for i := 0; i < BatchCount; i++ {
		batches = append(batches, gen.GenerateSpanBatch(spansPerBatch, 3, eventsPerSpan))
	}
	return batches
}

func generateLogBatches(gen *otlp.Generator) []otlp.ExportRequest {
	var batches []otlp.ExportRequest
	for i := 0; i < BatchCount; i++ {
		batch := gen.GenerateLogBatch(logsPerBatch, attrsPerLog)
		if batch == nil {
			return nil
		}
		batches = append(batches, batch)
	}
	return batches
}

func generateMetricInt64Batches(gen *otlp.Generator) []otlp.ExportRequest {
	var batches []otlp.ExportRequest
	for i := 0; i < BatchCount; i++ {
		batch := gen.GenerateMetricBatch(metricsPerBatch, 1, true)
		if batch == nil {
			return nil
		}
		batches = append(batches, batch)
	}
	return batches
}

func encode(request otlp.ExportRequest) []byte {
	bytes, err := proto.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func decode(bytes []byte, pb proto.Message) {
	err := proto.Unmarshal(bytes, pb)
	if err != nil {
		log.Fatal(err)
	}
}
