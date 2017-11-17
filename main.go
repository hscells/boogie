// Boogie is a domain specific language (DSL) around groove.
// For more information, see https://github.com/hscells/groove.
package main

import (
	"bytes"
	"encoding/json"
	"github.com/alexflint/go-arg"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/pipeline"
	"io/ioutil"
	"log"
	"github.com/hscells/groove/preprocess"
)

var (
	dsl Pipeline
)

type args struct {
	Queries  string `arg:"help:Path to queries.,required"`
	Pipeline string `arg:"help:Path to boogie pipeline.,required"`
}

func (args) Version() string {
	return "boogie 17.Nov.2017"
}

func (args) Description() string {
	return `DSL front-end for groove.
For further documentation see https://godoc.org/github.com/hscells/boogie.
To view the source or to contribute see https://github.com/hscells/boogie.

For information about groove, see https://github.com/hscells/groove.`
}

func main() {
	// Parse the command line arguments.
	var args args
	arg.MustParse(&args)

	// Read the contents of the dsl file.
	b, err := ioutil.ReadFile(args.Pipeline)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the dsl file into a struct.
	err = json.Unmarshal(b, &dsl)
	if err != nil {
		log.Fatal(err)
	}

	// Register the sources used in the groove pipeline.
	RegisterSources()

	// Create a groove pipeline from the boogie dsl.
	g := pipeline.GroovePipeline{}
	if s, ok := querySourceMapping[dsl.Query.Format]; ok {
		g.QueriesSource = s
	} else {
		log.Fatalf("%v is not a known query source", dsl.Query.Format)
	}

	if s, ok := statisticSourceMapping[dsl.Statistic.Source]; ok {
		g.StatisticsSource = s
	} else {
		log.Fatalf("%v is not a known statistics source", dsl.Statistic.Source)
	}

	g.Measurements = []analysis.Measurement{}
	for _, measurementName := range dsl.Measurements {
		if m, ok := measurementMapping[measurementName]; ok {
			g.Measurements = append(g.Measurements, m)
		} else {
			log.Fatalf("%v is not a known measurement", measurementName)
		}
	}

	g.OutputFormats = []output.Formatter{}
	for _, formatter := range dsl.Output {
		if o, ok := outputMapping[formatter.Format]; ok {
			g.OutputFormats = append(g.OutputFormats, o)
		} else {
			log.Fatalf("%v is not a known output format", formatter.Format)
		}
	}

	g.Preprocess = []preprocess.QueryProcessor{}
	for _, p := range dsl.Preprocess {
		if processor, ok := preprocessorMapping[p]; ok {
			g.Preprocess = append(g.Preprocess, processor)
		} else {
			log.Fatalf("%v is not a known preprocessor", p)
		}
	}

	// Execute the groove pipeline.
	outputs, err := g.Execute(args.Queries)
	if err != nil {
		log.Fatal(err)
	}

	// Process the outputs.
	for i, formatter := range dsl.Output {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(outputs[i]).Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
