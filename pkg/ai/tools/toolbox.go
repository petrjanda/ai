package tools

import (
	"fmt"
	"regexp"
	"strings"
)

type Toolbox []Tool

func NewToolbox(tools ...Tool) Toolbox {
	return tools
}

func (t Toolbox) FindTool(name string) (Tool, error) {
	for _, tool := range t {
		if tool.Name() == name {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("tool not found: %s", name)
}

// BUILDING

func (t Toolbox) AddTools(tools ...Tool) Toolbox {
	return append(t, tools...)
}

func (t Toolbox) ReplaceTool(tool Tool) Toolbox {
	for i, existingTool := range t {
		if existingTool.Name() == tool.Name() {
			t[i] = tool
			break
		}
	}
	return t
}

// FILTERING

var (
	// Filters tools by regex.
	ByRegex = func(regex *regexp.Regexp) func(Tool) bool {
		return func(tool Tool) bool {
			return regex.MatchString(tool.Name())
		}
	}

	// Filters tools by prefix.
	ByPrefix = func(prefix string) func(Tool) bool {
		return func(tool Tool) bool {
			return strings.HasPrefix(tool.Name(), prefix)
		}
	}

	Contains = func(name string) func(Tool) bool {
		return func(tool Tool) bool {
			return strings.Contains(tool.Name(), name)
		}
	}

	// Filters tools by list of names.
	ByList = func(names ...string) func(Tool) bool {
		return func(tool Tool) bool {
			for _, name := range names {
				if tool.Name() == name {
					return true
				}
			}
			return false
		}
	}

	Or = func(predicates ...func(Tool) bool) func(Tool) bool {
		return func(tool Tool) bool {
			for _, predicate := range predicates {
				if predicate(tool) {
					return true
				}
			}

			return false
		}
	}

	And = func(predicates ...func(Tool) bool) func(Tool) bool {
		return func(tool Tool) bool {
			for _, predicate := range predicates {
				if !predicate(tool) {
					return false
				}
			}

			return true
		}
	}
)

func (t Toolbox) KeepTools(predicate func(Tool) bool) Toolbox {
	var tools Toolbox
	for _, tool := range t {
		if predicate(tool) {
			tools = append(tools, tool)
		}
	}
	return tools
}

// Removes tools from toolbox that match the predicate.
func (t Toolbox) RemoveTools(predicate func(Tool) bool) Toolbox {
	var tools Toolbox
	for _, tool := range t {
		if !predicate(tool) {
			tools = append(tools, tool)
		}
	}
	return tools
}
