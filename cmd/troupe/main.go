package main

import (
	"encoding/json"
	"github.com/alexflint/go-arg"
	"io/ioutil"
	"os"
)

type args struct {
	Pipeline string `arg:"help:path to troupe pipeline,required,positional"`
}

func (args) Version() string {
	return "troupe 08.Oct.2018"
}

func (args) Description() string {
	return `manager for boogie on multiple server`
}

func main() {
	// Parse the command line arguments.
	var (
		args args
		dsl  ClientConfig
	)
	arg.MustParse(&args)

	f, err := os.OpenFile(args.Pipeline, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &dsl)
	if err != nil {
		panic(err)
	}

	c := NewClient()
	for _, job := range dsl.Jobs {
		err = c.AddJob(job.SSHUsername, job.SSHAddress, job.Pipeline, job.Logger, Interactive)
		if err != nil {
			panic(err)
		}
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
}
