package project

import (
	"fmt"
	"os"
	"path"
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

func Generate(configDir string) error {
	appPath := path.Join(configDir, "stack.cue")

	os.WriteFile(appPath, []byte(`package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		app: {
			v1.#Component
			traits.#Workload
			traits.#Exposable
			image: "app:v1"
			ports: [
				{
					port: 8080
				},
			]
			env: {
				PGDB_URL: db.url
			}
			volumes: [
				{
					source: "bla"
					target: "/tmp/bla"
				},
			]
		}
		db: {
			v1.#Component
			traits.#Postgres
			version:    "12.1"
			persistent: true
		}
	}
}
	`), 0700)

	builderPath := path.Join(configDir, "builder.cue")
	os.WriteFile(builderPath, []byte(`package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	"guku.io/devx/v1/transformers/compose"
)


builders: v1.#StackBuilder & {
	dev: {
		additionalComponents: {
			observedb: {
				v1.#Component
				traits.#Postgres
				version:    "12.1"
				persistent: true
			}
		}
		flows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#AddComposePostgres & {},
				]
			},
		]
	}
}	
	`), 0700)

	return nil
}
