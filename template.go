package boogie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func Template(r io.Reader, args ...string) (Pipeline, error) {
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
					if len(command[2]) > 0 && command[2][0] == '$' {
						index, err := strconv.Atoi(command[2][1:])
						if err != nil {
							return p, err
						}
						if index >= len(args) {
							return p, fmt.Errorf("index of template argument is higher than the number of arguments, see line %d", pc)
						}
						templates[command[1]] = args[index]
					} else {
						b, err := ioutil.ReadFile(command[2])
						if err != nil {
							return p, err
						}
						templates[command[1]] = string(b)
					}
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
