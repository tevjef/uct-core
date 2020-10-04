package trace

import (
	"log"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
)

func init() {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
}
