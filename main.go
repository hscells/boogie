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
	"github.com/hscells/groove/eval"
	"github.com/TimothyJones/trecresults"
	"github.com/hscells/groove"
	"os"
	"strings"
	"github.com/hscells/groove/rewrite"
	"io"
)

var (
	dsl Pipeline
)

type args struct {
	Queries  string `arg:"help:Path to queries.,required"`
	Pipeline string `arg:"help:Path to boogie pipeline.,required"`
	LogFile  string `arg:"help:File to output logs to."`
}

func (args) Version() string {
	return "boogie 17.Jan.2018"
}

func (args) Description() string {
	return `DSL front-end for groove.
For further documentation see https://godoc.org/github.com/hscells/boogie.
To view the source or to contribute see https://github.com/hscells/boogie.

For information about groove, see https://github.com/hscells/groove.`
}

// CreateQueryChainSVM creates a sub-pipeline which can train an SVM for query chains.
func CreateQueryChainSVM(dsl Pipeline, g pipeline.GroovePipeline) (qc *pipeline.QueryChainSVM) {
	var ok bool
	if qc.Selector, ok = g.QueryChain.CandidateSelector.(rewrite.OracleQueryChainCandidateSelector); !ok {
		log.Fatalf("oracle must be used to extract features for svm, got %v", dsl.Rewrite.Chain)
	}

	qc.Transformations = g.QueryChain.Transformations
	qc.Features = dsl.Rewrite.SVM.Features
	qc.Model = dsl.Rewrite.SVM.Model
	qc.ShouldTrain = dsl.Rewrite.SVM.ShouldTrain
	qc.ShouldExtract = dsl.Rewrite.SVM.ShouldExtract

	if qc.ShouldExtract && len(qc.Features) == 0 {
		log.Fatal("a feature file must be specified when extracting features for a query chain SVM")
	}

	if qc.ShouldTrain && len(qc.Model) == 0 {
		log.Fatal("a model file must be specified when training an SVM for query chains")
	}

	return
}

// CreatePipeline creates the main groove pipeline.
func CreatePipeline(dsl Pipeline) pipeline.GroovePipeline {
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

	if len(dsl.Output.Evaluations.Qrels) > 0 {
		b, err := ioutil.ReadFile(dsl.Output.Evaluations.Qrels)
		if err != nil {
			log.Fatalln(err)
		}
		qrels, err := trecresults.QrelsFromReader(bytes.NewReader(b))
		if err != nil {
			log.Fatalln(err)
		}
		g.EvaluationQrels = qrels
	}

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
			log.Fatalf("%v is not a known preprocessing transformation", t)
		}
	}

	//g.QueryChain
	if len(dsl.Rewrite.Chain) > 0 && len(dsl.Rewrite.Transformations) > 0 {
		var transformations []rewrite.Transformation
		for _, transformation := range dsl.Rewrite.Transformations {
			if t, ok := rewriteTransformationMapping[transformation]; ok {
				transformations = append(transformations, t)
			} else {
				log.Fatalf("%v is not a known rewrite transformation", transformation)
			}
		}

		if qc, ok := queryChainCandidateSelectorMapping[dsl.Rewrite.Chain]; ok {
			g.QueryChain = rewrite.NewQueryChain(qc, transformations...)
		} else {
			log.Fatalf("%v is not a known query chain candidate selector", dsl.Rewrite.Chain)
		}

	}

	g.Transformations.Output = dsl.Transformations.Output
	g.OutputTrec.Path = dsl.Output.Trec.Output
	return g
}

func main() {
	// Parse the command line arguments.
	var args args
	arg.MustParse(&args)

	if len(args.LogFile) > 0 {
		f, err := os.OpenFile(args.LogFile, os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		f.Truncate(0)
		if err != nil {
			log.Fatal(err)
		}
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	}

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

	// Create the main pipeline.
	g := CreatePipeline(dsl)
	var qc *pipeline.QueryChainSVM
	if dsl.Rewrite.SVM != (PipelineQueryChainSVM{}) {
		qc = CreateQueryChainSVM(dsl, g)
	}

	// Execute the groove pipeline. This is done in a go routine, and the results are sent back through the channel.
	pipelineChannel := make(chan groove.PipelineResult)
	go g.Execute(args.Queries, pipelineChannel)

	evaluations := make([]string, len(dsl.Evaluations))

	var trecEvalFile *os.File
	if len(dsl.Output.Trec.Output) > 0 {
		trecEvalFile, err = os.OpenFile(dsl.Output.Trec.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		trecEvalFile.Truncate(0)
		trecEvalFile.Seek(0, 0)
		defer trecEvalFile.Close()
	}

	for {
		result := <-pipelineChannel
		if result.Type == groove.Done {
			break
		}
		switch result.Type {
		case groove.Measurement:
			// Process the measurement outputs.
			for i, formatter := range dsl.Output.Measurements {
				err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Measurements[i]).Bytes(), 0644)
				if err != nil {
					log.Fatalln(err)
				}
			}
		case groove.Transformation:
			// Output the transformed queries
			if len(dsl.Transformations.Output) > 0 {
				if qc != nil && (qc.ShouldExtract || qc.ShouldTrain) && qc.Transformations != nil {
					qc.AppendQuery(result.Transformation.ToGroovePipelineQuery())
				}

				s, err := backend.NewCQRQuery(result.Transformation.Transformation).StringPretty()
				if err != nil {
					log.Fatalln(err)
				}
				q := bytes.NewBufferString(s).Bytes()
				err = ioutil.WriteFile(filepath.Join(g.Transformations.Output, result.Transformation.Name), q, 0644)
				if err != nil {
					log.Fatalln(err)
				}
			}
		case groove.Evaluation:
			for i, e := range result.Evaluations {
				evaluations[i] = e
			}
		case groove.TrecResult:
			if result.TrecResults != nil && len(*result.TrecResults) > 0 {
				l := make([]string, len(*result.TrecResults))
				for i, r := range *result.TrecResults {
					l[i] = r.String()
				}
				trecEvalFile.Write([]byte(strings.Join(l, "\n") + "\n"))
				result.TrecResults = nil
			}
		case groove.Error:
			if result.Topic > 0 {
				log.Printf("an error occurred in topic %v", result.Topic)
			} else {
				log.Println("an error occurred")
			}
			log.Fatalln(result.Error)
			return
		}
	}

	// Process the evaluation outputs.
	for i, formatter := range dsl.Output.Evaluations.Measurements {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(evaluations[i]).Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// ------------Run any other sub-pipelines------------
	// Extracting features.
	if qc != nil && qc.ShouldExtract && len(qc.Queries) > 0 {
		f, err := os.OpenFile(qc.Features, os.O_WRONLY|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		_, err = qc.WriteFeatures(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Training a model.
	if qc != nil && qc.ShouldTrain && len(qc.Queries) > 0 {
		err := qc.TrainModel()
		if err != nil {
			log.Fatal(err)
		}
	}

}
