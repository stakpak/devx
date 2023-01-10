package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
	"github.com/go-git/go-billy/v5"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ContextKey string

const (
	ConfigDirKey ContextKey = "configDir"
	DryRunKey    ContextKey = "dryRun"
)

func LoadInstances(configDir string, overlays *map[string]string) []*build.Instance {
	sourceOverlays := map[string]cueload.Source{}

	if overlays != nil {
		for file, content := range *overlays {
			absConfigDir, _ := filepath.Abs(configDir)
			filePath := path.Join(absConfigDir, file)
			sourceOverlays[filePath] = cueload.FromString(content)
		}
	}

	buildConfig := &cueload.Config{
		Dir:     configDir,
		Overlay: sourceOverlays,
	}
	return cueload.Instances([]string{}, buildConfig)
}

func LoadProject(configDir string, overlays *map[string]string) (cue.Value, string, []string) {
	instances := LoadInstances(configDir, overlays)

	ctx := cuecontext.New()
	stackID := strings.Split(instances[0].ID(), ":")[0]

	return ctx.BuildInstance(instances[0]), stackID, instances[0].Deps
}

func GetLastPathFragment(value cue.Value) string {
	selector := value.Path().Selectors()
	return selector[len(selector)-1].String()
}

func GetComments(value cue.Value) string {
	comments := value.Doc()
	result := ""
	for _, comment := range comments {
		result += comment.Text()
	}
	return strings.ReplaceAll(result, "\n", " ")
}

func HasComments(value cue.Value) bool {
	comments := value.Doc()

	return len(comments) > 0
}

func Walk(v cue.Value, before func(cue.Value) bool, after func(cue.Value)) {
	switch v.Kind() {
	case cue.StructKind:
		if before != nil && !before(v) {
			return
		}
		fieldIter, _ := v.Fields(cue.All())
		for fieldIter.Next() {
			Walk(fieldIter.Value(), before, after)
		}
	case cue.ListKind:
		if before != nil && !before(v) {
			return
		}
		valueIter, _ := v.List()
		for valueIter.Next() {
			Walk(valueIter.Value(), before, after)
		}
	default:
		if before != nil {
			before(v)
		}
	}
	if after != nil {
		after(v)
	}
}

func RemoveMeta(value cue.Value) (cue.Value, error) {
	result := value.Context().CompileString("_")

	iter, err := value.Fields()
	if err != nil {
		return result, err
	}

	for iter.Next() {
		v := iter.Value()
		label, _ := v.Label()
		if !strings.HasPrefix(label, "$") {
			result = result.FillPath(cue.ParsePath(label), v)
		}
	}

	return result, nil
}

func FsWalk(fs billy.Filesystem, filePath string, process func(p string, content []byte) error) error {
	file, err := fs.Lstat(filePath)
	if err != nil {
		return err
	}

	if file.IsDir() {
		files, err := fs.ReadDir(filePath)
		if err != nil {
			return err
		}

		for _, f := range files {
			childPath := path.Join(filePath, f.Name())
			err := FsWalk(fs, childPath, process)
			if err != nil {
				return err
			}
		}
	} else {
		content, err := fs.Open(filePath)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(bufio.NewReader(content))
		if err != nil {
			return err
		}

		return process(filePath, data)
	}

	return nil
}

func IsReference(value cue.Value) bool {
	_, vs := value.Expr()
	for _, v := range vs {
		op, _ := v.Expr()
		if op.String() == "." {
			return true
		}
	}
	return false
}

func GetOverlays(configDir string) (map[string]string, error) {
	overlays := map[string]string{}

	files, err := os.ReadDir(configDir)
	if err != nil {
		return overlays, err
	}

	for _, f := range files {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".devx.yaml") || strings.HasSuffix(f.Name(), ".devx.yml")) {
			file, err := os.ReadFile(path.Join(configDir, f.Name()))
			if err != nil {
				return overlays, err
			}

			var n yaml.Node
			err = yaml.Unmarshal(file, &n)
			if err != nil {
				return overlays, err
			}

			overlays[f.Name()+".cue"] = BuildCUEFile("", &n)
		}
	}

	return overlays, nil
}

