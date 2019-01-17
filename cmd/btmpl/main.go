package main

import (
	"bytes"
	"encoding/json"
	"github.com/alexflint/go-arg"
	"github.com/hscells/boogie"
	"io"
	"io/ioutil"
	"os"
)

type args struct {
	Pipeline     string   `arg:"help:Path to boogie pipeline."`
	TemplateArgs []string `arg:"help:Additional arguments to pass to template file.,positional"`
}

func (args) Version() string {
	return "btmpl 18.Jan.2019"
}

func (args) Description() string {
	return `Template boogie pipeline files.`
}

func main() {
	// Parse the command line arguments.
	var args args
	arg.MustParse(&args)

	var input io.Reader
	if len(args.Pipeline) == 0 {
		input = os.Stdin
	} else {
		b, err := ioutil.ReadFile(args.Pipeline)
		if err != nil {
			panic(err)
		}
		input = bytes.NewBuffer(b)
	}

	p, err := boogie.Template(input, args.TemplateArgs...)
	if err != nil {
		panic(err)
	}
	r, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.Write(r)
	if err != nil {
		panic(err)
	}
}
