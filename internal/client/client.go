package client

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	cueflow "cuelang.org/go/tools/flow"

	"go.dagger.io/dagger/compiler"
)

var (
	EnvironmentSelector = cue.Str("environments")
	ComponentSelector   = cue.Str("components")
)

var environmentMap *compiler.Value
var env string

func Run(chosenEnvironment string, configDir string) error {
	env = chosenEnvironment

	v, err := compiler.Build(context.TODO(), configDir, nil)
	if err != nil {
		return err
	}

	environmentPath := cue.MakePath(EnvironmentSelector)
	environmentMap = v.LookupPath(environmentPath)
	if !environmentMap.Exists() {
		return fmt.Errorf("Couldn't find environments")
	}

	componentsPath := cue.MakePath(ComponentSelector)
	components := v.LookupPath(componentsPath)
	if !components.Exists() {
		return fmt.Errorf("Couldn't find components")
	}

	result, err := RunFlow(v)
	if err != nil {
		return err
	}

	printManifest(result.Cue())

	return nil
}

func printManifest(result cue.Value) {
	manifest := result.Context().CompileString("_")

	componentsPath := cue.MakePath(ComponentSelector)
	iter, err := result.LookupPath(componentsPath).Fields()
	if err != nil {
		panic(err)
	}
	for iter.Next() {
		manifest = manifest.Fill(
			iter.Value().LookupPath(cue.ParsePath("$guku.children")),
		)
	}

	iter, _ = manifest.Fields()
	for iter.Next() {
		v, err := removeMeta(iter.Value())
		if err != nil {
			panic(err)
		}

		data, err := yaml.Encode(v)
		if err != nil {
			panic(err)
		}
		fmt.Printf("---\n%s", string(data))
	}
}

func taskFunc(flowVal cue.Value) (cueflow.Runner, error) {
	v := compiler.Wrap(flowVal)

	namePath := cue.ParsePath("$guku.component")
	componentName := v.LookupPath(namePath)
	if !componentName.Exists() {
		// Not a task
		return nil, nil
	}
	componentNameString, _ := componentName.String()

	return cueflow.RunnerFunc(func(t *cueflow.Task) error {
		value := compiler.Wrap(t.Value())

		transformer := environmentMap.LookupPath(cue.ParsePath(fmt.Sprintf("%s.%s", env, componentNameString)))
		if transformer.Cue().Err() != nil {
			if strings.Contains(transformer.Cue().Err().Error(), fmt.Sprintf("field not found: %s", componentNameString)) {
				return nil
			}
			return transformer.Cue().Err()
		}

		deps := []string{}
		for _, t := range t.Dependencies() {
			taskValue := compiler.Wrap(t.Value())
			id, err := taskValue.LookupPath(cue.ParsePath("$guku.id")).String()
			if err != nil {
				return err
			}
			deps = append(deps, id)
		}

		components, err := applyTransformerFeedForward(
			transformer,
			map[string][]string{
				"dependencies": deps,
			},
			value,
		)
		if err != nil {
			return err
		}

		// run flow for nested components (this will update components)
		// components, err = RunFlow(components)
		// if err != nil {
		// 	return err
		// }

		component, err := applyTransformerFeedBack(
			transformer,
			components.LookupPath(cue.MakePath(ComponentSelector)),
		)
		if err != nil {
			return err
		}

		if err := t.Fill(component.Cue()); err != nil {
			return err
		}

		return nil
	}), nil
}

func applyTransformerFeedForward(transformer *compiler.Value, context interface{}, component *compiler.Value) (*compiler.Value, error) {
	bottom, _ := compiler.Compile("", "_|_")

	transformerInputTraits := transformer.LookupPath(cue.ParsePath("input.component.$guku.traits"))
	componentTraits := component.LookupPath(cue.ParsePath("$guku.traits"))

	transFields, _ := transformerInputTraits.Fields()
	for _, field := range transFields {
		trait := cue.MakePath(field.Selector)
		hasTrait := componentTraits.LookupPath(trait).Exists()
		if !hasTrait {
			metadata := component.LookupPath(cue.ParsePath("$guku"))
			return bottom, fmt.Errorf("Transformer input component %s missing trait %s", metadata, field.Selector)
		}
	}

	input, _ := compiler.Compile("", "_")
	input.FillPath(cue.ParsePath("input.component"), component)
	input.FillPath(cue.ParsePath("input.context"), context)

	err := transformer.FillPath(
		cue.MakePath(),
		input,
	)
	if err != nil {
		return bottom, err
	}

	err = populateGeneratedFields(transformer)
	if err != nil {
		return bottom, err
	}

	// components don't have to be concrete
	components := transformer.LookupPath(cue.ParsePath("feedforward"))

	return components, nil
}