func BuildCUEFile(content string, n *yaml.Node) string {
	newContent := content

	switch n.Kind {
	case yaml.DocumentNode:
		newContent += "package main\n"
		for _, child := range n.Content {
			newContent = BuildCUEFile(newContent, child)
		}
	case yaml.SequenceNode:
		newContent = fmt.Sprintf("%s [\n", newContent)
		for _, child := range n.Content {
			newContent = fmt.Sprintf("%s,\n", BuildCUEFile(newContent, child))
		}
		newContent = fmt.Sprintf("%s\n]\n", newContent)
	case yaml.MappingNode:
		addBrace := false
		for i := 0; i < len(n.Content); i += 2 {
			name := n.Content[i].Value

			if name == "import" && newContent == "package main\n" {
				for j := 0; j < len(n.Content[i+1].Content); j += 2 {
					newContent = fmt.Sprintf(
						"%simport %s \"%s\"\n",
						newContent,
						n.Content[i+1].Content[j].Value,
						n.Content[i+1].Content[j+1].Value,
					)
				}
				continue
			}

			if name == "$schema" {
				schmaValues := []string{}
				for _, child := range n.Content[i+1].Content {
					schmaValues = append(schmaValues, child.Value)
				}
				schema := strings.Join(schmaValues, " & ")
				if !addBrace {
					newContent = fmt.Sprintf("%s%s & {\n", newContent, schema)
					addBrace = true
				}
				continue
			}

			if name == "$traits" {
				if !addBrace {
					newContent = fmt.Sprintf("%s {\n", newContent)
					addBrace = true
				}
				for _, child := range n.Content[i+1].Content {
					newContent = fmt.Sprintf("%s%s\n", newContent, child.Value)
				}
				continue
			}

			if !addBrace {
				newContent = fmt.Sprintf("%s {\n", newContent)
				addBrace = true
			}

			child := n.Content[i+1]
			newContent = fmt.Sprintf("%s%s: ", newContent, name)
			newContent = fmt.Sprintf("%s\n", BuildCUEFile(newContent, child))
		}
		newContent = fmt.Sprintf("%s\n}\n", newContent)
	case yaml.ScalarNode:
		matched, _ := regexp.MatchString(`^\$\{.*\}$`, n.Value)
		value := n.Value

		if matched {
			value = n.Value[2 : len(n.Value)-1]
		}

		if matched || n.Tag != "!!str" {
			return fmt.Sprintf("%s%s", newContent, value)
		} else {
			return fmt.Sprintf("%s\"%s\"", newContent, n.Value)

		}
	}

	return newContent
}

type Leaf struct {
	Path  string
	Value string
}

func GetLeaves(value cue.Value, skipReserved bool) []Leaf {
	result := make([]Leaf, 0)

	value.Walk(func(v cue.Value) bool {
		switch v.Kind() {
		case cue.BoolKind, cue.NumberKind, cue.StringKind, cue.BytesKind:
			sel := v.Path().Selectors()
			path := cue.MakePath(sel[2:]...).String()
			if skipReserved && strings.Contains(path, "$") || strings.Contains(path, "$metadata") {
				return true
			}
			result = append(
				result,
				Leaf{
					Path:  path,
					Value: fmt.Sprint(v),
				},
			)
		}
		return true
	}, nil)

	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i].Path, result[j].Path) == -1
	})

	return result
}

func SendTelemtry(telemetryEndpoint string, apiPath string, data interface{}) ([]byte, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	log.Debug("Sending: ", data)

	url, _ := url.Parse(telemetryEndpoint)
	url.Path = path.Join(url.Path, "api", apiPath)
	log.Debug("URL: ", url)

	request, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	log.Debug("Response Status: ", response.Status)
	log.Debug("Response Headers: ", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	log.Debug("Response Body: ", string(body))

	if response.StatusCode > 201 {
		errResponse := struct {
			Message string `json:"message"`
		}{
			Message: "API request failed",
		}

		err := json.Unmarshal(body, &errResponse)
		if err != nil {
			log.Fatalf("failed to parse error response body: %s", body)
		}

		return body, errors.New(errResponse.Message)
	}

	return body, nil
}
