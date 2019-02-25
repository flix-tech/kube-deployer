package main

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

func UnmarshalYaml(yamlFile string) ([]map[string]interface{}, error) {
	var objects = make([]map[string]interface{}, 0)

	var rp = regexp.MustCompile(`(?m:^---$)`)
	templates := rp.Split(yamlFile, -1)

	templates = Filter(templates, func(v string) bool {
		return strings.Contains(v, "kind")
	})

	var err error

	for _, splitTemplate := range templates {
		var object map[string]interface{}

		err = yaml.Unmarshal([]byte(splitTemplate), &object)

		if err != nil {
			return nil, err
		}

		objects = append(objects, object)
	}

	return objects, nil
}
