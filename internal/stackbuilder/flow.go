package stackbuilder

import (
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

	ctx := value.Context()
	flow := Flow{
		match:    ctx.CompileString("_"),
		exclude:  ctx.CompileString("_"),
		pipeline: make([]cue.Value, 0),
	}
	flow.match = flow.match.Fill(matchValue)
	flow.exclude = flow.exclude.Fill(excludeValue)
	pipelineIter, _ := pipelineValue.List()
	for pipelineIter.Next() {
		transformer := ctx.CompileString("_")
		transformer = transformer.Fill(pipelineIter.Value())
		flow.pipeline = append(flow.pipeline, transformer)
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

		matchSubfieldsIter, _ := matchIter.Value().Fields()
		for matchSubfieldsIter.Next() {
			matchSubfieldName := utils.GetLastPathFragement(matchSubfieldsIter.Value())
			componentSubfield := componentField.LookupPath(cue.ParsePath(matchSubfieldName))

			if !componentSubfield.Exists() || !componentSubfield.Equals(matchSubfieldsIter.Value()) {
				return false
			}
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

func (f *Flow) Run(stack *stack.Stack, componentId string) error {
	component, err := stack.GetComponent(componentId)
	if err != nil {
		return err
	}

	if !f.Match(component) {
		return nil
	}

	dependencies, err := stack.GetDependencies(componentId)
	if err != nil {
		return err
	}

	context := component.Context().CompileString("_")
	context = context.FillPath(cue.ParsePath("dependencies"), dependencies)

	for _, transformer := range f.pipeline {
		component, err = Transform(
			transformer,
			component,
			context,
		)
		if err != nil {
			return err
		}
	}

	err = stack.UpdateComponent(componentId, component)
	if err != nil {
		return err
	}

	return nil
}

func Transform(transformer cue.Value, component cue.Value, context cue.Value) (cue.Value, error) {
	bottom := transformer.Context().CompileString("_|_")

	transformerContext := transformer.LookupPath(cue.ParsePath("context"))
	transformerContext = transformerContext.Fill(context)
	if transformerContext.Err() != nil {
		return bottom, transformer.Err()
	}
	transformerContext = populateGeneratedFields(transformerContext)
	if transformerContext.Err() != nil {
		return bottom, transformer.Err()
	}

	transformer = transformer.FillPath(
		cue.ParsePath("context"),
		transformerContext,
	)
	if transformer.Err() != nil {
		return bottom, transformer.Err()
	}
	transformer = transformer.FillPath(
		cue.ParsePath("input"),
		component,
	)
	if transformer.Err() != nil {
		return bottom, transformer.Err()
	}

	return transformer.LookupPath(cue.ParsePath("output")), nil
}

func populateGeneratedFields(value cue.Value) cue.Value {
	result := value.Context().CompileString("_")
	result = result.Fill(value)

	pathsToFill := []cue.Path{}
	utils.Walk(value, func(v cue.Value) bool {
		gukuAttr := v.Attribute("guku")
		if !v.IsConcrete() && gukuAttr.Err() == nil {
			isGenerated, _ := gukuAttr.Flag(0, "generate")
			if isGenerated {
				selectors := v.Path().Selectors()
				pathsToFill = append(pathsToFill, cue.MakePath(selectors[1:]...))
			}
		}
		return true
	}, nil)

	for _, path := range pathsToFill {
		result = result.FillPath(path, "dummy")
		if result.Err() != nil {
			return result
		}
	}

	return result
}
