package boogie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func Template(r io.Reader) (Pipeline, error) {
	s := bufio.NewScanner(r)
	templates := make(map[string]string)
	parsing := false
	pc := 0
	var p Pipeline
	var buff string
	for s.Scan() {
		pc++
		line := s.Text()
		if len(line) > 0 {
			if strings.TrimSpace(line)[0] == '{' {
				parsing = true
			}
		}
		if !parsing {
			command := strings.Split(line, " ")
			if len(command) == 3 {
				if command[0] == "template" {
					b, err := ioutil.ReadFile(command[2])
					if err != nil {
						return p, err
					}
					templates[command[1]] = string(b)
				} else {
					return p, fmt.Errorf("unrecognised templating command '%s' on line %d", command[0], pc)
				}
			}
		} else {
			for k, v := range templates {
				template := fmt.Sprintf("%%%s", k)
				if strings.Contains(line, template) {
					line = strings.Replace(line, template, v, -1)
				}
			}
			buff += fmt.Sprintln(line)
		}
	}

	err := json.Unmarshal([]byte(buff), &p)
	return p, err
}
