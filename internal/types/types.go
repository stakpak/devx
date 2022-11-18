package types

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"devopzilla.com/guku/internal/utils"
)

func Validate(configDir string) error {
	value := utils.LoadProject(configDir)
	err := value.Validate()
	if err == nil {
		fmt.Println("Looks good 👀👌")
	}
	return err
}

func Discover(configDir string, showTraitDef bool) error {
	instances := utils.LoadInstances(configDir)

	deps := instances[0].Dependencies()

	for _, dep := range deps {
		if strings.Contains(dep.ID(), "traits") {
			ctx := cuecontext.New()
			value := ctx.BuildInstance(dep)

			fieldIter, _ := value.Fields(cue.Definitions(true), cue.Docs(true))

			fmt.Printf("📜 %s\n", dep.ID())
			for fieldIter.Next() {
				traits := fieldIter.Value().LookupPath(cue.ParsePath("$metadata.traits"))
				if traits.Exists() && traits.IsConcrete() {
					fmt.Printf("traits.%s", fieldIter.Selector().String())
					if utils.HasComments(fieldIter.Value()) {
						fmt.Printf("\t%s", utils.GetComments(fieldIter.Value()))
					}
					fmt.Println()
					if showTraitDef {
						fmt.Println(fieldIter.Value())
						fmt.Println()
					}
				}
			}
			fmt.Println()
		}
	}

	return nil
}
