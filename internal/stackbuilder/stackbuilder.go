package stackbuilder

import (
	"context"
	"fmt"
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
	DriverConfig         map[string]map[string]string
	AdditionalComponents *cue.Value
	Flows                []*Flow
}

func NewEnvironments(value cue.Value) (Environments, error) {
	environments := map[string]*StackBuilder{}

	envIter, err := value.Fields()
	if err != nil {
		return environments, err
	}

	for envIter.Next() {
		name := utils.GetLastPathFragment(envIter.Value())
		environments[name], err = NewStackBuilder(envIter.Value())
		if err != nil {
			return environments, err
		}
	}

	return environments, nil
}

func NewStackBuilder(value cue.Value) (*StackBuilder, error) {
	flows := value.LookupPath(cue.ParsePath("flows"))
	if flows.Err() != nil {
		return nil, flows.Err()
	}

	var additionalComponents *cue.Value
	additionalComponentsValue := value.LookupPath(cue.ParsePath("additionalComponents"))
	if additionalComponentsValue.Exists() {
		additionalComponents = &additionalComponentsValue
	}

	driverConfig := make(map[string]map[string]string)
	driverConfigValue := value.LookupPath(cue.ParsePath("drivers"))
	if driverConfigValue.Exists() {
		driverIter, err := driverConfigValue.Fields()
		if err != nil {
			return nil, err
		}
		for driverIter.Next() {
			driverConfig[driverIter.Label()] = make(map[string]string)
			configIter, err := driverIter.Value().Fields()
			if err != nil {
				return nil, err
			}
			for configIter.Next() {
				value, err := configIter.Value().String()
				if err != nil {
					return nil, err
				}
				driverConfig[driverIter.Label()][configIter.Label()] = value
			}
		}
	}

	stackBuilder := StackBuilder{
		DriverConfig:         driverConfig,
		AdditionalComponents: additionalComponents,
		Flows:                make([]*Flow, 0),
	}
	flowIter, _ := flows.List()
	for flowIter.Next() {
		flow, err := NewFlow(flowIter.Value())
		if err != nil {
			return nil, err
		}
		stackBuilder.Flows = append(stackBuilder.Flows, flow)
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
			c, _ := stack.GetComponent(componentId)
			err := c.Validate(cue.Concrete(true), cue.All())
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
