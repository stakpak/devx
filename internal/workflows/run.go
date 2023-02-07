package workflows

import (
	"context"
	"fmt"
	"os"
	"time"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"

	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/args"
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
	// parallel    bool
	// concurrency int
	// output      taskfile.Output
	Color    bool
	Interval time.Duration
}

func Run(configDir string, buildersPath string, runFlags RunFlags, environment string, tasks []string) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}
	value, _, _ := utils.LoadProject(configDir, &overlays)

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return err
	}

	builder, ok := builders[environment]
	if !ok {
		return fmt.Errorf("environment %s was not found", environment)
	}

	log.Debug(builder.Taskfile)

	taskFileContent, err := cueyaml.Encode(*builder.Taskfile)
	if err != nil {
		return err
	}

	taskFile, err := os.CreateTemp(configDir, ".taskfile-*.yml")
	if err != nil {
		return err
	}
	defer os.RemoveAll(taskFile.Name())

	_, err = taskFile.Write(taskFileContent)
	if err != nil {
		return err
	}

	e := task.Executor{
		Force:       runFlags.Force,
		Watch:       runFlags.Watch,
		Verbose:     false,
		Silent:      true,
		Dir:         "",
		Dry:         runFlags.Dry,
		Entrypoint:  taskFile.Name(),
		Summary:     runFlags.Summary,
		Parallel:    false,
		Color:       runFlags.Color,
		Concurrency: 0,
		Interval:    0,

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
		log.Fatal(err)
	}

	if listOptions.ShouldListTasks() {
		e.ListTaskNames(runFlags.ListAll)
		return nil
	}

	err = e.Setup()
	if err != nil {
		return err
	}

	if listOptions.ShouldListTasks() {
		if foundTasks, err := e.ListTasks(listOptions); !foundTasks || err != nil {
			return err
		}
		return nil
	}

	calls, globals := args.ParseV3(tasks...)
	e.Taskfile.Vars.Merge(globals)
	e.InterceptInterruptSignals()

	if err := e.Run(context.TODO(), calls...); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