func applyTransformerFeedBack(transformer *compiler.Value, components *compiler.Value) (*compiler.Value, error) {
	bottom, _ := compiler.Compile("", "_|_")

	err := components.Validate(cue.Concrete(true))
	if err != nil {
		return bottom, err
	}

	input, _ := compiler.Compile("", "_")
	input.FillPath(cue.ParsePath("feedforward.components"), components)

	err = transformer.FillPath(
		cue.MakePath(),
		input,
	)
	if err != nil {
		return bottom, err
	}

	transformer.FillPath(
		cue.ParsePath("feedback.component.$guku.children"),
		components,
	)

	component, err := GetConcrete(transformer, "feedback.component")
	if err != nil {
		return bottom, err
	}
	return component, nil
}

func removeMeta(value cue.Value) (cue.Value, error) {
	result := value.Context().CompileString("_")

	iter, err := value.Fields()
	if err != nil {
		return result, err
	}

	for iter.Next() {
		v := iter.Value()
		selectors := v.Path().Selectors()
		selectors = selectors[1:]
		path := cue.MakePath(selectors...)
		if !strings.HasPrefix(path.String(), "$") {
			result = result.FillPath(path, v)
		}
	}

	return result, nil
}

func populateGeneratedFields(v *compiler.Value) error {
	pathsToFill := []cue.Path{}
	Walk(v, func(v *compiler.Value) bool {
		gukuAttr := v.Cue().Attribute("guku")
		if !v.IsConcrete() && gukuAttr.Err() == nil {
			isGenerated, _ := gukuAttr.Flag(0, "generate")
			if isGenerated {
				selectors := v.Path().Selectors()
				pathsToFill = append(pathsToFill, cue.MakePath(selectors[2:]...))
			}
		}
		return true
	}, nil)

	for _, path := range pathsToFill {
		selectors := path.Selectors()[1:]
		fieldPath := cue.MakePath(selectors...)
		err := v.FillPath(fieldPath, "dummy")
		if err != nil {
			return err
		}
	}

	return nil
}

func GetConcrete(v *compiler.Value, path string) (*compiler.Value, error) {
	bottom, _ := compiler.Compile("", "_|_")
	valuePath := cue.ParsePath(path)
	value := v.LookupPath(valuePath)
	if value.Cue().Err() != nil {
		return bottom, value.Cue().Err()
	}
	err := value.Validate(cue.Concrete(true))
	if err != nil {
		return bottom, err
	}
	return value, nil
}

func Walk(v *compiler.Value, before func(*compiler.Value) bool, after func(*compiler.Value)) {
	switch v.Kind() {
	case cue.StructKind:
		if before != nil && !before(v) {
			return
		}
		fields, _ := v.Fields(cue.All())

		for _, field := range fields {
			Walk(field.Value, before, after)
		}
	case cue.ListKind:
		if before != nil && !before(v) {
			return
		}
		values, _ := v.List()
		for _, value := range values {
			Walk(value, before, after)
		}
	default:
		if before != nil {
			before(v)
		}
	}
	if after != nil {
		after(v)
	}
}

func RunFlow(v *compiler.Value) (*compiler.Value, error) {
	cfg := &cueflow.Config{
		FindHiddenTasks: true,
		Root:            cue.MakePath(ComponentSelector),
	}

	flow := cueflow.New(
		cfg,
		v.Cue(),
		taskFunc,
	)

	err := flow.Run(context.TODO())
	if err != nil {
		return v, err
	}

	return compiler.Wrap(flow.Value()), nil
}
