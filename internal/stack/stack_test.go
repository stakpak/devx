package stack

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
)

var stackString1 = `
components: {
	a: {
		$metadata: id: "a"

		tada: b.todo
	}
	b: {
		$metadata: id: "b"

		todo: 123
	}
	c: {
		$metadata: id: "c"

		bla: a.tada
	}
}
`

func TestNew(t *testing.T) {
	ctx := cuecontext.New()
	value := ctx.CompileString(stackString1)

	stack, err := NewStack(value)
	if err != nil {
		t.Error(err)
	}

	keys := []string{"a", "b", "c"}
	parents := [][]string{
		{"b"},
		{},
		{"a"},
	}

	for i, key := range keys {
		dependencies, err := stack.GetDependencies(key)
		if err != nil {
			t.Errorf("Error in key %s: %s", key, err)
		}
		if len(dependencies) != len(parents[i]) {
			t.Errorf("Error in key %s: expected dependencies length %s but found %s", key, parents[i], dependencies)
		}
		for j := range parents[i] {
			if parents[i][j] != dependencies[j] {
				t.Errorf("Error in key %s: expected dependencies %s but found %s", key, parents[i], dependencies)
			}
		}
	}
}

func TestTaskOrder(t *testing.T) {
	stackString := `
	components: {
		a: {
			$metadata: id: "a"
			todo: 123
		}
		f: {
			$metadata: id: "f"
			todo: 123
		}
		b: {
			$metadata: id: "b"
			todo: a.todo
		}
		c: {
			$metadata: id: "c"
			todo: b.todo
		}
	}
	`
	ctx := cuecontext.New()
	value := ctx.CompileString(stackString)

	stack, err := NewStack(value)
	if err != nil {
		t.Error(err)
	}

	taskOrder := stack.GetTasks()

	expectedOrder := []string{"f", "a", "b", "c"}

	for i, k := range expectedOrder {
		if taskOrder[i] != k {
			t.Errorf("Error expected element %d in order to be %s but got %s", i, k, taskOrder[i])
		}
	}
}
