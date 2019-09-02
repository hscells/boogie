// Boogie is a domain specific language (DSL) around groove.
// For more information, see https://github.com/hscells/groove.
package main

import (
	"bytes"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/go-errors/errors"
	"github.com/hscells/boogie"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/pipeline"
	"github.com/hscells/transmute/backend"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type args struct {
	Pipeline     string   `arg:"help:Path to boogie pipeline.,required"`
	LogFile      string   `arg:"help:File to output logs to."`
	TemplateArgs []string `arg:"help:Additional arguments to pass to template file.,positional"`
}

func (args) Version() string {
	return "boogie 16.Jan.2019"
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
			panic(err)
		}
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	}

	// Read the contents of the dsl file.
	b, err := ioutil.ReadFile(args.Pipeline)
	if err != nil {
		panic(err)
	}

	// Parse the dsl file into a struct.
	dsl, err := boogie.Template(bytes.NewBuffer(b), args.TemplateArgs...)
	if err != nil {
		panic(err)
	}

	// Create the main pipeline.
	g, err := boogie.CreatePipeline(dsl)
	if err != nil {
		panic(err)
	}

	eval.RelevanceGrade = dsl.Output.Evaluations.RelevanceGrade
	// Execute the groove pipeline. This is done in a go routine, and the results are sent back through the channel.
	pipelineChannel := make(chan pipeline.Result)
	go g.Execute(pipelineChannel)

	evaluations := make([]string, len(dsl.Evaluations))

	var trecEvalFile *os.File
	if len(dsl.Output.Trec.Output) > 0 {
		trecEvalFile, err = os.OpenFile(dsl.Output.Trec.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		//trecEvalFile.Truncate(0)
		//trecEvalFile.Seek(0, 0)
		defer trecEvalFile.Close()
	}

	for result := range pipelineChannel {
		switch result.Type {
		case pipeline.Measurement:
			// Process the measurement outputs.
			for i, formatter := range dsl.Output.Measurements {
				err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(result.Measurements[i]).Bytes(), 0644)
				if err != nil {
					panic(err)
				}
			}
		case pipeline.Transformation:
			// Output the transformed queries
			if len(dsl.Transformations.Output) > 0 {
				s, err := backend.NewCQRQuery(result.Transformation.Transformation).StringPretty()
				if err != nil {
					panic(err)
				}
				q := bytes.NewBufferString(s).Bytes()
				err = ioutil.WriteFile(filepath.Join(g.Transformations.Output, result.Transformation.Name), q, 0644)
				if err != nil {
					panic(err)
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
				trecEvalFile.Write([]byte(strings.Join(l, "\n") + "\n"))
				result.TrecResults = nil
			}
		case pipeline.Error:
			if len(result.Topic) > 0 {
				log.Printf("an error occurred in topic %v", result.Topic)
			} else {
				log.Println("an error occurred")
			}
			fmt.Println(errors.Wrap(result.Error, 1).ErrorStack())
			panic(result.Error)
			return
		}
	}

	// Process the evaluation outputs.
	for i, formatter := range dsl.Output.Evaluations.Measurements {
		err := ioutil.WriteFile(formatter.Filename, bytes.NewBufferString(evaluations[i]).Bytes(), 0644)
		if err != nil {
			panic(err)
		}
	}
}
