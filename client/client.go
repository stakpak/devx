package client

import (
	"context"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
	cueflow "cuelang.org/go/tools/flow"
)

var (
	EnvironmentSelector = cue.Str("environments")
	ComponentSelector   = cue.Str("components")
)

var environmentMap cue.Value
var env string

func Run(environment string, configDir string) error {
	env = environment
	buildConfig := &cueload.Config{
		Dir:     configDir,
		Overlay: map[string]cueload.Source{},
	}
	instances := cueload.Instances([]string{}, buildConfig)

	ctx := cuecontext.New()

	v := ctx.BuildInstance(instances[0])
	if v.Err() != nil {
		return v.Err()
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

	cfg := &cueflow.Config{
		FindHiddenTasks: true,
		Root:            cue.MakePath(ComponentSelector),
	}

	flow := cueflow.New(
		cfg,
		v,
		taskFunc,
	)

	err := flow.Run(context.TODO())
	if err != nil {
		return err
	}

	result := flow.Value()

	printManifest(result)

	return nil
}

func printManifest(result cue.Value) {
	manifest := result.Context().CompileString("{}")

	componentsPath := cue.MakePath(ComponentSelector)
	iter, err := result.LookupPath(componentsPath).Fields()
	if err != nil {
		panic(err)
	}
	for iter.Next() {
		manifest = manifest.Fill(
			iter.Value().Lookup("$children"),
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

func taskFunc(v cue.Value) (cueflow.Runner, error) {
	namePath := cue.ParsePath("$guku.component")
	componentName := v.LookupPath(namePath)
	if !componentName.Exists() {
		// Not a task
		return nil, nil
	}

	return cueflow.RunnerFunc(func(t *cueflow.Task) error {
		componentName, err := t.Value().LookupPath(namePath).String()
		if err != nil {
			return err
		}

		deps := []string{}

		for _, t := range t.Dependencies() {
			componentSelectors := t.Path().Selectors()
			componentId := componentSelectors[len(componentSelectors)-1].String()
			deps = append(deps, componentId)
		}

		transformer := environmentMap.LookupPath(cue.ParsePath(fmt.Sprintf("%s.%s", env, componentName)))
		if transformer.Err() != nil {
			return transformer.Err()
		}

		transformer, components, err := applyTransformerFeedForward(
			transformer,
			map[string][]string{
				"dependencies": deps,
			},
			t.Value(),
		)
		if err != nil {
			return err
		}

		// call next transformer

		component, err := applyTransformerFeedBack(
			transformer,
			components,
		)
		if err != nil {
			return err
		}

		return t.Fill(component)
	}), nil
}

func applyTransformerFeedForward(transformer cue.Value, context interface{}, component cue.Value) (cue.Value, cue.Value, error) {
	ctx := transformer.Context()
	bottom := ctx.CompileString("_|_")

	transformerInputType, _ := transformer.LookupPath(cue.ParsePath("$guku.transformer.component")).String()
	componentType, _ := component.LookupPath(cue.ParsePath("$guku.component")).String()
	if transformerInputType != componentType {
		return transformer, bottom, fmt.Errorf("Transformer expecting input component %s but got %s", transformerInputType, componentType)
	}

	input := ctx.CompileString("{}")
	input = input.FillPath(cue.ParsePath("input.component"), component)
	input = input.FillPath(cue.ParsePath("input.context"), context)

	transformer = transformer.FillPath(
		cue.MakePath(),
		input,
	)
	if transformer.Err() != nil {
		return transformer, bottom, transformer.Err()
	}

	transformer, err := populate("feedforward", transformer)
	if err != nil {
		return transformer, bottom, err
	}

	components, err := GetConcrete(transformer, "feedforward.components")
	if err != nil {
		return transformer, bottom, err
	}
	return transformer, components, nil
}

func applyTransformerFeedBack(transformer cue.Value, components cue.Value) (cue.Value, error) {
	ctx := transformer.Context()
	bottom := ctx.CompileString("_|_")

	input := ctx.CompileString("{}")
	input = input.FillPath(cue.ParsePath("feedforward.components"), components)

	transformer = transformer.FillPath(
		cue.MakePath(),
		input,
	)
	if transformer.Err() != nil {
		return bottom, transformer.Err()
	}

	transformer, err := populate("feedback", transformer)
	if err != nil {
		return bottom, err
	}

	transformer = transformer.FillPath(
		cue.ParsePath("feedback.component.$children"),
		components,
	)

	component, err := GetConcrete(transformer, "feedback.component")
	if err != nil {
		return bottom, err
	}
	return component, nil
}

func removeMeta(value cue.Value) (cue.Value, error) {
	result := value.Context().CompileString("{}")

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

func populate(path string, v cue.Value) (cue.Value, error) {
	bottom := v.Context().CompileString("_|_")

	pathsToFill := []cue.Path{}
	output := v.LookupPath(cue.ParsePath(path))
	Walk(output, func(v cue.Value) bool {
		gukuAttr := v.Attribute("guku")
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
		v = v.FillPath(fieldPath, "dummy")
		if v.Err() != nil {
			return bottom, v.Err()
		}
	}

	return v, nil
}

func GetConcrete(v cue.Value, path string) (cue.Value, error) {
	bottom := v.Context().CompileString("_|_")
	valuePath := cue.ParsePath(path)
	value := v.LookupPath(valuePath)
	if value.Err() != nil {
		return bottom, value.Err()
	}
	err := value.Validate(cue.Concrete(true))
	if err != nil {
		return bottom, err
	}
	return value, nil
}
func Walk(v cue.Value, before func(cue.Value) bool, after func(cue.Value)) {
	switch v.Kind() {
	case cue.StructKind:
		if before != nil && !before(v) {
			return
		}
		iter, _ := v.Fields(cue.All())

		for iter.Next() {
			Walk(iter.Value(), before, after)
		}
	case cue.ListKind:
		if before != nil && !before(v) {
			return
		}
		iter, _ := v.List()
		for iter.Next() {
			Walk(iter.Value(), before, after)
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
