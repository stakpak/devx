package stack

import (
	"fmt"
	"sort"

	"cuelang.org/go/cue"
	cueflow "cuelang.org/go/tools/flow"
	"devopzilla.com/guku/internal/utils"
)

type Stack struct {
	components   cue.Value
	tasks        []string
	dependencies map[string][]string
}

func NewStack(value cue.Value) (*Stack, error) {
	components := value.LookupPath(cue.ParsePath("components"))
	if components.Err() != nil {
		return nil, components.Err()
	}

	stack := &Stack{
		components:   components,
		dependencies: make(map[string][]string),
	}

	cfg := &cueflow.Config{
		FindHiddenTasks: true,
		Root:            cue.ParsePath(""),
	}

	flow := cueflow.New(
		cfg,
		components,
		taskFunc,
	)

	tasks := flow.Tasks()
	for _, task := range tasks {
		id := utils.GetLastPathFragement(task.Value())
		parents := task.Dependencies()
		stack.dependencies[id] = make([]string, 0)
		for _, parent := range parents {
			parentId := utils.GetLastPathFragement(parent.Value())
			stack.addDependency(id, parentId)
		}
	}

	stack.tasks = computeOrderedTasks(stack)

	return stack, nil
}

func (s *Stack) Print() {
	fmt.Println(s.components)
}

func (s *Stack) GetDependencies(id string) ([]string, error) {
	val, ok := s.dependencies[id]
	if !ok {
		return nil, fmt.Errorf("Id %s not found in dependencies map", id)
	}
	return val, nil
}

func (s *Stack) UpdateComponent(id string, value cue.Value) error {
	s.components = s.components.FillPath(cue.ParsePath(id), value)
	if s.components.Err() != nil {
		return s.components.Err()
	}
	return nil
}

func (s *Stack) GetComponent(id string) (cue.Value, error) {
	result := s.components.LookupPath(cue.ParsePath(id))
	return result, result.Err()
}

func (s *Stack) IsConcreteComponent(id string) bool {
	component := s.components.LookupPath(cue.ParsePath(id))
	err := component.Validate(cue.Concrete(true))
	return err == nil
}

func (s *Stack) HasConcreteResourceDrivers(id string) bool {
	component := s.components.LookupPath(cue.ParsePath(id))
	resources := component.LookupPath(cue.ParsePath("$resource"))

	if resources.Exists() {
		resourceIter, _ := resources.Fields()
		for resourceIter.Next() {
			driver := resourceIter.Value().LookupPath(cue.ParsePath("$metadata.labels.driver"))
			if !driver.Exists() {
				return false
			}
			err := driver.Validate(cue.Concrete(true))
			if err != nil {
				return false
			}
		}
	}
	return true
}

func (s *Stack) GetContext() *cue.Context {
	return s.components.Context()
}

func (s *Stack) GetTasks() []string {
	return s.tasks
}

func (s *Stack) AddComponents(value cue.Value) {
	s.components = s.components.FillPath(cue.ParsePath(""), value)

	cfg := &cueflow.Config{
		FindHiddenTasks: true,
		Root:            cue.ParsePath(""),
	}

	flow := cueflow.New(
		cfg,
		s.components,
		taskFunc,
	)

	tasks := flow.Tasks()
	for _, task := range tasks {
		id := utils.GetLastPathFragement(task.Value())
		parents := task.Dependencies()
		s.dependencies[id] = make([]string, 0)
		for _, parent := range parents {
			parentId := utils.GetLastPathFragement(parent.Value())
			s.addDependency(id, parentId)
		}
	}

	s.tasks = computeOrderedTasks(s)
}

func (s *Stack) addDependency(id string, depId string) {
	s.dependencies[id] = append(s.dependencies[id], depId)
}

// cue flow already checks for cycles, so we don't have to
func computeOrderedTasks(s *Stack) []string {

	result := make([]string, 0)

	// initialize data structures
	visited := make(map[string]bool)
	stack := make([]string, 0)
	for id := range s.dependencies {
		stack = append(stack, id)
		visited[id] = false
	}

	// make scheduling deterministic
	sort.Strings(stack)

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		if isVisited, _ := visited[current]; isVisited {
			stack = stack[:len(stack)-1]
			continue
		}

		parents, _ := s.GetDependencies(current)
		isReady := true
		for _, parent := range parents {
			if isVisited, _ := visited[parent]; !isVisited {
				isReady = false
				stack = append(stack, parent)
			}
		}
		if !isReady {
			continue
		}

		visited[current] = true
		result = append(result, current)
		stack = stack[:len(stack)-1]
	}

	return result
}

func taskFunc(v cue.Value) (cueflow.Runner, error) {
	idPath := cue.ParsePath("$metadata.id")
	componentId := v.LookupPath(idPath)
	if !componentId.Exists() {
		// Not a task
		return nil, nil
	}

	return cueflow.RunnerFunc(func(t *cueflow.Task) error {
		return nil
	}), nil
}
