package utils

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
)

func LoadProject(configDir string) cue.Value {
	buildConfig := &cueload.Config{
		Dir:     configDir,
		Overlay: map[string]cueload.Source{},
	}
	instances := cueload.Instances([]string{}, buildConfig)

	ctx := cuecontext.New()

	return ctx.BuildInstance(instances[0])
}

func GetLastPathFragement(value cue.Value) string {
	selector := value.Path().Selectors()
	return selector[len(selector)-1].String()
}
