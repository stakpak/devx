package project

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/devopzilla/guku-devx/pkg/auth"
	"github.com/devopzilla/guku-devx/pkg/catalog"
	"github.com/devopzilla/guku-devx/pkg/gitrepo"
	"github.com/devopzilla/guku-devx/pkg/stack"
	"github.com/devopzilla/guku-devx/pkg/stackbuilder"
	"github.com/devopzilla/guku-devx/pkg/utils"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
)

const stakpakPrefix = "stakpak://"

func Validate(configDir string, stackPath string, buildersPath string, strict bool) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}

	value, _, _ := utils.LoadProject(configDir, &overlays)
	if err := ValidateProject(value, stackPath, buildersPath, strict); err != nil {
		return err
	}

	log.Info("👌 All looks good")
	return nil
}

func ValidateProject(value cue.Value, stackPath string, buildersPath string, strict bool) error {
	err := value.Validate()
	if err != nil {
		return err
	}

	stackValue := value.LookupPath(cue.ParsePath(stackPath))
	if stackValue.Err() != nil {
		return stackValue.Err()
	}

	isValid := true
	err = errors.New("invalid Components")
	utils.Walk(stackValue, func(v cue.Value) bool {
		gukuAttr := v.Attribute("guku")

		isRequired, _ := gukuAttr.Flag(0, "required")
		if isRequired && !v.IsConcrete() && !utils.IsReference(v) {
			isValid = false
			err = fmt.Errorf("%w\n%s is a required field", err, v.Path())
		}
		return true
	}, nil)

	if !isValid {
		return err
	}

	if strict {
		builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
		if err != nil {
			return err
		}

		stack, err := stack.NewStack(stackValue, "", []string{})
		if err != nil {
			return err
		}

		err = stackbuilder.CheckTraitFulfillment(builders, stack)
		if err != nil {
			return err
		}
	}

	return nil
}

func Discover(configDir string, showDefs bool, showTransformers bool) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}
	instances := utils.LoadInstances(configDir, &overlays)

	deps := instances[0].Dependencies()

	for _, dep := range deps {
		if strings.Contains(dep.ID(), "traits") {
			ctx := cuecontext.New()
			value := ctx.BuildInstance(dep)

			fieldIter, _ := value.Fields(cue.Definitions(true), cue.Docs(true))
			message := fmt.Sprintf("[🏷️  traits] \"%s\"\n", dep.ID())
			for fieldIter.Next() {
				traits := fieldIter.Value().LookupPath(cue.ParsePath("$metadata.traits"))
				if traits.Exists() && traits.IsConcrete() {
					message += fmt.Sprintf("%s.%s", dep.PkgName, fieldIter.Selector().String())
					if utils.HasComments(fieldIter.Value()) {
						message += fmt.Sprintf("\t%s", utils.GetComments(fieldIter.Value()))
					}
					message += "\n"
					if showDefs {
						message += fmt.Sprintf("%s\n\n", fieldIter.Value())
					}
				}
			}
			log.Info(message)
		}
		if showTransformers && strings.Contains(dep.ID(), "transformers") {
			ctx := cuecontext.New()
			value := ctx.BuildInstance(dep)

			fieldIter, _ := value.Fields(cue.Definitions(true), cue.Docs(true))

			message := fmt.Sprintf("[🏭 transformers] \"%s\"\n", dep.ID())
			for fieldIter.Next() {
				required := ""

				traits := fieldIter.Value().LookupPath(cue.ParsePath("input.$metadata.traits"))
				if traits.Exists() {
					traitIter, _ := traits.Fields()
					for traitIter.Next() {
						required = fmt.Sprintf("%s trait:%s", required, traitIter.Label())
					}
				}

				message += fmt.Sprintf("%s.%s", dep.PkgName, fieldIter.Selector().String())
				if utils.HasComments(fieldIter.Value()) {
					message += fmt.Sprintf("\t%s", utils.GetComments(fieldIter.Value()))
				}
				if len(required) > 0 {
					message += fmt.Sprintf(" (requires%s)", required)
				}
				message += "\n"
				if showDefs {
					message += fmt.Sprintf("%s\n\n", fieldIter.Value())
				}
			}
			log.Info(message)
		}
	}

	return nil
}

func Generate(configDir string) error {
	appPath := path.Join(configDir, "stack.cue")

	os.WriteFile(appPath, []byte(`package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		cowsay: {
			traits.#Workload
			containers: default: {
				image: "docker/whalesay"
				command: ["cowsay"]
				args: ["Hello DevX!"]
			}
		}
	}
}
`), 0700)

	builderPath := path.Join(configDir, "builder.cue")
	os.WriteFile(builderPath, []byte(`package main

import (
	"guku.io/devx/v2alpha1"
	"guku.io/devx/v2alpha1/environments"
)

builders: v2alpha1.#Environments & {
	dev: environments.#Compose
}
`), 0700)

	return nil
}

