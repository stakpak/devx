package drivers

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"cuelang.org/go/cue"
	log "github.com/sirupsen/logrus"
	"github.com/stakpak/devx/pkg/stack"
	"github.com/stakpak/devx/pkg/stackbuilder"
	"github.com/stakpak/devx/pkg/utils"
)

type TerraformDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *TerraformDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "terraform"
}

func (d *TerraformDriver) ApplyAll(stack *stack.Stack, stdout bool) error {

	terraformFiles := map[string]cue.Value{}
	defaultFilePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
	foundResources := false

	common := stack.GetContext().CompileString("_")
	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			v := resourceIter.Value()
			if d.match(v) {
				foundResources = true
				filePath := defaultFilePath

				outputSubdirLabel := v.LookupPath(cue.ParsePath("$metadata.labels.\"output-subdir\""))
				if outputSubdirLabel.Exists() {
					outputSubdir, err := outputSubdirLabel.String()
					if err != nil {
						return err
					}

					if outputSubdir == "*" {
						v, err := utils.RemoveMeta(v)
						if err != nil {
							return err
						}
						common = common.FillPath(cue.ParsePath(""), v)
						continue
					}

					filePath = path.Join(d.Config.Output.Dir, outputSubdir, d.Config.Output.File)
				}

				if _, ok := terraformFiles[filePath]; !ok {
					terraformFiles[filePath] = stack.GetContext().CompileString("_")
				}

				v, err := utils.RemoveMeta(v)
				if err != nil {
					return err
				}
				terraformFiles[filePath] = terraformFiles[filePath].FillPath(cue.ParsePath(""), v)
			}
		}
	}

	if !foundResources {
		return nil
	}

	for filePath, fileValue := range terraformFiles {
		fileValue := fileValue.FillPath(cue.ParsePath(""), common)
		data, err := json.MarshalIndent(fileValue, "", "  ")
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

		if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(filePath), 0700)
		}
		os.WriteFile(filePath, data, 0700)

		log.Infof("[terraform] applied resources to \"%s\"", filePath)
	}

	return nil
}
