package drivers

import (
	"path"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
)

type Driver interface {
	match(resource cue.Value) bool
	ApplyAll(stack *stack.Stack) error
}

// TODO we need to decompose this into DI pattern
func NewDriversMap(environment string) map[string]Driver {
	return map[string]Driver{
		"compose": &ComposeDriver{
			Path: path.Join("build", environment, "compose"),
		},
		"terraform": &TerraformDriver{
			Path: path.Join("build", environment, "terraform"),
		},
		"kubernetes": &KubernetesDriver{
			Path: path.Join("build", environment, "kubernetes"),
		},
		"gitlab": &GitlabDriver{
			Path: path.Join("build", environment, "gitlab"),
		},
	}
}
