package boogie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func templateString(r io.Reader, args ...string) (string, error) {
	s := bufio.NewScanner(r)
	templates := make(map[string]string)
	parsing := false
	pc := 0
	var buff string
	for s.Scan() {
		pc++
		line := s.Text()
		if len(line) > 0 {
			x := strings.TrimSpace(line)[0]
			if x == '{' || x == '"' || x == '%' {
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
							return buff, err
						}
						if index >= len(args) {
							return buff, fmt.Errorf("index of template argument is higher than the number of arguments, see line %d", pc)
						}
						templates[command[1]] = args[index]
					} else {
						f, err := os.OpenFile(command[2], os.O_RDONLY, 0644)
						if err != nil {
							return buff, err
						}
						templates[command[1]], err = templateString(f, args...)
						if err != nil {
							return buff, err
						}
						//b, err := ioutil.ReadFile(command[2])
						//if err != nil {
						//	return buff, err
						//}
						//templates[command[1]] = string(b)
						//if err != nil {
						//	return buff, err
						//}
					}
				} else {
					fmt.Println("------------")
					fmt.Println(">>>", line)
					fmt.Println("------------")
					fmt.Println(buff)
					return buff, fmt.Errorf("unrecognised templating command '%s' on line %d", command[0], pc)
				}
			}
		} else {
			for k, v := range templates {
				template := fmt.Sprintf("%%%s", k)
				if strings.Contains(line, template) {
					line = strings.Replace(line, template, v, -1)
				}
				template2 := fmt.Sprintf("@%s", k)
				if strings.Contains(line, template2) {
					line = strings.Replace(line, template2, v, -1)
				}
			}
			buff += fmt.Sprintln(line)
		}
	}
	return buff, nil
}

func Template(r io.Reader, args ...string) (Pipeline, error) {
	var p Pipeline
	t, err := templateString(r, args...)
	if err != nil {
		return Pipeline{}, err
	}
	err = json.Unmarshal([]byte(t), &p)
	return p, err
}