func Update(configDir string, server auth.ServerConfig) error {
	cuemodulePath := path.Join(configDir, "cue.mod", "module.cue")
	data, err := os.ReadFile(cuemodulePath)
	if err != nil {
		return err
	}

	ctx := cuecontext.New()
	cuemodule := ctx.CompileBytes(data)
	if cuemodule.Err() != nil {
		return cuemodule.Err()
	}

	deps := map[string]catalog.ModuleDependency{}

	oldPackages := cuemodule.LookupPath(cue.ParsePath("packages"))
	if oldPackages.Exists() {
		packages := []string{}
		err = oldPackages.Decode(&packages)
		if err != nil {
			return err
		}

		for _, pkg := range packages {
			parts := strings.SplitN(pkg, "@", 2)
			name := parts[0]

			parts = strings.SplitN(parts[1], ":", 2)
			version := parts[0]
			dep := catalog.ModuleDependency{
				V: nil,
			}

			if semver.IsValid(version) {
				dep.V = &version
			}
			deps[name] = dep
		}

		moduleName, err := cuemodule.LookupPath(cue.ParsePath("module")).String()
		if err != nil {
			return err
		}
		if err := updateModuleFile(configDir, ctx, moduleName, deps); err != nil {
			return err
		}

		log.Info("Updated module.cue format")
	}

	depsValue := cuemodule.LookupPath(cue.ParsePath("deps"))
	if depsValue.Exists() {
		if depsValue.Err() != nil {
			return depsValue.Err()
		}

		err = depsValue.Decode(&deps)
		if err != nil {
			return err
		}
	}

	for name, pkg := range deps {
		if strings.HasPrefix(name, stakpakPrefix) {
			name = strings.TrimPrefix(name, stakpakPrefix)
			queryParams := map[string]string{
				"name": name,
			}

			if pkg.V != nil {
				queryParams["version"] = *pkg.V
			}

			data, err := utils.GetData(
				server,
				path.Join("package", "fetch"),
				nil,
				queryParams,
			)
			if err != nil {
				return err
			}
			packageItem := catalog.ModuleItem{}
			err = json.Unmarshal(data, &packageItem)
			if err != nil {
				return err
			}

			installedVersion := "<untagged>"
			if len(packageItem.Git.Tags) > 0 {
				installedVersion = packageItem.Git.Tags[0]
			}

			log.Infof("📦 Installing %s@%s", name, installedVersion)

			pkgDir := path.Join(configDir, "cue.mod", "pkg", name)
			err = os.RemoveAll(pkgDir)
			if err != nil {
				return err
			}

			for path, content := range packageItem.Source {
				writePath := filepath.Join(pkgDir, path)
				writeDirPath := filepath.Dir(writePath)
				if err := os.MkdirAll(writeDirPath, 0755); err != nil {
					return err
				}
				if err != os.WriteFile(writePath, []byte(content), 0700) {
					return err
				}
			}
			continue
		}

		repoURL := "https://" + name
		repoRevision := "main"
		if pkg.V != nil {
			repoRevision = *pkg.V
		}
		repoPath := ""

		repo, mfs, err := getRepo(repoURL)
		if err != nil {
			return err
		}

		hash, err := repo.ResolveRevision(plumbing.Revision(repoRevision))
		if err != nil {
			return err
		}

		log.Infof("📦 Installing %s@%s", name, hash)

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: *hash,
		})
		if err != nil {
			return err
		}

		moduleFilePath := filepath.Join("cue.mod", "module.cue")
		_, err = (*mfs).Lstat(moduleFilePath)
		if err == nil {
			content, err := (*mfs).Open(moduleFilePath)
			if err != nil {
				return err
			}
			moduleData, err := io.ReadAll(bufio.NewReader(content))
			if err != nil {
				return err
			}
			module := ctx.CompileBytes(moduleData)
			moduleName := module.LookupPath(cue.ParsePath("module"))
			if moduleName.Err() != nil {
				return moduleName.Err()
			}

			modulePrefix, err := moduleName.String()
			if err != nil {
				return err
			}

			log.Debug("Module prefix: ", modulePrefix)
			pkgDir := path.Join(configDir, "cue.mod", "pkg", modulePrefix)
			pkgSubDir := path.Join(pkgDir, repoPath)
			log.Debug("Updating package ", pkgSubDir)
			err = os.RemoveAll(pkgSubDir)
			if err != nil {
				return err
			}

			err = utils.FsWalk(*mfs, repoPath, func(file string, content []byte) error {
				if strings.HasPrefix(file, ".") ||
					strings.HasPrefix(file, "cue.mod") ||
					// strings.HasPrefix(file, "pkg") ||
					!strings.HasSuffix(file, ".cue") {
					return nil
				}

				writePath := path.Join(pkgDir, file)
				if err := os.MkdirAll(filepath.Dir(writePath), 0755); err != nil {
					return err
				}
				return os.WriteFile(writePath, content, 0700)
			})
			if err != nil {
				return err
			}

			log.Debugf("Updating packages %s dependencies", pkgDir)
			if strings.HasPrefix(modulePrefix, "guku.io/devx") {
				moduleDepPkgPath := path.Join("cue.mod", "pkg")
				packageInfo, err := (*mfs).ReadDir(moduleDepPkgPath)
				if err != nil {
					return err
				}

				for _, info := range packageInfo {
					pkgDir := path.Join(configDir, moduleDepPkgPath, info.Name())
					log.Debug("Updating dependency ", pkgDir)
					err = os.RemoveAll(pkgDir)
					if err != nil {
						return err
					}

					err = utils.FsWalk(*mfs, pkgDir, func(file string, content []byte) error {
						if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
							return err
						}
						return os.WriteFile(file, content, 0700)
					})
					if err != nil {
						return err
					}
				}
			}

			continue
		}

		// fallback to legacy package management
		packageInfo, err := (*mfs).ReadDir(repoPath)
		if err != nil {
			return err
		}

		for _, info := range packageInfo {
			pkgDir := path.Join(configDir, "cue.mod", repoPath, info.Name())
			err = os.RemoveAll(pkgDir)
			if err != nil {
				return err
			}
		}

		err = utils.FsWalk(*mfs, repoPath, func(file string, content []byte) error {
			writePath := path.Join(configDir, "cue.mod", file)

			if err := os.MkdirAll(filepath.Dir(writePath), 0755); err != nil {
				return err
			}

			return os.WriteFile(writePath, content, 0700)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func getRepo(repoURL string) (*git.Repository, *billy.Filesystem, error) {
	// try without auth
	mfs := memfs.New()
	storer := memory.NewStorage()
	repo, err := git.Clone(storer, mfs, &git.CloneOptions{
		NoCheckout: true,
		URL:        repoURL,
		Depth:      1,
	})
	if err == nil {
		return repo, &mfs, nil
	}
	if err.Error() != "authentication required" {
		return nil, nil, err
	}

	// fetch credentials
	gitUsername := os.Getenv("GIT_USERNAME")
	gitPassword := os.Getenv("GIT_PASSWORD")
	gitPrivateKeyFile := os.Getenv("GIT_PRIVATE_KEY_FILE")
	gitPrivateKeyFilePassword := os.Getenv("GIT_PRIVATE_KEY_FILE_PASSWORD")

	if gitPrivateKeyFile == "" && gitPassword == "" {
		return nil, nil, fmt.Errorf(`To access private repos please provide
GIT_USERNAME & GIT_PASSWORD
or
GIT_PRIVATE_KEY_FILE & GIT_PRIVATE_KEY_FILE_PASSWORD`)
	}

	if gitPassword != "" {
		auth := http.BasicAuth{
			Username: gitUsername,
			Password: gitPassword,
		}

		mfs = memfs.New()
		storer = memory.NewStorage()
		repo, err = git.Clone(storer, mfs, &git.CloneOptions{
			URL:   repoURL,
			Auth:  &auth,
			Depth: 1,
		})
		if err != nil {
			return nil, nil, err
		}
		return repo, &mfs, nil
	}

	if gitPrivateKeyFile != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", gitPrivateKeyFile, gitPrivateKeyFilePassword)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to use git private key %s: %s", gitPrivateKeyFile, err.Error())
		}

		mfs = memfs.New()
		storer = memory.NewStorage()
		repo, err = git.Clone(storer, mfs, &git.CloneOptions{
			URL:   repoURL,
			Auth:  publicKeys,
			Depth: 1,
		})
		if err != nil {
			return nil, nil, err
		}
		return repo, &mfs, nil
	}

	return nil, nil, fmt.Errorf("Could not fetch repo")
}

