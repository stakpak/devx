package drivers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type KubernetesDriver struct {
	Path string
}

func (d *KubernetesDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "kubernetes"
}

func (d *KubernetesDriver) ApplyAll(stack *stack.Stack) error {
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

				kind := resource.LookupPath(cue.ParsePath("kind"))
				if kind.Err() != nil {
					return kind.Err()
				}

				kindString, err := kind.String()
				if err != nil {
					return err
				}

				name := resource.LookupPath(cue.ParsePath("metadata.name"))
				if kind.Err() != nil {
					return kind.Err()
				}

				nameString, err := name.String()
				if err != nil {
					return err
				}

				data, err := yaml.Encode(resource)
				if err != nil {
					return err
				}

				fileName := fmt.Sprintf("%s-%s.yml", nameString, strings.ToLower(kindString))
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

	log.Infof("[kubernetes] applied resources to \"%s/*.yml\"", d.Path)

	return nil
}
