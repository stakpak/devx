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

type GitlabDriver struct {
	Path string
	File string
}

func (d *GitlabDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "gitlab"
}

func (d *GitlabDriver) ApplyAll(stack *stack.Stack) error {
	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				resource, err := utils.RemoveMeta(resourceIter.Value())
				if err != nil {
					return err
				}

				data, err := yaml.Encode(resource)
				if err != nil {
					return err
				}

				resourceFilePath := path.Join(d.Path, d.File)
				if _, err := os.Stat(d.Path); os.IsNotExist(err) {
					os.MkdirAll(d.Path, 0700)
				}
				os.WriteFile(resourceFilePath, data, 0700)

				fmt.Printf("[gitlab] applied a resource to \"%s\"\n", resourceFilePath)
			}
		}
	}

	return nil
}
