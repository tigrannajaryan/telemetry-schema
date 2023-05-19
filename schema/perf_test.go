package schema

import (
	"log"
	"strconv"
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
	batchGen func(gen *otlp.Generator) otlp.ExportRequest
}{
	//{name: "Logs", batchGen: generateLogBatches},
	{name: "Trace/Attribs", batchGen: generateAttrBatches},
	//{name: "Trace/Events", batchGen: generateTimedEventBatches},
	{name: "Metric/Int64", batchGen: generateMetricInt64Batches},
}

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
				gen := otlp.NewGenerator()
				batch := batchType.batchGen(gen)
				if batch == nil {
					// Unsupported test type and batch type combination.
					b.SkipNow()
					return
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					encode(batch)
				}
			},
		)
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				batch := batchType.batchGen(otlp.NewGenerator())
				if batch == nil {
					// Unsupported test type and batch type combination.
					b.SkipNow()
					return
				}

				encodedBytes := encode(batch)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					decode(encodedBytes, batch.(proto.Message))
				}
			},
		)
	}
}

func BenchmarkConvertSchema(b *testing.B) {
	ast, err := Parse("testdata/schema-example.yaml")
	require.NoError(b, err)

	schema := Compile(ast)

	for _, batchType := range batchTypes {
		b.Run(
			batchType.name, func(b *testing.B) {
				var msgs []proto.Message
				for i := 0; i < b.N; i++ {
					msg := batchType.batchGen(otlp.NewGenerator())
					if msg == nil {
						// Unsupported test type and batch type combination.
						b.SkipNow()
						return
					}
					msgs = append(msgs, msg)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					converter.ConvertRequest(msgs[i].(otlp.ExportRequest), schema)
				}
			},
		)
	}
}

func generateAttrBatches(gen *otlp.Generator) otlp.ExportRequest {
	return gen.GenerateSpanBatch(spansPerBatch, attrsPerSpans, 0)
}

func generateTimedEventBatches(gen *otlp.Generator) otlp.ExportRequest {
	return gen.GenerateSpanBatch(spansPerBatch, 3, eventsPerSpan)
}

func generateLogBatches(gen *otlp.Generator) otlp.ExportRequest {
	return gen.GenerateLogBatch(logsPerBatch, attrsPerLog)
}

func generateMetricInt64Batches(gen *otlp.Generator) otlp.ExportRequest {
	return gen.GenerateMetricBatch(metricsPerBatch, 1, true)
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

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			m := map[string]string{}
			m[strconv.Itoa(j)] = "def"
			l := 0
			for k, v := range m {
				l++
				k = v
				k = k
			}
			l = l
		}
	}
}
