package drivers

import (
	"encoding/json"
	"os"
	"path"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type TerraformDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *TerraformDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "terraform"
}

func (d *TerraformDriver) ApplyAll(stack *stack.Stack, stdout bool) error {

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

	if stdout {
		_, err := os.Stdout.Write(data)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write([]byte("\n"))
		return err
	}

	if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Output.Dir, 0700)
	}
	filePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
	os.WriteFile(filePath, data, 0700)

	log.Infof("[terraform] applied resources to \"%s\"", filePath)

	return nil
}
