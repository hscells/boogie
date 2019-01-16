package main

import (
	"encoding/json"
	"github.com/hscells/boogie"
	"os"
)

func main() {
	input := os.Stdin
	p, err := boogie.Template(input)
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
