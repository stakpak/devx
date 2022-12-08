package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
	"github.com/go-git/go-billy/v5"
	"gopkg.in/yaml.v3"
)

func LoadInstances(configDir string) []*build.Instance {
	buildConfig := &cueload.Config{
		Dir:     configDir,
		Overlay: map[string]cueload.Source{},
	}
	return cueload.Instances([]string{}, buildConfig)
}

func LoadProject(configDir string, overlays *map[string]string) cue.Value {
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
	instances := cueload.Instances([]string{}, buildConfig)

	ctx := cuecontext.New()

	return ctx.BuildInstance(instances[0])
}

func GetLastPathFragement(value cue.Value) string {
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

		data, err := ioutil.ReadAll(bufio.NewReader(content))
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

	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return overlays, err
	}

	for _, f := range files {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".devx.yaml") || strings.HasSuffix(f.Name(), ".devx.yml")) {
			file, err := ioutil.ReadFile(path.Join(configDir, f.Name()))
			if err != nil {
				return overlays, err
			}

			var n yaml.Node
			err = yaml.Unmarshal(file, &n)
			if err != nil {
				return overlays, err
			}

			overlays[f.Name()+".cue"] = BuildCUEFile("package main", nil, &n)
		}
	}

	return overlays, nil
}

func BuildCUEFile(content string, meta *DevxMeta, n *yaml.Node) string {
	newContent := content

	if meta == nil {
		meta = &DevxMeta{
			isExpr:  false,
			attrs:   []string{},
			merge:   []string{},
			imports: map[string]string{},
		}
	}
	if n.HeadComment != "" {
		for _, entry := range strings.Split(n.HeadComment, "\n") {
			setDirectives(meta, entry)
		}
	}
	// if n.LineComment != "" {
	// 	setDirectives(value.meta, n.LineComment)
	// }

	switch n.Kind {
	case yaml.DocumentNode:
		for alias, importPath := range meta.imports {
			newContent = fmt.Sprintf("%s\nimport %s %s", newContent, alias, importPath)
		}
		newContent += "\n"
		for _, child := range n.Content {
			newContent = BuildCUEFile(newContent, nil, child)
		}
	case yaml.SequenceNode:
		newContent = fmt.Sprintf("%s [\n", newContent)
		for _, child := range n.Content {
			newContent = fmt.Sprintf("%s,\n", BuildCUEFile(newContent, nil, child))
		}
		newContent = fmt.Sprintf("%s\n]\n", newContent)
	case yaml.MappingNode:
		newContent = fmt.Sprintf("%s {\n", newContent)
		for _, item := range meta.merge {
			newContent = fmt.Sprintf("%s\n%s\n", newContent, item)
		}
		for i := 0; i < len(n.Content); i += 2 {
			name := n.Content[i].Value
			childMeta := &DevxMeta{
				isExpr:  false,
				attrs:   []string{},
				merge:   []string{},
				imports: map[string]string{},
			}
			if n.Content[i].HeadComment != "" {
				for _, entry := range strings.Split(n.Content[i].HeadComment, "\n") {
					setDirectives(childMeta, entry)
				}
			}

			child := n.Content[i+1]
			newContent = fmt.Sprintf("%s%s: ", newContent, name)
			newContent = fmt.Sprintf("%s\n", BuildCUEFile(newContent, childMeta, child))
		}
		newContent = fmt.Sprintf("%s\n}\n", newContent)
	case yaml.ScalarNode:
		if meta.isExpr || n.Tag != "!!str" {
			return fmt.Sprintf("%s%s", newContent, n.Value)
		} else {
			return fmt.Sprintf("%s\"%s\"", newContent, n.Value)
		}
	}

	return newContent
}

type DevxMeta struct {
	attrs   []string
	merge   []string
	imports map[string]string
	isExpr  bool
}

func setDirectives(meta *DevxMeta, entry string) {
	trimmed := strings.TrimPrefix(entry, "#")
	trimmed = strings.TrimSpace(trimmed)

	if strings.HasPrefix(trimmed, "devx:expr") {
		meta.isExpr = true
	} else if strings.HasPrefix(trimmed, "devx:attr") {
		_, attr, _ := strings.Cut(trimmed, "devx:attr")
		meta.attrs = append(meta.attrs, strings.TrimSpace(attr))
	} else if strings.HasPrefix(trimmed, "devx:merge") {
		_, trait, _ := strings.Cut(trimmed, "devx:merge")
		meta.merge = append(meta.merge, strings.TrimSpace(trait))
	} else if strings.HasPrefix(trimmed, "devx:import") {
		_, imports, _ := strings.Cut(trimmed, "devx:import")
		importPair := strings.Split(strings.TrimSpace(imports), " ")
		meta.imports[importPair[0]] = importPair[1]
	}
}
