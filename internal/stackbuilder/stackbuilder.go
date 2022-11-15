package stackbuilder

import "cuelang.org/go/cue"

type StackBuilder struct {
	// AdditionalComponents *cue.Value
	Flows []*Flow
}

func NewStackBuilder(value cue.Value) (*StackBuilder, error) {
	flows := value.LookupPath(cue.ParsePath("flows"))
	if flows.Err() != nil {
		return nil, flows.Err()
	}

	stackBuilder := StackBuilder{
		// AdditionalComponents: nil,
		Flows: make([]*Flow, 0),
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

func (sb *StackBuilder) TransformComponent(component cue.Value) (cue.Value, error) {
	temp := component

	// for flow := range sb.Flows {

	// }

	return temp, nil
}
