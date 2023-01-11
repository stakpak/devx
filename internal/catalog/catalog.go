package catalog

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"

	"devopzilla.com/guku/internal/utils"
	log "github.com/sirupsen/logrus"
)

type CatalogItem struct {
	Dependencies []string
	Package      string
	Name         string
	Code         string
	Types        string
	Traits       []string
}

func Publish(configDir string, telemetry string) error {
	if telemetry == "" {
		return fmt.Errorf("telemtry endpoint is required to publish catalog")
	}

	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}

	instances := utils.LoadInstances(configDir, &overlays)
	instance := instances[0]

	pkg := instance.ID()
	deps := instance.Deps

	ctx := cuecontext.New()
	value := ctx.BuildInstance(instance)

	catalog := []CatalogItem{}

	fieldIter, err := value.Fields(cue.Definitions(true))
	if err != nil {
		return err
	}
	for fieldIter.Next() {
		item := fieldIter.Value()
		metadata := item.LookupPath(cue.ParsePath("$metadata"))
		if !metadata.Exists() {
			continue
		}

		stackMeta := metadata.LookupPath(cue.ParsePath("stack"))
		if stackMeta.Exists() {
			log.Infof("Found a stack %s", stackMeta)
			log.Info(item)
			continue
		}

		traitMeta := metadata.LookupPath(cue.ParsePath("traits"))
		if traitMeta.Exists() {
			traits := []string{}
			traitIter, _ := traitMeta.Fields()
			for traitIter.Next() {
				traits = append(traits, traitIter.Label())
			}

			data, _ := format.Node(item.Source())

			code := string(data)
			catalog = append(catalog, CatalogItem{
				Package:      pkg,
				Dependencies: deps,
				Code:         code,
				Name:         fieldIter.Label(),
				Traits:       traits,
			})
			continue
		}
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}

	log.Info(string(data))

	return nil
}