func Init(ctx context.Context, parentDir, module string) error {
	absParentDir, err := filepath.Abs(parentDir)
	if err != nil {
		return err
	}

	modDir := path.Join(absParentDir, "cue.mod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	modFile := path.Join(modDir, "module.cue")
	if _, err := os.Stat(modFile); err != nil {
		statErr, ok := err.(*os.PathError)
		if !ok {
			return statErr
		}

		ctx := cuecontext.New()
		if err := updateModuleFile(parentDir, ctx, module, map[string]catalog.ModuleDependency{
			"github.com/devopzilla/guku-devx-catalog": {
				V: nil,
			},
		}); err != nil {
			return err
		}
	}

	if err := os.Mkdir(path.Join(modDir, "pkg"), 0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	return nil
}

type ProjectData struct {
	Stack        string                  `json:"stack"`
	Environments []string                `json:"environments"`
	Imports      []string                `json:"imports"`
	Git          *gitrepo.ProjectGitData `json:"git"`
}

func Publish(configDir string, stackPath string, buildersPath string, server auth.ServerConfig) error {
	if !server.Enable {
		return fmt.Errorf("-T telemtry should be enabled to publish stack")
	}

	project := ProjectData{}

	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}

	value, stackId, depIds := utils.LoadProject(configDir, &overlays)
	if err := ValidateProject(value, stackPath, buildersPath, false); err != nil {
		return err
	}

	if stackId == "" {
		return fmt.Errorf("cannot publish this stack without a module id. please set the \"module\" key to a unique value in \"cue.mod/module.cue\"")
	}

	project.Stack = stackId
	project.Imports = depIds

	_, err = stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)), stackId, depIds)
	if err != nil {
		return err
	}

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return err
	}
	environments := make([]string, 0, len(builders))
	for k := range builders {
		environments = append(environments, k)
	}
	project.Environments = environments

	gitData, err := gitrepo.GetProjectGitData(configDir)
	if err != nil {
		return err
	}
	project.Git = gitData

	_, err = utils.SendData(server, "stacks", &project)
	if err != nil {
		return err
	}

	log.Infof("🚀 Published %s successfully", stackId)

	return nil
}

