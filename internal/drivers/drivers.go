package drivers

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
)

type Driver interface {
	match(resource cue.Value) bool
	ApplyAll(stack *stack.Stack) error
}

type ComposeDriver struct{}

func (d *ComposeDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "compose"
}

func (d *ComposeDriver) ApplyAll(stack *stack.Stack) error {

	composeFile := stack.GetContext().CompileString("_")

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				composeFile = composeFile.Fill(resourceIter.Value())
			}
		}
	}

	composeFile, err := utils.RemoveMeta(composeFile)
	if err != nil {
		return err
	}
	data, err := yaml.Encode(composeFile)
	if err != nil {
		return err
	}
	fmt.Printf("---\n%s", string(data))

	return nil
}
