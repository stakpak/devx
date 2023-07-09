package taskfile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"cuelang.org/go/cue"
	cueyaml "cuelang.org/go/encoding/yaml"

	"github.com/stakpak/devx/pkg/auth"
	"github.com/stakpak/devx/pkg/gitrepo"
	"github.com/stakpak/devx/pkg/stackbuilder"
	"github.com/stakpak/devx/pkg/utils"

	"mvdan.cc/sh/v3/syntax"

	"github.com/acarl005/stripansi"
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

func Run(configDir string, buildersPath string, server auth.ServerConfig, runFlags RunFlags, environment string, doubleDashPos int, args []string) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}
	value, stackId, _ := utils.LoadProject(configDir, &overlays)
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

	outBuff := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	outWriter := io.MultiWriter(os.Stdout, outBuff)
	errWriter := io.MultiWriter(os.Stderr, errBuff)

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
		Stdout: outWriter,
		Stderr: errWriter,

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

	err = e.Setup()
	if err != nil {
		log.Error(err)
		return err
	}

	if len(args) == 0 && !listOptions.ShouldListTasks() {
		listOptions.ListAllTasks = true
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

	taskMap := map[string]map[string]string{}
	for _, call := range calls {
		args := map[string]string{}
		if call.Vars != nil {
			for k, v := range call.Vars.Mapping {
				args[k] = v.Static
				if v.Sh != "" {
					args[k] = v.Sh
				}
			}
		}
		taskMap[call.Task] = args
	}
	globalsMap := map[string]string{}
	if globals != nil {
		for k, v := range globals.Mapping {
			globalsMap[k] = v.Static
			if v.Sh != "" {
				globalsMap[k] = v.Sh
			}
		}
	}
	source := string(taskFileContent)

	if err := e.Run(context.TODO(), calls...); err != nil {
		taskErr := fmt.Sprintf("%s\n%s", err.Error(), stripansi.Strip(errBuff.String()))
		taskOutput := stripansi.Strip(outBuff.String())
		if taskId, err := sendTask(configDir, server, environment, stackId, string(source), taskMap, globalsMap, &taskErr, &taskOutput); err != nil {
			log.Error("failed to save task run data: ", err.Error())
		} else {
			log.Infof("\nSaved failed task run at %s/tasks/%s\n", server.Endpoint, taskId)
		}
		return err
	}

	if auth.IsLoggedIn(server) {
		taskOutput := stripansi.Strip(outBuff.String())
		if taskId, err := sendTask(configDir, server, environment, stackId, string(source), taskMap, globalsMap, nil, &taskOutput); err != nil {
			log.Error("failed to save task run data: ", err.Error())
		} else {
			log.Infof("\nCreated task run at %s/tasks/%s\n", server.Endpoint, taskId)
		}
	}
	return nil
}

type TaskData struct {
	Stack       string                       `json:"stack"`
	Identity    string                       `json:"identity,omitempty"`
	Environment string                       `json:"environment"`
	Git         *gitrepo.GitData             `json:"git,omitempty"`
	Error       *string                      `json:"error"`
	Output      *string                      `json:"output"`
	Source      string                       `json:"source"`
	Tasks       map[string]map[string]string `json:"tasks"`
	Globals     map[string]string            `json:"globals"`
}

func sendTask(configDir string, server auth.ServerConfig, environment string, stack string, source string, tasks map[string]map[string]string, globals map[string]string, taskError *string, taskOutput *string) (string, error) {
	taskData := TaskData{
		Stack:       stack,
		Identity:    "",
		Environment: environment,
		Git:         nil,
		Source:      source,
		Tasks:       tasks,
		Globals:     globals,
		Error:       taskError,
		Output:      taskOutput,
	}

	gitData, err := gitrepo.GetGitData(configDir)
	if err != nil {
		return "", err
	}
	taskData.Git = gitData

	data, err := utils.SendData(server, "tasks", &taskData)
	if err != nil {
		return "", err
	}

	taskResponse := make(map[string]string)
	err = json.Unmarshal(data, &taskResponse)
	if err != nil {
		return "", err
	}

	return taskResponse["id"], nil
}
