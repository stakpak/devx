package drivers

import (
	"fmt"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
)

type ComposeDriver struct {
	Path string
}

func (d *ComposeDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "compose"
}

func (d *ComposeDriver) ApplyAll(stack *stack.Stack) error {

	composeFile := stack.GetContext().CompileString("_")
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
				composeFile = composeFile.Fill(resourceIter.Value())
			}
		}
	}

	if !foundResources {
		return nil
	}

	composeFile, err := utils.RemoveMeta(composeFile)
	if err != nil {
		return err
	}
	data, err := yaml.Encode(composeFile)
	if err != nil {
		return err
	}

	composeFilePath := path.Join(d.Path, "docker-compose.yml")
	if _, err := os.Stat(d.Path); os.IsNotExist(err) {
		os.MkdirAll(d.Path, 0700)
	}
	os.WriteFile(composeFilePath, data, 0700)

	fmt.Printf("[compose] applied resources to \"%s\"\n", composeFilePath)

	return nil
}
