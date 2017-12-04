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
	"github.com/hscells/groove/eval"
	"github.com/TimothyJones/trecresults"
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

	if len(dsl.Measurements) > 0 && len(dsl.Output.Measurements) == 0 {
		log.Fatal("At least one output format must be supplied when using analysis measurements")
	}

	if len(dsl.Output.Measurements) > 0 && len(dsl.Measurements) == 0 {
		log.Fatal("At least one analysis measurement must be supplied for the output formats")
	}

	if len(dsl.Evaluations) > 0 && len(dsl.Output.Evaluations.Measurements) == 0 {
		log.Fatal("At least one output format must be supplied when using evaluation measurements")
	}

	if len(dsl.Output.Evaluations.Measurements) > 0 && len(dsl.Evaluations) == 0 {
		log.Fatal("At least one evaluation measurement must be supplied for the output formats")
	}

	g.Measurements = []analysis.Measurement{}
	for _, measurementName := range dsl.Measurements {
		if m, ok := measurementMapping[measurementName]; ok {
			g.Measurements = append(g.Measurements, m)
		} else {
			log.Fatalf("%v is not a known measurement", measurementName)
		}
	}

	g.Evaluations = []eval.Evaluator{}
	for _, evaluationMeasurement := range dsl.Evaluations {
		if m, ok := evaluationMapping[evaluationMeasurement.Evaluation]; ok {
			g.Evaluations = append(g.Evaluations, m)
		} else {
			log.Fatalf("%v is not a known evaluation measurement", evaluationMeasurement.Evaluation)
		}
	}

	b, err = ioutil.ReadFile(dsl.Output.Evaluations.Qrels)
	if err != nil {
		log.Fatalln(err)
	}
	qrels, err := trecresults.QrelsFromReader(bytes.NewReader(b))
	if err != nil {
		log.Fatalln(err)
	}
	g.EvaluationQrels = qrels

	g.MeasurementFormatters = []output.MeasurementFormatter{}
	for _, formatter := range dsl.Output.Measurements {
		if o, ok := measurementFormatters[formatter.Format]; ok {
			g.MeasurementFormatters = append(g.MeasurementFormatters, o)
		} else {
			log.Fatalf("%v is not a known measurement output format", formatter.Format)
		}
	}

	g.EvaluationFormatters = []output.EvaluationFormatter{}
	for _, formatter := range dsl.Output.Evaluations.Measurements {
		if o, ok := evaluationFormatters[formatter.Format]; ok {
			g.EvaluationFormatters = append(g.EvaluationFormatters, o)
		} else {
			log.Fatalf("%v is not a known evaluation output format", formatter.Format)
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

	g.OutputTrec.Path = dsl.Output.Trec.Output

	// Execute the groove pipeline.
	result, err := g.Execute(args.Queries)
	if err != nil {
		log.Fatal(err)
	}

	// Process the measurement outputs.
	for i, formatter := range dsl.Output.Measurements {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Measurements[i]).Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Process the evaluation outputs.
	for i, formatter := range dsl.Output.Evaluations.Measurements {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Evaluations[i]).Bytes(), 0644)
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
