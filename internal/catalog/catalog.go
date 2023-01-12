package catalog

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	log "github.com/sirupsen/logrus"

	"devopzilla.com/guku/internal/gitrepo"
	"devopzilla.com/guku/internal/utils"
)

type CatalogItem struct {
	Dependencies []string               `json:"dependencies"`
	Package      string                 `json:"package"`
	Name         string                 `json:"name"`
	Source       string                 `json:"source"`
	Git          Git                    `json:"git"`
	Metadata     map[string]interface{} `json:"metadata"`
}
type Git struct {
	gitrepo.ProjectGitData
	gitrepo.GitData
}

func Publish(gitDir string, configDir string, telemetry string) error {
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

	projectGitData, err := gitrepo.GetProjectGitData(gitDir)
	if err != nil {
		return nil
	}
	if projectGitData == nil {
		return fmt.Errorf("git is not initialized, cannot publish a catalog without version control")
	}
	gitData, err := gitrepo.GetGitData(gitDir)
	if err != nil {
		return nil
	}
	if gitData == nil {
		return fmt.Errorf("git is not initialized, cannot publish a catalog without version control")
	}

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

		traitMeta := metadata.LookupPath(cue.ParsePath("traits"))
		if traitMeta.Exists() {
			traits := []string{}
			traitIter, _ := traitMeta.Fields()
			for traitIter.Next() {
				traits = append(traits, traitIter.Label())
			}

			data, _ := format.Node(item.Source())
			catalogItem := CatalogItem{
				Package:      pkg,
				Dependencies: deps,
				Source:       strings.TrimSpace(string(data)),
				Name:         fieldIter.Label(),
				Git: Git{
					*projectGitData,
					*gitData,
				},
				Metadata: map[string]interface{}{
					"traits": traits,
					"type":   "Trait",
				},
			}
			err = publishCatalogItem(telemetry, &catalogItem)
			if err != nil {
				return err
			}
			continue
		}

		stackMeta := metadata.LookupPath(cue.ParsePath("stack"))
		if stackMeta.Exists() {
			componentsMeta := map[string]interface{}{}

			components := item.LookupPath(cue.ParsePath("components"))
			componentIter, _ := components.Fields()
			for componentIter.Next() {
				traits := []string{}
				traitIter, _ := componentIter.Value().LookupPath(cue.ParsePath("$metadata.traits")).Fields()
				for traitIter.Next() {
					traits = append(traits, traitIter.Label())
				}
				componentsMeta[componentIter.Label()] = map[string]interface{}{
					"traits": traits,
				}
			}

			data, _ := format.Node(item.Source())
			catalogItem := CatalogItem{
				Package:      pkg,
				Dependencies: deps,
				Source:       strings.TrimSpace(string(data)),
				Name:         fieldIter.Label(),
				Git: Git{
					*projectGitData,
					*gitData,
				},
				Metadata: map[string]interface{}{
					"components": componentsMeta,
					"type":       "Stack",
				},
			}
			err = publishCatalogItem(telemetry, &catalogItem)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func publishCatalogItem(telemetry string, catalogItem *CatalogItem) error {
	data, err := utils.SendTelemtry(telemetry, "catalog", catalogItem)
	if err != nil {
		log.Debug(string(data))
		return err
	}

	log.Infof("🚀 Published %s %s successfully", catalogItem.Package, catalogItem.Name)

	return nil
}
