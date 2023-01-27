package drivers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type KubernetesDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *KubernetesDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "kubernetes"
}

func (d *KubernetesDriver) ApplyAll(stack *stack.Stack) error {
	foundResources := false

	if _, err := os.Stat(d.Config.Output.Dir); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Output.Dir, 0700)
	}

	var singleFile *os.File
	if d.Config.Output.File != "" {
		filePath := filepath.Join(d.Config.Output.Dir, d.Config.Output.File)
		_, err := os.Stat(filePath)
		if !os.IsNotExist(err) {
			os.Remove(filePath)
		}
		singleFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
		if err != nil {
			log.Fatal(err)
		}
		defer singleFile.Close()

		singleFile.WriteString("")
	}

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

				if singleFile != nil {
					singleFile.Write([]byte("---\n"))
					singleFile.Write(data)
				} else {
					fileName := fmt.Sprintf("%s-%s.yml", nameString, strings.ToLower(kindString))
					filePath := path.Join(d.Config.Output.Dir, fileName)
					os.WriteFile(filePath, data, 0700)
				}
			}
		}
	}

	if !foundResources {
		return nil
	}

	log.Infof("[kubernetes] applied resources to \"%s/*.yml\"", d.Config.Output.Dir)

	return nil
}
