package boogie

import (
	"bytes"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/pipeline"
	"github.com/hscells/transmute"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Execute(dsl Pipeline, pipelineChannel chan pipeline.Result) error {
	var trecEvalFile *os.File
	if len(dsl.Output.Trec.Output) > 0 {
		var err error
		trecEvalFile, err = os.OpenFile(dsl.Output.Trec.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}

	measurements := make(map[string]map[string]float64)
	evaluations := make(map[string]map[string]float64)

	defer trecEvalFile.Close()
	for result := range pipelineChannel {
		switch result.Type {
		case pipeline.Measurement:
			measurements[result.Topic] = result.Measurements
			// Process the measurement outputs.

		case pipeline.Transformation:
			// Output the transformed queries
			if len(dsl.Transformations.Output) > 0 {
				s, err := transmute.CompileCqr2PubMed(result.Transformation.Transformation)
				if err != nil {
					return err
				}
				q := bytes.NewBufferString(s).Bytes()
				err = ioutil.WriteFile(filepath.Join(dsl.Transformations.Output, result.Transformation.Name), q, 0644)
				if err != nil {
					return err
				}
			}
		case pipeline.Evaluation:
			evaluations[result.Topic] = result.Evaluations
		case pipeline.TrecResult:
			if result.TrecResults != nil && len(*result.TrecResults) > 0 {
				l := make([]string, len(*result.TrecResults))
				for i, r := range *result.TrecResults {
					l[i] = r.String()
				}
				_, err := trecEvalFile.Write([]byte(strings.Join(l, "\n") + "\n"))
				if err != nil {
					return err
				}
				result.TrecResults = nil
			}
		case pipeline.Error:
			if len(result.Topic) > 0 {
				log.Printf("an error occurred in topic %v", result.Topic)
			} else {
				log.Println("an error occurred")
			}
			return result.Error
		}
	}

	topics := make([]string, len(evaluations))
	i := 0
	for topic := range evaluations {
		topics[i] = topic
		i++
	}

	if len(evaluations) > 0 {
		for _, formatter := range dsl.Output.Evaluations.Measurements {
			var f output.EvaluationFormatter
			switch formatter.Format {
			case "json":
				f = output.JsonEvaluationFormatter
			default:
				log.Println("unexpected evaluation formatter, using json")
				f = output.JsonEvaluationFormatter
			}

			formatted, err := f(evaluations)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(formatted).Bytes(), 0644)
			if err != nil {
				return err
			}
		}
	}

	if len(measurements) > 0 {
		headers := make([]string, len(dsl.Measurements))
		data := make([][]float64, len(dsl.Measurements))
		for i, measure := range dsl.Measurements {
			headers[i] = measurementMapping[measure].Name()
		}

		for i, topic := range topics {
			data[i] = make([]float64, len(headers))
			for j, header := range headers {
				data[i][j] = measurements[topic][header]
			}
		}
		for _, formatter := range dsl.Output.Measurements {
			var (
				r   string
				err error
			)
			switch formatter.Format {
			case "csv":
				r, err = output.CsvMeasurementFormatter(topics, headers, data)
			case "json":
				r, err = output.JsonMeasurementFormatter(topics, headers, data)
			}
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(r).Bytes(), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
