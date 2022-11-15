package stackbuilder

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
)

var flowString1 = `
match: {
	traits: {
		balabizo: null
		tada: 123
	}
}
exclude: {
	labels: {
		tada: "abc"
		toto: 123
	}
}
pipeline: []
`

var componentMatchFlow1 = `
$metadata: {
	traits: {
		balabizo: null
		tada: 123
		bla: null
	}
}
`

var componentMissingFlow1 = `
$metadata: {
	traits: {
		balabizo: null
		bla: null
	}
}
`

var componentDiffFlow1 = `
$metadata: {
	traits: {
		balabizo: null
		tada: null
	}
}
`

var componentExcludeLabelFlow1 = `
$metadata: {
	traits: {
		balabizo: null
		tada: 123
	}
	labels: toto: 123
}
`

func TestNewFlow(t *testing.T) {
	ctx := cuecontext.New()
	flowValue := ctx.CompileString(flowString1)

	_, err := NewFlow(flowValue)
	if err != nil {
		t.Error(err)
	}
}

func TestMatchFlow(t *testing.T) {
	ctx := cuecontext.New()
	flowValue := ctx.CompileString(flowString1)

	flow, err := NewFlow(flowValue)
	if err != nil {
		t.Error(err)
	}

	componentMatch := ctx.CompileString(componentMatchFlow1)
	if !flow.Matches(componentMatch) {
		t.Error("Expected component to match flow")
	}

	componentMissing := ctx.CompileString(componentMissingFlow1)
	if flow.Matches(componentMissing) {
		t.Error("Expected component not to match flow: missing trait")
	}

	componentDiff := ctx.CompileString(componentDiffFlow1)
	if flow.Matches(componentDiff) {
		t.Error("Expected component not to match flow: different trait value")
	}
}

func TestMatchExcludeFlow(t *testing.T) {
	ctx := cuecontext.New()
	flowValue := ctx.CompileString(flowString1)

	flow, err := NewFlow(flowValue)
	if err != nil {
		t.Error(err)
	}

	componentExcludeLabel := ctx.CompileString(componentExcludeLabelFlow1)
	if flow.Matches(componentExcludeLabel) {
		t.Error("Expected component not to match flow: excluded label")
	}
}
