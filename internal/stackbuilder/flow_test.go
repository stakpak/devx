package stackbuilder

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"devopzilla.com/guku/internal/stack"
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
	if !flow.Match(componentMatch) {
		t.Error("Expected component to match flow")
	}

	componentMissing := ctx.CompileString(componentMissingFlow1)
	if flow.Match(componentMissing) {
		t.Error("Expected component not to match flow: missing trait")
	}

	componentDiff := ctx.CompileString(componentDiffFlow1)
	if flow.Match(componentDiff) {
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
	if flow.Match(componentExcludeLabel) {
		t.Error("Expected component not to match flow: excluded label")
	}
}

func TestTransform(t *testing.T) {
	ctx := cuecontext.New()
	transformer := ctx.CompileString(`
	args: {
		arg1: 123
	}
	context: {
		field1: string
		gen: string @guku(generate)
	}
	input: {
		field1: int
		field2: int
	}
	output: {
		input
		field3: input.field1
		$resources: {
			a: {
				c: context.field1
				d: context.gen
			}
		}
	}
	`)

	context := ctx.CompileString(`
	field1: "canary"
	`)

	input := ctx.CompileString(`
	field1: 1
	field2: 2
	`)

	expectedOutput := ctx.CompileString(`
	field1: 1
	field2: 2
	field3: 1
	$resources: {
		a: {
			c: "canary"
			d: "dummy"
		}
	}
	`)

	output, err := Transform(transformer, input, context)
	if err != nil {
		t.Error(err)
	}

	if !output.Equals(expectedOutput) {
		t.Errorf("Expected output to be \n%v \nbut got \n%v\n", expectedOutput, output)
	}
}

var flowString2 = `
match: {}
exclude: {}
pipeline: [
	{
		args: {}
		context: {}
		input: {
			...
		}
		output: {
			input
			$resources: trans1: "done"
		}
	},
	{
		args: {}
		context: {
			test: string @guku(generate)
		}
		input: {
			...
		}
		output: {
			input
			test: context.test
			$resources: trans2: "done"
		}
	},
	{
		args: {}
		context: {
			dependencies: [...string]
		}
		input: {
			...
		}
		output: {
			input
			trans3: context.dependencies
			$resources: trans3: "done"
		}
	},
]
`
var stackString = `
components: {
	a: {
		$metadata: id: "a"
		todo: b.todo
	}
	b: {
		$metadata: id: "b"
		todo: 123
	}
}
`
var expectedOutputComponent = `
$metadata: id: "a"
todo: 123
test: "dummy"
trans3: ["b"]
$resources: {
	trans1: "done"
	trans2: "done"
	trans3: "done"
}
`

func TestFlowRun(t *testing.T) {
	ctx := cuecontext.New()
	flowValue := ctx.CompileString(flowString2)
	flow, _ := NewFlow(flowValue)

	stackValue := ctx.CompileString(stackString)
	stack, _ := stack.NewStack(stackValue)

	err := flow.Run(stack, "a")
	if err != nil {
		t.Error(err)
	}

	component, _ := stack.GetComponent("a")

	expectedOutput := ctx.CompileString(expectedOutputComponent)
	if !component.Equals(expectedOutput) {
		t.Errorf("Expected component to be \n%v \nbut got \n%v\n", expectedOutput, component)
	}
}
