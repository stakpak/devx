package drivers

import (
	"path"
	"strings"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
)

type Driver interface {
	match(resource cue.Value) bool
	ApplyAll(stack *stack.Stack) error
}

// TODO we need to decompose this into DI pattern
func NewDriversMap(environment string, config map[string]map[string]string) map[string]Driver {

	composePath := path.Join("build", environment, "compose")
	composeFile := "docker-compose.yml"
	if composeConfig, ok := config["compose"]; ok {
		if output, ok := composeConfig["output"]; ok {
			if strings.HasSuffix(output, ".yml") || strings.HasSuffix(output, ".yaml") {
				composePath, composeFile = path.Split(output)
			} else {
				composePath = output
			}
		}
	}

	terraformPath := path.Join("build", environment, "terraform")
	if terraformConfig, ok := config["terraform"]; ok {
		if output, ok := terraformConfig["output"]; ok {
			terraformPath = output
		}
	}

	kubernetesPath := path.Join("build", environment, "kubernetes")
	if kubernetesConfig, ok := config["kubernetes"]; ok {
		if output, ok := kubernetesConfig["output"]; ok {
			kubernetesPath = output
		}
	}

	gitlabPath := path.Join("build", environment, "gitlab")
	gitlabFile := ".gitlab-ci.yml"
	if gitlabConfig, ok := config["gitlab"]; ok {
		if output, ok := gitlabConfig["output"]; ok {
			if strings.HasSuffix(output, ".yml") || strings.HasSuffix(output, ".yaml") {
				gitlabPath, gitlabFile = path.Split(output)
			} else {
				gitlabPath = output
			}
		}
	}

	githubPath := path.Join("build", environment, "github")
	if githubConfig, ok := config["github"]; ok {
		if output, ok := githubConfig["output"]; ok {
			githubPath = output
		}
	}

	return map[string]Driver{
		"compose": &ComposeDriver{
			Path: composePath,
			File: composeFile,
		},
		"terraform": &TerraformDriver{
			Path: terraformPath,
		},
		"kubernetes": &KubernetesDriver{
			Path: kubernetesPath,
		},
		"gitlab": &GitlabDriver{
			Path: gitlabPath,
			File: gitlabFile,
		},
		"github": &GitHubDriver{
			Path: githubPath,
		},
	}
}
