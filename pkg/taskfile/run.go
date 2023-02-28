package taskfile

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"

	"devopzilla.com/guku-devx/pkg/stackbuilder"
	"devopzilla.com/guku-devx/pkg/utils"

	"mvdan.cc/sh/v3/syntax"

	"github.com/go-task/task/v3"
	taskargs "github.com/go-task/task/v3/args"

	"github.com/go-task/task/v3/taskfile"

	log "github.com/sirupsen/logrus"
)

type RunFlags struct {
	List     bool
	ListAll  bool
	ListJson bool
	Status   bool
	Force    bool
	Watch    bool
	Dry      bool
	Summary  bool
	ExitCode bool
	Parallel bool
	Verbose  bool
	Color    bool
	Interval time.Duration
}

func Run(configDir string, buildersPath string, runFlags RunFlags, environment string, doubleDashPos int, args []string) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}
	value, _, _ := utils.LoadProject(configDir, &overlays)
	err = value.Validate()
	if err != nil {
		return err
	}

	buildersValue := value.LookupPath(cue.ParsePath(buildersPath))
	if !buildersValue.Exists() {
		return fmt.Errorf("missing builders field")
	}

	builders, err := stackbuilder.NewEnvironments(buildersValue)
	if err != nil {
		return err
	}

	builder, ok := builders[environment]
	if !ok {
		return fmt.Errorf("environment %s was not found", environment)
	}

	if builder.Taskfile == nil {
		return fmt.Errorf("no taskfile definition found in environment %s", environment)
	}

	log.Debug(builder.Taskfile)
	err = builder.Taskfile.Validate(cue.Concrete(true))
	if err != nil {
		log.Error(err)
		return err
	}

	taskFileContent, err := cueyaml.Encode(*builder.Taskfile)
	if err != nil {
		log.Error(err)
		return err
	}

	taskFile, err := os.CreateTemp(configDir, ".taskfile-*.yml")
	if err != nil {
		log.Error(err)
		return err
	}
	defer os.RemoveAll(taskFile.Name())

	_, err = taskFile.Write(taskFileContent)
	if err != nil {
		log.Error(err)
		return err
	}

	e := task.Executor{
		Force:       runFlags.Force,
		Watch:       runFlags.Watch,
		Verbose:     runFlags.Verbose,
		Silent:      true,
		Dir:         "",
		Dry:         runFlags.Dry,
		Entrypoint:  taskFile.Name(),
		Summary:     runFlags.Summary,
		Parallel:    runFlags.Parallel,
		Color:       runFlags.Color,
		Concurrency: 0,
		Interval:    runFlags.Interval,

		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,

		OutputStyle: taskfile.Output{
			Name: "",
			Group: taskfile.OutputGroup{
				Begin: "",
				End:   "",
			},
		},
	}

	var listOptions = task.NewListOptions(runFlags.List, runFlags.ListAll, runFlags.ListJson)
	if err := listOptions.Validate(); err != nil {
		log.Error(err)
		return err
	}

	if listOptions.ShouldListTasks() {
		e.ListTaskNames(runFlags.ListAll)
		return nil
	}

	err = e.Setup()
	if err != nil {
		log.Error(err)
		return err
	}

	if listOptions.ShouldListTasks() {
		if foundTasks, err := e.ListTasks(listOptions); !foundTasks || err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	var tasks []string
	cliArgs := ""
	if doubleDashPos == -1 {
		tasks = args
	} else {
		var quotedCliArgs []string
		for _, arg := range args[doubleDashPos-1:] {
			quotedCliArg, err := syntax.Quote(arg, syntax.LangBash)
			if err != nil {
				log.Error(err)
				return nil
			}
			quotedCliArgs = append(quotedCliArgs, quotedCliArg)
		}
		tasks = args[:doubleDashPos-1]
		cliArgs = strings.Join(quotedCliArgs, " ")
	}

	calls, globals := taskargs.ParseV3(tasks...)

	globals.Set("CLI_ARGS", taskfile.Var{Static: cliArgs})
	e.Taskfile.Vars.Merge(globals)
	e.InterceptInterruptSignals()

	if err := e.Run(context.TODO(), calls...); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
