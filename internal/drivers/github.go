package drivers

import (
	"fmt"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type GitHubDriver struct {
	Path string
}

func (d *GitHubDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "github"
}

func (d *GitHubDriver) ApplyAll(stack *stack.Stack) error {
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

				fileName := fmt.Sprintf("%s-github-workflow.yml", resourceIter.Label())
				resourceFilePath := path.Join(d.Path, fileName)
				if _, err := os.Stat(d.Path); os.IsNotExist(err) {
					os.MkdirAll(d.Path, 0700)
				}
				os.WriteFile(resourceFilePath, data, 0700)
			}
		}
	}

	if !foundResources {
		return nil
	}

	log.Infof("[github] applied resources to \"%s/*-github-workflow.yml\"", d.Path)

	return nil
}
