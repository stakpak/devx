package drivers

import (
	"encoding/json"
	"os"
	"path"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type TerraformDriver struct {
	Path string
}

func (d *TerraformDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "terraform"
}

func (d *TerraformDriver) ApplyAll(stack *stack.Stack) error {

	terraformFile := stack.GetContext().CompileString("_")
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
				terraformFile = terraformFile.Fill(resourceIter.Value())
			}
		}
	}

	if !foundResources {
		return nil
	}

	terraformFile, err := utils.RemoveMeta(terraformFile)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(terraformFile, "", "  ")
	if err != nil {
		return err
	}

	terraformFilePath := path.Join(d.Path, "generated.tf.json")
	if _, err := os.Stat(d.Path); os.IsNotExist(err) {
		os.MkdirAll(d.Path, 0700)
	}
	os.WriteFile(terraformFilePath, data, 0700)

	log.Infof("[terraform] applied resources to \"%s\"", terraformFilePath)

	return nil
}