func Import(newPackage string, configDir string, server auth.ServerConfig) error {
	pkgParts := strings.Split(newPackage, "@")
	if len(pkgParts) < 2 {
		return fmt.Errorf("invalid package format, expected \"<git repo>@<git revision>\"")
	}
	if len(pkgParts[0]) == 0 {
		return fmt.Errorf("invalid package format, git repo should not be empty")
	}
	if len(pkgParts[1]) == 0 {
		return fmt.Errorf("invalid package format, git revision should not be empty")
	}
	gitRepo := pkgParts[0]
	gitRevision := pkgParts[1]

	cuemodulePath := path.Join(configDir, "cue.mod", "module.cue")
	data, err := os.ReadFile(cuemodulePath)
	if err != nil {
		return err
	}

	ctx := cuecontext.New()
	cuemodule := ctx.CompileBytes(data)
	if cuemodule.Err() != nil {
		return cuemodule.Err()
	}

	moduleName, err := cuemodule.LookupPath(cue.ParsePath("module")).String()
	if err != nil {
		return err
	}

	deps := map[string]catalog.ModuleDependency{}
	depsValue := cuemodule.LookupPath(cue.ParsePath("deps"))
	if depsValue.Exists() {
		err = depsValue.Decode(&deps)
		if err != nil {
			return err
		}
	}

	for name := range deps {
		if strings.HasPrefix(name, gitRepo) {
			log.Infof("Module %s already exists", gitRepo)
			return nil
		}
	}

	deps[gitRepo] = catalog.ModuleDependency{
		V: &gitRevision,
	}

	if err := updateModuleFile(configDir, ctx, moduleName, deps); err != nil {
		return err
	}

	err = Update(configDir, server)
	if err != nil {
		log.Error(err.Error())
		return errors.New("failed to update packages, fix this issue and re-run devx project update")
	}

	return nil
}

func updateModuleFile(configDir string, ctx *cue.Context, module string, deps map[string]catalog.ModuleDependency) error {
	cuemodulePath := path.Join(configDir, "cue.mod", "module.cue")
	newcuemodule := ctx.CompileString("")
	newcuemodule = newcuemodule.FillPath(cue.ParsePath("module"), module)
	newcuemodule = newcuemodule.FillPath(cue.ParsePath("cue.lang"), "v0.6.0-alpha.1")
	newcuemodule = newcuemodule.FillPath(cue.ParsePath("deps"), deps)
	bytes, err := format.Node(newcuemodule.Syntax(cue.Concrete(true), cue.Final()), format.Simplify())
	if err != nil {
		return err
	}
	err = os.WriteFile(cuemodulePath, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}
