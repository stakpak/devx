package stack

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	cueflow "cuelang.org/go/tools/flow"
	"github.com/devopzilla/devx/pkg/auth"
	"github.com/devopzilla/devx/pkg/gitrepo"
	"github.com/devopzilla/devx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Stack struct {
	ID           string
	DepIDs       []string
	BuildSource  string
	components   cue.Value
	tasks        []string
	dependencies map[string][]string
}

func NewStack(value cue.Value, stackId string, depIds []string) (*Stack, error) {
	components := value.LookupPath(cue.ParsePath("components"))
	if components.Err() != nil {
		return nil, components.Err()
	}

	stack := &Stack{
		ID:           stackId,
		DepIDs:       depIds,
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
		id := utils.GetLastPathFragment(task.Value())
		parents := task.Dependencies()
		stack.dependencies[id] = make([]string, 0)
		for _, parent := range parents {
			parentId := utils.GetLastPathFragment(parent.Value())
			stack.addDependency(id, parentId)
		}
	}

	stack.tasks = computeOrderedTasks(stack)

	return stack, nil
}

func (s *Stack) Print() {
	log.Info(s.components)
}

func (s *Stack) GetDependencies(id string) ([]string, error) {
	val, ok := s.dependencies[id]
	if !ok {
		return nil, fmt.Errorf("Id %s not found in dependencies map", id)
	}
	return val, nil
}

func (s *Stack) UpdateComponent(id string, value cue.Value) error {
	// s.components.
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

func (s *Stack) IsConcreteComponent(component cue.Value) bool {
	err := component.Validate(cue.Concrete(true))
	return err == nil
}

func (s *Stack) HasConcreteResourceDrivers(component cue.Value) bool {
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
		id := utils.GetLastPathFragment(task.Value())
		parents := task.Dependencies()
		s.dependencies[id] = make([]string, 0)
		for _, parent := range parents {
			parentId := utils.GetLastPathFragment(parent.Value())
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

func (s *Stack) GetComponents() cue.Value {
	return s.components
}

type BuildData struct {
	Stack       string                 `json:"stack"`
	Identity    string                 `json:"identity,omitempty"`
	Result      cue.Value              `json:"result"`
	Imports     []string               `json:"imports"`
	References  map[string][]Reference `json:"references"`
	Environment string                 `json:"environment"`
	Git         *gitrepo.GitData       `json:"git,omitempty"`
	Error       *string                `json:"error"`
	Source      string                 `json:"source"`
}
type Reference struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func (s *Stack) SendBuild(configDir string, server auth.ServerConfig, environment string, buildError *string) (string, error) {
	build := BuildData{
		Stack:       s.ID,
		Identity:    "",
		Imports:     s.DepIDs,
		Environment: environment,
		Git:         nil,
		Error:       buildError,
		Source:      s.BuildSource,
	}

	if buildError == nil {
		build.Result = s.GetComponents()
		build.References = s.GetReferences()
	} else {
		build.Result = s.GetContext().CompileString("{}")
		build.References = map[string][]Reference{}
	}

	gitData, err := gitrepo.GetGitData(configDir)
	if err != nil {
		return "", err
	}
	build.Git = gitData

	data, err := utils.SendData(server, "builds", &build)
	if err != nil {
		return "", err
	}

	buildResponse := make(map[string]string)
	err = json.Unmarshal(data, &buildResponse)
	if err != nil {
		return "", err
	}

	return buildResponse["id"], nil
}

func (s *Stack) GetReferences() map[string][]Reference {
	refMap := map[string][]Reference{}

	refs := removeDuplicates(GetRef(s.components.Syntax()))
	for _, r := range refs {
		if r.Target == "" {
			continue
		}
		parts := strings.SplitN(r.Target, ".", 2)
		refMap[parts[0]] = append(refMap[parts[0]], r)
	}

	return refMap
}

func GetRef(node ast.Node) []Reference {
	refs := []Reference{}

	if node == nil {
		return refs
	}

	astutil.Apply(node, func(c astutil.Cursor) bool {
		path := GetPath(c)
		if strings.Contains(path, "#") || strings.Contains(path, "$") {
			return false
		}

		switch n := c.Node().(type) {
		case *ast.SelectorExpr:
			sourcePath := GetName(n)
			if strings.Contains(sourcePath, "#") || strings.Contains(sourcePath, "?") {
				return false
			}
			refs = append(refs, Reference{
				Source: sourcePath,
				Target: path,
			})
			return false
		case *ast.IndexExpr:
			sourcePath := GetName(n)
			if strings.Contains(sourcePath, "#") || strings.Contains(sourcePath, "?") {
				return false
			}
			refs = append(refs, Reference{
				Source: sourcePath,
				Target: path,
			})
			return false
		}

		return true
	}, nil)

	return refs
}

func GetPath(c astutil.Cursor) string {
	if c == nil {
		return ""
	}

	path := GetPath(c.Parent())
	if c.Parent() != nil {
		switch p := c.Parent().Node().(type) {
		case *ast.ListLit:
			for i, e := range p.Elts {
				if e == c.Node() {
					path = path + fmt.Sprintf("[%d]", i)
					break
				}
			}
		}
	}

	switch n := c.Node().(type) {
	case *ast.Field:
		path = path + "." + GetName(n.Label)
	case *ast.Ident:
		path = path + "." + GetName(n)
	}
	return strings.TrimPrefix(path, ".")
}

func GetName(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.Field:
		return GetName(n.Label)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", GetName(n.X), GetName(n.Sel))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", GetName(n.X), GetName(n.Index))
	case *ast.BasicLit:
		return n.Value
	}

	return "?"
}

func removeDuplicates(refs []Reference) []Reference {
	allKeys := make(map[string]bool)
	list := []Reference{}
	for _, item := range refs {
		key := fmt.Sprintf(item.Source, item.Target)
		if _, value := allKeys[key]; !value {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}
