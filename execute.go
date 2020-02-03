package boogie

import (
	"bytes"
	"github.com/hscells/groove/pipeline"
	"github.com/hscells/transmute"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Execute(dsl Pipeline, pipelineChannel chan pipeline.Result) error {
	evaluations := make([]string, len(dsl.Evaluations))
	var trecEvalFile *os.File
	if len(dsl.Output.Trec.Output) > 0 {
		var err error
		trecEvalFile, err = os.OpenFile(dsl.Output.Trec.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer trecEvalFile.Close()
	for result := range pipelineChannel {
		switch result.Type {
		case pipeline.Measurement:
			// Process the measurement outputs.
			for i, formatter := range dsl.Output.Measurements {
				err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Measurements[i]).Bytes(), 0644)
				if err != nil {
					return err
				}
			}
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
			for i, e := range result.Evaluations {
				evaluations[i] = e
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
		case pipeline.Error:
			if len(result.Topic) > 0 {
				log.Printf("an error occurred in topic %v", result.Topic)
			} else {
				log.Println("an error occurred")
			}
			return result.Error
		}
	}

	// Process the evaluation outputs.
	for i, formatter := range dsl.Output.Evaluations.Measurements {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(evaluations[i]).Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
