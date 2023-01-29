package stackbuilder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/utils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type Environments = map[string]*StackBuilder

type StackBuilder struct {
	DriverConfig         map[string]DriverConfig
	AdditionalComponents *cue.Value
	Flows                []*Flow
}
type DriverConfig struct {
	Output DriverOutput `json:"output"`
}
type DriverOutput struct {
	Dir  string `json:"dir"`
	File string `json:"file"`
}

func NewEnvironments(value cue.Value) (Environments, error) {
	environments := map[string]*StackBuilder{}

	envIter, err := value.Fields()
	if err != nil {
		return environments, err
	}

	for envIter.Next() {
		name := utils.GetLastPathFragment(envIter.Value())
		environments[name], err = NewStackBuilder(name, envIter.Value())
		if err != nil {
			return environments, err
		}
	}

	return environments, nil
}

func NewStackBuilder(environment string, value cue.Value) (*StackBuilder, error) {
	isV2Builder := false
	envName := value.LookupPath(cue.ParsePath("environment"))
	if envName.Exists() {
		isV2Builder = true
	}

	flows := value.LookupPath(cue.ParsePath("flows"))
	if flows.Err() != nil {
		return nil, flows.Err()
	}

	var additionalComponents *cue.Value
	additionalComponentsPath := "additionalComponents"
	if isV2Builder {
		additionalComponentsPath = "components"
	}
	additionalComponentsValue := value.LookupPath(cue.ParsePath(additionalComponentsPath))
	if additionalComponentsValue.Exists() {
		additionalComponents = &additionalComponentsValue
	}

	driverConfig := map[string]DriverConfig{}
	driverConfigValue := value.LookupPath(cue.ParsePath("drivers"))
	if driverConfigValue.Exists() {
		driverIter, err := driverConfigValue.Fields()
		if err != nil {
			return nil, err
		}
		for driverIter.Next() {
			driverConfig[driverIter.Label()] = DriverConfig{}
			configIter, err := driverIter.Value().Fields()
			if err != nil {
				return nil, err
			}
			for configIter.Next() {
				switch configIter.Value().Kind() {
				case cue.StringKind:
					value, err := configIter.Value().String()
					if err != nil {
						return nil, err
					}
					dir, file := filepath.Split(value)
					if filepath.Ext(file) == "" {
						dir, file = value, ""
					}
					driverConfig[driverIter.Label()] = DriverConfig{
						Output: DriverOutput{
							Dir:  dir,
							File: file,
						},
					}
				case cue.StructKind:
					dirValue := configIter.Value().LookupPath(cue.ParsePath("dir"))
					fileValue := configIter.Value().LookupPath(cue.ParsePath("file"))

					var dirPaths []string
					var file string

					err := dirValue.Decode(&dirPaths)
					if err != nil {
						return nil, err
					}
					err = fileValue.Decode(&file)
					if err != nil {
						return nil, err
					}

					driverConfig[driverIter.Label()] = DriverConfig{
						Output: DriverOutput{
							Dir:  filepath.Join(dirPaths...),
							File: file,
						},
					}
				}
			}
		}
	}

	if !isV2Builder {
		driverDefaults := map[string]string{
			"compose":    "docker-compose.yml",
			"gitlab":     ".gitlab-ci.yml",
			"terraform":  "generated.tf.json",
			"github":     "",
			"kubernetes": "",
		}
		for name, file := range driverDefaults {
			config, ok := driverConfig[name]
			if !ok {
				driverConfig[name] = DriverConfig{
					Output: DriverOutput{"", ""},
				}
			}
			if config.Output.Dir == "" && config.Output.File == "" {
				config.Output.Dir = filepath.Join("build", environment, name)
			}
			if config.Output.File == "" {
				config.Output.File = file
			}
			driverConfig[name] = config
		}
	}

	stackBuilder := StackBuilder{
		DriverConfig:         driverConfig,
		AdditionalComponents: additionalComponents,
		Flows:                make([]*Flow, 0),
	}

	if isV2Builder {
		flowIter, err := flows.Fields()
		if err != nil {
			return nil, err
		}
		for flowIter.Next() {
			flow, err := NewFlow(flowIter.Value())
			if err != nil {
				return nil, err
			}
			stackBuilder.Flows = append(stackBuilder.Flows, flow)
		}
	} else {
		flowIter, err := flows.List()
		if err != nil {
			return nil, err
		}
		for flowIter.Next() {
			flow, err := NewFlow(flowIter.Value())
			if err != nil {
				return nil, err
			}
			stackBuilder.Flows = append(stackBuilder.Flows, flow)
		}
	}

	return &stackBuilder, nil
}

func (sb *StackBuilder) TransformStack(ctx context.Context, stack *stack.Stack) error {
	if sb.AdditionalComponents != nil {
		stack.AddComponents(*sb.AdditionalComponents)
	}
	orderedTasks := stack.GetTasks()

	total := 0
	for _, flow := range sb.Flows {
		total += len(orderedTasks) * len(flow.pipeline)
	}
	bar := progressbar.Default(int64(total), "🏭 Transforming stack")
	defer bar.Finish()
	for _, componentId := range orderedTasks {
		component, err := stack.GetComponent(componentId)
		if err != nil {
			return err
		}
		for _, flow := range sb.Flows {
			component, err = flow.Run(ctx, stack, componentId, component)
			if err != nil {
				return err
			}
			if !stack.HasConcreteResourceDrivers(component) {
				return fmt.Errorf(
					"component %s resources do not have concrete drivers",
					componentId,
				)
			}
			bar.Add(len(flow.pipeline))
		}
		if !stack.IsConcreteComponent(component) {
			err := component.Validate(cue.Concrete(true), cue.All())
			log.Debugln(component)
			return fmt.Errorf("component %s is not concrete after transformation:\n%s", componentId, errors.Details(err, nil))
		}
		stack.UpdateComponent(componentId, component)
	}
	return nil
}

func CheckTraitFulfillment(builders Environments, stack *stack.Stack) error {
	compIter, err := stack.GetComponents().Fields()
	if err != nil {
		return err
	}
	// component -> trait -> env -> handled
	compFlowMap := map[string]map[string]map[string]bool{}
	unmatched := []string{}
	for compIter.Next() {
		component := compIter.Label()
		compFlowMap[component] = map[string]map[string]bool{}
		traitIter, _ := compIter.Value().LookupPath(cue.ParsePath("$metadata.traits")).Fields()
		for traitIter.Next() {
			compFlowMap[component][traitIter.Label()] = map[string]bool{}
		}
		for env, builder := range builders {
			for trait := range compFlowMap[component] {
				compFlowMap[component][trait][env] = false
			}
			for _, flow := range builder.Flows {
				if flow.Match(compIter.Value()) {
					for _, trait := range flow.GetHandledTraits() {
						compFlowMap[component][trait][env] = true
					}
				}
			}
			for trait := range compFlowMap[component] {
				if !compFlowMap[component][trait][env] {
					unmatched = append(unmatched, fmt.Sprintf("%s/#%s in %s", component, trait, env))
				}
			}
		}
	}
	if len(unmatched) > 0 {
		return fmt.Errorf("trait is not fulfilled by any flow: \n  %s", strings.Join(unmatched, "\n  "))
	}
	return nil
}
