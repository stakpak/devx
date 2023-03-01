package drivers

import (
	"fmt"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku-devx/pkg/stack"
	"devopzilla.com/guku-devx/pkg/stackbuilder"
	"devopzilla.com/guku-devx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type GitHubDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *GitHubDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "github"
}

func (d *GitHubDriver) ApplyAll(stack *stack.Stack, stdout bool) error {
	foundResources := false

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			if d.match(resourceIter.Value()) {
				foundResources = true
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
				fileName := fmt.Sprintf("%s.yml", resourceIter.Label())
				if d.Config.Output.File != "" {
					fileName = d.Config.Output.File
				}
				filePath := path.Join(d.Config.Output.Dir, fileName)
				os.WriteFile(filePath, data, 0700)
			}
		}
	}

	if !foundResources {
		return nil
	}

	log.Infof("[github] applied resources to \"%s/*-github-workflow.yml\"", d.Config.Output.Dir)

	return nil
}
