package drivers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/stakpak/devx/pkg/stack"
	"github.com/stakpak/devx/pkg/stackbuilder"
	"github.com/stakpak/devx/pkg/utils"
)

type KubernetesDriver struct {
	Config stackbuilder.DriverConfig
}

func (d *KubernetesDriver) match(resource cue.Value) bool {
	driverName, _ := resource.LookupPath(cue.ParsePath("$metadata.labels.driver")).String()
	return driverName == "kubernetes"
}

func (d *KubernetesDriver) ApplyAll(stack *stack.Stack, stdout bool) error {
	manifests := map[string][]byte{}
	defaultFilePath := path.Join(d.Config.Output.Dir, d.Config.Output.File)

	for _, componentId := range stack.GetTasks() {
		component, _ := stack.GetComponent(componentId)

		resourceIter, _ := component.LookupPath(cue.ParsePath("$resources")).Fields()
		for resourceIter.Next() {
			v := resourceIter.Value()
			if d.match(v) {
				filePath := defaultFilePath
				outputSubdirLabel := v.LookupPath(cue.ParsePath("$metadata.labels.\"output-subdir\""))
				if outputSubdirLabel.Exists() {
					outputSubdir, err := outputSubdirLabel.String()
					if err != nil {
						return err
					}

					filePath = path.Join(d.Config.Output.Dir, outputSubdir, d.Config.Output.File)
				}

				resource, err := utils.RemoveMeta(v)
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

				filePath = filepath.Join(filePath, fmt.Sprintf("%s-%s.yml", nameString, strings.ToLower(kindString)))

				manifests[filePath] = data
			}
		}
	}

	if len(manifests) == 0 {
		return nil
	}

	for filePath, fileValue := range manifests {

		if stdout {
			if _, err := os.Stdout.Write([]byte("---\n")); err != nil {
				return err
			}
			if _, err := os.Stdout.Write(fileValue); err != nil {
				return err
			}
			_, err := os.Stdout.Write([]byte("\n"))
			return err
		}

		if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(filePath), 0700)
		}
		_, err := os.Stat(filePath)

		if !os.IsNotExist(err) {
			os.Remove(filePath)
		}

		os.WriteFile(filePath, fileValue, 0700)

		log.Infof("[kubernetes] applied resources to \"%s\"", filePath)
	}

	return nil
}
