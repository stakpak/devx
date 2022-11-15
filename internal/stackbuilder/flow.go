package stackbuilder

import (
	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/utils"
)

type Flow struct {
	match    cue.Value
	exclude  cue.Value
	pipeline cue.Value
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
		pipeline: ctx.CompileString("_"),
	}
	flow.match = flow.match.Fill(matchValue)
	flow.exclude = flow.exclude.Fill(excludeValue)
	flow.pipeline = flow.pipeline.Fill(pipelineValue)

	return &flow, nil
}

func (f *Flow) Matches(component cue.Value) bool {
	metadata := component.LookupPath(cue.ParsePath("$metadata"))

	// Check matches
	matchIter, _ := f.match.Fields()
	for matchIter.Next() {
		fieldName := utils.GetLastPathFragement(matchIter.Value())
		componentField := metadata.LookupPath(cue.ParsePath(fieldName))
		err := matchIter.Value().Subsume(componentField)
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
