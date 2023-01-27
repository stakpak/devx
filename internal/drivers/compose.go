package drivers

import (
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type ComposeDriver struct {
	Config stackbuilder.DriverConfig
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

	if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Output.Dir, 0700)
	}
	filePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
	os.WriteFile(filePath, data, 0700)

	log.Infof("[compose] applied resources to \"%s\"", filePath)

	return nil
}
