package utils

import (
	"os"
	"strconv"
	"strings"
)

func ParseYAML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	root := map[string]interface{}{}
	stack := []map[string]interface{}{root}
	indents := []int{0}

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		spaceCount := len(line) - len(strings.TrimLeft(line, " "))
		line = strings.TrimLeft(line, " ")

		for len(indents) > 1 && spaceCount <= indents[len(indents)-1] {
			stack = stack[:len(stack)-1]
			indents = indents[:len(indents)-1]
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		current := stack[len(stack)-1]

		if value == "" {
			newObj := map[string]interface{}{}
			current[key] = newObj
			stack = append(stack, newObj)
			indents = append(indents, spaceCount+1)
			continue
		}

		current[key] = castValue(value)
	}

	return root, nil
}

func castValue(v string) interface{} {
	if v == "true" {
		return true
	}

	if v == "false" {
		return false
	}

	if i, err := strconv.Atoi(v); err == nil {
		return i
	}

	return v
}
