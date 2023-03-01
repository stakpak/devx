package drivers

import (
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"github.com/devopzilla/guku-devx/pkg/stack"
	"github.com/devopzilla/guku-devx/pkg/stackbuilder"
	"github.com/devopzilla/guku-devx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type YAMLDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *YAMLDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "yaml"
}

func (d *YAMLDriver) ApplyAll(stack *stack.Stack, stdout bool) error {
	yamlFile := stack.GetContext().CompileString("_")
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
				yamlFile = yamlFile.FillPath(cue.ParsePath(""), resourceIter.Value())
			}
		}
	}

	if !foundResources {
		return nil
	}

	yamlFile, err := utils.RemoveMeta(yamlFile)
	if err != nil {
		return err
	}
	data, err := yaml.Encode(yamlFile)
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

	log.Infof("[yaml] applied resources to \"%s\"", filePath)

	return nil
}
