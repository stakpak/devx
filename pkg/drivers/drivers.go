package drivers

import (
	"cuelang.org/go/cue"
	"github.com/stakpak/devx/pkg/stack"
	"github.com/stakpak/devx/pkg/stackbuilder"
)

type Driver interface {
	match(resource cue.Value) bool
	ApplyAll(stack *stack.Stack, stdout bool) error
}

// TODO we need to decompose this into DI pattern
func NewDriversMap(environment string, config map[string]stackbuilder.DriverConfig) map[string]Driver {
	return map[string]Driver{
		"compose": &ComposeDriver{
			Config: config["compose"],
		},
		"terraform": &TerraformDriver{
			Config: config["terraform"],
		},
		"kubernetes": &KubernetesDriver{
			Config: config["kubernetes"],
		},
		"gitlab": &GitlabDriver{
			Config: config["gitlab"],
		},
		"github": &GitHubDriver{
			Config: config["github"],
		},
		"yaml": &YAMLDriver{
			Config: config["yaml"],
		},
		"json": &JSONDriver{
			Config: config["json"],
		},
	}
}
