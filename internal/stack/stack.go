package stack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"cuelang.org/go/cue"
	cueflow "cuelang.org/go/tools/flow"
	"devopzilla.com/guku/internal/utils"
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
)

type Stack struct {
	ID           string
	DepIDs       []string
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
	Git         *GitData               `json:"git,omitempty"`
}
type GitData struct {
	Commit  string    `json:"commit"`
	Branch  string    `json:"branch"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Time    time.Time `json:"time"`
	IsClean bool      `json:"clean"`
	Parents []string  `json:"parents"`
}
type Reference struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func (s *Stack) SendBuild(configDir string, telemetryEndpoint string, environment string) error {
	build := BuildData{
		Stack:       s.ID,
		Identity:    "",
		Result:      s.GetComponents(),
		Imports:     s.DepIDs,
		References:  s.GetReferences(),
		Environment: environment,
		Git:         nil,
	}

	repo, err := git.PlainOpen(configDir)
	if err != git.ErrRepositoryNotExists {
		if err != nil {
			return err
		}

		ref, err := repo.Head()
		if err != nil {
			return err
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		parents := []string{}
		for _, p := range commit.ParentHashes {
			parents = append(parents, p.String())
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}
		status, err := w.Status()
		if err != nil {
			return err
		}

		isClean := status.IsClean()

		build.Git = &GitData{
			IsClean: isClean,
			Commit:  commit.ID().String(),
			Message: strings.TrimSpace(commit.Message),
			Author:  commit.Author.String(),
			Time:    commit.Author.When,
			Parents: parents,
			Branch:  ref.Name().Short(),
		}
	}

	data, err := json.Marshal(build)
	if err != nil {
		return err
	}
	log.Debug("Stack ID: ", s.ID)

	url, _ := url.Parse(telemetryEndpoint)
	url.Path = path.Join(url.Path, "api", "builds")
	log.Debug("URL: ", url)

	request, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	log.Debug("Response Status: ", response.Status)
	log.Debug("Response Headers: ", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	log.Debug("Response Body: ", string(body))

	return nil
}

func (s *Stack) GetReferences() map[string][]Reference {
	componentIter, _ := s.components.Fields()

	refMap := map[string][]Reference{}
	for componentIter.Next() {
		refs := removeDuplicates(GetRef(componentIter.Value()))
		if len(refs) > 0 {
			refMap[componentIter.Label()] = refs
		}
	}

	return refMap
}

func GetRef(v cue.Value) []Reference {
	refs := []Reference{}

	v.Walk(func(val cue.Value) bool {
		if strings.HasSuffix(val.Path().String(), "$resources") {
			return false
		}
		op, vals := val.Expr()
		switch op {
		case cue.AndOp, cue.InterpolationOp:
			for _, value := range vals {
				refs = append(refs, GetRef(value)...)
			}
			// return false
		case cue.SelectorOp:
			_, structPath := vals[0].ReferencePath()
			if len(structPath.Selectors()) > 1 {
				refs = append(refs, Reference{
					Source: fmt.Sprintf("%s.%s", getPathSuffix(structPath), vals[1]),
					Target: getPathSuffix(val.Path()),
				})
			}
			// return false
		case cue.IndexOp:
			_, structPath := vals[0].ReferencePath()
			if len(structPath.Selectors()) > 1 {
				refs = append(refs, Reference{
					Source: fmt.Sprintf("%s[%s]", getPathSuffix(structPath), vals[1]),
					Target: getPathSuffix(val.Path()),
				})
			}
			// return false
		}
		return true
	}, nil)

	return refs
}

func getPathSuffix(p cue.Path) string {
	sel := p.Selectors()
	return cue.MakePath(sel[2:]...).String()
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
