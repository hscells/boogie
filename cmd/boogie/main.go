// Boogie is a domain specific language (DSL) around groove.
// For more information, see https://github.com/hscells/groove.
package main

import (
	"bytes"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/hscells/boogie"
	"github.com/hscells/groove"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/pipeline"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type args struct {
	Pipeline     string   `arg:"help:Path to boogie pipeline.,required"`
	LogFile      string   `arg:"help:File to output logs to."`
	TemplateArgs []string `arg:"help:Additional arguments to pass to template file.,positional"`
}

func (args) Version() string {
	return fmt.Sprintf("boogie 2.Feb.2022 v1 using groove %s", groove.Version)
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
	err = boogie.Execute(dsl, pipelineChannel)
	if err != nil {
		panic(err)
	}
}
