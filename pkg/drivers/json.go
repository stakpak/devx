package drivers

import (
	"encoding/json"
	"os"
	"path"

	"cuelang.org/go/cue"
	"github.com/devopzilla/guku-devx/pkg/stack"
	"github.com/devopzilla/guku-devx/pkg/stackbuilder"
	"github.com/devopzilla/guku-devx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type JSONDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *JSONDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "json"
}

func (d *JSONDriver) ApplyAll(stack *stack.Stack, stdout bool) error {
	jsonFile := stack.GetContext().CompileString("_")
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
				jsonFile = jsonFile.FillPath(cue.ParsePath(""), resourceIter.Value())
			}
		}
	}

	if !foundResources {
		return nil
	}

	jsonFile, err := utils.RemoveMeta(jsonFile)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(jsonFile, "", "  ")
	if err != nil {
		return err
	}

	if stdout {
		_, err := os.Stdout.Write(data)
		return err
	}

	if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Output.Dir, 0700)
	}
	filePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
	os.WriteFile(filePath, data, 0700)

	log.Infof("[json] applied resources to \"%s\"", filePath)

	return nil
}
