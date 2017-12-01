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
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/transmute/backend"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var (
	dsl Pipeline
)

type args struct {
	Queries  string `arg:"help:Path to queries.,required"`
	Pipeline string `arg:"help:Path to boogie pipeline.,required"`
}

func (args) Version() string {
	return "boogie 1.Dec.2017"
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

	if len(dsl.Statistic.Source) > 0 {
		if s, ok := statisticSourceMapping[dsl.Statistic.Source]; ok {
			g.StatisticsSource = s
		} else {
			log.Fatalf("%v is not a known statistics source", dsl.Statistic.Source)
		}
	}

	if len(dsl.Statistic.Source) == 0 && len(dsl.Measurements) > 0 {
		log.Fatal("A statistic source is required for measurements")
	}

	if len(dsl.Measurements) > 0 && len(dsl.Output) == 0 {
		log.Fatal("At least one output format must be supplied when using measurements")
	}

	if len(dsl.Output) > 0 && len(dsl.Measurements) == 0 {
		log.Fatal("At least one measurement must be supplied for the output formats")
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

	g.Transformations = preprocess.QueryTransformations{}
	for _, t := range dsl.Transformations.Operations {
		if transformation, ok := transformationMappingBoolean[t]; ok {
			g.Transformations.BooleanTransformations = append(g.Transformations.BooleanTransformations, transformation)
		} else if transformation, ok := transformationMappingElasticsearch[t]; ok {
			g.Transformations.ElasticsearchTransformations = append(g.Transformations.ElasticsearchTransformations, transformation)
		} else {
			log.Fatalf("%v is not a known transformation", t)
		}
	}
	g.Transformations.Output = dsl.Transformations.Output

	g.OutputTrec.Path = dsl.Trec.Output

	// Execute the groove pipeline.
	result, err := g.Execute(args.Queries)
	if err != nil {
		log.Fatal(err)
	}

	// Process the outputs.
	for i, formatter := range dsl.Output {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Measurements[i]).Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Output the transformed queries
	if len(dsl.Transformations.Output) > 0 {
		for _, queryResult := range result.Transformations {
			q := bytes.NewBufferString(backend.NewCQRQuery(queryResult.Transformation).StringPretty()).Bytes()
			err := ioutil.WriteFile(filepath.Join(g.Transformations.Output, queryResult.Name), q, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Output the trec results, if specified.
	if result.TrecResults != nil && len(*result.TrecResults) > 0 {
		l := make([]string, len(*result.TrecResults))
		for i, r := range *result.TrecResults {
			l[i] = r.String()
		}
		ioutil.WriteFile(g.OutputTrec.Path, []byte(strings.Join(l, "\n")), 0644)
	}
}
