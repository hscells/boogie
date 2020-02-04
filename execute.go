package boogie

import (
	"bytes"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/pipeline"
	"github.com/hscells/transmute"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func Execute(dsl Pipeline, pipelineChannel chan pipeline.Result) error {
	// Handle the case if the method is not run as a command.s
	if measurementMapping == nil || len(measurementMapping) == 0 {
		err := RegisterSources(dsl)
		if err != nil {
			return err
		}
	}

	// File that will contain TREC run data.
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
		case pipeline.Evaluation:
			evaluations[result.Topic] = result.Evaluations
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
		case pipeline.Formulation:
			for _, s := range result.Formulation.Sup {
				// Create the folder the data will be contained in.
				err := os.MkdirAll(path.Join(dsl.Formulation.Method, s.Name), 0777)
				if err != nil {
					return err
				}

				for _, d := range s.Data {
					log.Printf("writing supplimentary file %s\n", path.Join(dsl.Formulation.Method, s.Name, d.Name))
					// Create and open the file that will contain the data.
					f, err := os.OpenFile(path.Join(dsl.Formulation.Method, s.Name, d.Name), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
					if err != nil {
						return err
					}
					// Marshal the data into bytes for writing to disk.
					b, err := d.Value.Marshal()
					if err != nil {
						return err
					}
					// Write those bytes to disk.
					_, err = f.Write(b)
					if err != nil {
						return err
					}
					// Close that file.
					err = f.Close()
					if err != nil {
						return err
					}
				}
			}

			// Create the folder that will contain the formulated query/queries.
			err := os.MkdirAll(dsl.Formulation.Method, 0777)
			if err != nil {
				return err
			}
			for i, q := range result.Formulation.Queries {
				log.Println(q)
				err := os.MkdirAll(path.Join(dsl.Formulation.Method, strconv.Itoa(i)), 0777)
				if err != nil {
					return err
				}
				// Compile the query to CQR.
				s, err := transmute.CompileCqr2PubMed(q)
				if err != nil {
					return err
				}
				// Open the file that will contain the query.
				f, err := os.OpenFile(path.Join(dsl.Formulation.Method, strconv.Itoa(i), result.Topic), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
				if err != nil {
					return err
				}
				// Write the query to disk.
				_, err = f.WriteString(s)
				if err != nil {
					return err
				}
				// Close the file.
				err = f.Close()
				if err != nil {
					return err
				}
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
		topics := make([]string, len(measurements))
		i := 0
		for topic := range measurements {
			topics[i] = topic
			i++
		}
		i = 0
		headers := make([]string, len(dsl.Measurements))
		data := make([][]float64, len(dsl.Measurements))
		for i, measure := range dsl.Measurements {
			headers[i] = measurementMapping[measure].Name()
		}
		for i, header := range headers {
			data[i] = make([]float64, len(topics))
			for j, topic := range topics {
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
