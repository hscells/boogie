// Boogie is a domain specific language (DSL) around groove.
// For more information, see https://github.com/hscells/groove.
package main

import (
	"bytes"
	"encoding/json"
	"github.com/alexflint/go-arg"
	"github.com/hscells/boogie"
	"github.com/hscells/groove"
	"github.com/hscells/transmute/backend"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	dsl boogie.Pipeline
)

type args struct {
	Queries  string `arg:"help:Path to queries.,required"`
	Pipeline string `arg:"help:Path to boogie pipeline.,required"`
	LogFile  string `arg:"help:File to output logs to."`
}

func (args) Version() string {
	return "boogie 19.Apr.2018"
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
	g, err := boogie.CreatePipeline(dsl)
	if err != nil {
		log.Fatal(err)
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
			if len(result.Topic) > 0 {
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
}
