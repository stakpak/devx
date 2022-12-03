package stackbuilder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
)

type Flow struct {
	match    cue.Value
	exclude  cue.Value
	pipeline []cue.Value
}

func NewFlow(value cue.Value) (*Flow, error) {
	matchValue := value.LookupPath(cue.ParsePath("match"))
	if matchValue.Err() != nil {
		return nil, matchValue.Err()
	}
	excludeValue := value.LookupPath(cue.ParsePath("exclude"))
	if excludeValue.Err() != nil {
		return nil, excludeValue.Err()
	}
	pipelineValue := value.LookupPath(cue.ParsePath("pipeline"))
	if pipelineValue.Err() != nil {
		return nil, pipelineValue.Err()
	}

	flow := Flow{
		match:    matchValue,
		exclude:  excludeValue,
		pipeline: make([]cue.Value, 0),
	}
	pipelineIter, _ := pipelineValue.List()
	for pipelineIter.Next() {
		flow.pipeline = append(flow.pipeline, pipelineIter.Value())
	}

	return &flow, nil
}

func (f *Flow) Match(component cue.Value) bool {
	metadata := component.LookupPath(cue.ParsePath("$metadata"))

	// Check matches
	matchIter, _ := f.match.Fields()
	for matchIter.Next() {
		fieldName := utils.GetLastPathFragement(matchIter.Value())
		componentField := metadata.LookupPath(cue.ParsePath(fieldName))

		if !componentField.Exists() {
			return false
		}

		err := matchIter.Value().Subsume(componentField, cue.Final())
		if err != nil {
			return false
		}
	}

	// Check excludes
	excludeIter, _ := f.exclude.Fields()
	for excludeIter.Next() {
		fieldName := utils.GetLastPathFragement(excludeIter.Value())
		componentField := metadata.LookupPath(cue.ParsePath(fieldName))

		excludedSubfieldsIter, _ := excludeIter.Value().Fields()
		for excludedSubfieldsIter.Next() {
			excludedSubfieldName := utils.GetLastPathFragement(excludedSubfieldsIter.Value())
			componentSubfield := componentField.LookupPath(cue.ParsePath(excludedSubfieldName))

			if componentSubfield.Exists() && componentSubfield.Equals(excludedSubfieldsIter.Value()) {
				return false
			}
		}
	}

	return true
}

func (f *Flow) Run(stack *stack.Stack, componentId string, component cue.Value) (cue.Value, error) {
	if !f.Match(component) {
		return component, nil
	}

	dependencies, err := stack.GetDependencies(componentId)
	if err != nil {
		return component, err
	}

	// Transform
	component = component.FillPath(cue.ParsePath("$dependencies"), dependencies)
	for _, transformer := range f.pipeline {
		component = component.FillPath(cue.ParsePath(""), transformer)
		if component.Err() != nil {
			return component, component.Err()
		}
	}
	component = populateGeneratedFields(component)
	if component.Err() != nil {
		return component, component.Err()
	}

	return component, nil
}

func populateGeneratedFields(value cue.Value) cue.Value {
	pathsToFill := []cue.Path{}
	valuesToFill := []string{}
	utils.Walk(value, func(v cue.Value) bool {
		gukuAttr := v.Attribute("guku")
		if !v.IsConcrete() && gukuAttr.Err() == nil {
			valueToFill := ""

			filePath, found, _ := gukuAttr.Lookup(0, "file")
			if found {
				filePath, err := verifyPath(filePath)
				if err != nil {
					fmt.Printf("\nPath error %s\n", err)
					return true
				}
				content, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Printf("\nFile error %s\n", err)
					return true
				}
				valueToFill = string(content)
			}

			env, found, _ := gukuAttr.Lookup(0, "env")
			if found && valueToFill == "" {
				content, found := os.LookupEnv(env)
				if !found {
					fmt.Printf("\nEnvironment variable %s not set\n", env)
					return true
				}
				valueToFill = content
			}

			isGenerated, _ := gukuAttr.Flag(0, "generate")
			if isGenerated && valueToFill == "" {
				valueToFill = "dummy"
			}

			if valueToFill != "" {
				selectors := v.Path().Selectors()
				pathsToFill = append(pathsToFill, cue.MakePath(selectors[3:]...))
				valuesToFill = append(valuesToFill, valueToFill)
			}
		}
		return true
	}, nil)

	for i, path := range pathsToFill {
		value = value.FillPath(path, valuesToFill[i])
		if value.Err() != nil {
			return value
		}
	}

	return value
}

func verifyPath(path string) (string, error) {
	c := filepath.Clean(path)
	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		return c, errors.New("Unsafe or invalid path specified")
	}
	return r, nil
}
