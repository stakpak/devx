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

type GitlabDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *GitlabDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "gitlab"
}

func (d *GitlabDriver) ApplyAll(stack *stack.Stack, stdout bool) error {
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

				if stdout {
					_, err := os.Stdout.Write(data)
					if err != nil {
						return err
					}
					continue
				}

				if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
					os.MkdirAll(d.Config.Output.Dir, 0700)
				}
				filePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)
				os.WriteFile(filePath, data, 0700)

				log.Infof("[gitlab] applied a resource to \"%s\"", filePath)
			}
		}
	}

	return nil
}
