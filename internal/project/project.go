package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"devopzilla.com/guku/internal/utils"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

func Validate(configDir string, stackPath string) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}

	value := utils.LoadProject(configDir, &overlays)
	if err := ValidateProject(value, stackPath); err != nil {
		return err
	}

	fmt.Println("👌 All looks good")
	return nil
}

func ValidateProject(value cue.Value, stackPath string) error {
	err := value.Validate()
	if err != nil {
		return err
	}

	stack := value.LookupPath(cue.ParsePath(stackPath))
	if stack.Err() != nil {
		return stack.Err()
	}

	isValid := true
	err = errors.New("Invalid Components")
	utils.Walk(stack, func(v cue.Value) bool {
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

	return nil
}

func Discover(configDir string, showDefs bool, showTransformers bool) error {
	instances := utils.LoadInstances(configDir)

	deps := instances[0].Dependencies()

	for _, dep := range deps {
		if strings.Contains(dep.ID(), "traits") {
			ctx := cuecontext.New()
			value := ctx.BuildInstance(dep)

			fieldIter, _ := value.Fields(cue.Definitions(true), cue.Docs(true))
			fmt.Printf("[🏷️  traits] \"%s\"\n", dep.ID())
			for fieldIter.Next() {
				traits := fieldIter.Value().LookupPath(cue.ParsePath("$metadata.traits"))
				if traits.Exists() && traits.IsConcrete() {
					fmt.Printf("%s.%s", dep.PkgName, fieldIter.Selector().String())
					if utils.HasComments(fieldIter.Value()) {
						fmt.Printf("\t%s", utils.GetComments(fieldIter.Value()))
					}
					fmt.Println()
					if showDefs {
						fmt.Println(fieldIter.Value())
						fmt.Println()
					}
				}
			}
			fmt.Println()
		}
		if showTransformers && strings.Contains(dep.ID(), "transformers") {
			ctx := cuecontext.New()
			value := ctx.BuildInstance(dep)

			fieldIter, _ := value.Fields(cue.Definitions(true), cue.Docs(true))

			fmt.Printf("[🏭 transformers] \"%s\"\n", dep.ID())
			for fieldIter.Next() {
				required := ""

				traits := fieldIter.Value().LookupPath(cue.ParsePath("input.$metadata.traits"))
				if traits.Exists() {
					traitIter, _ := traits.Fields()
					for traitIter.Next() {
						required = fmt.Sprintf("%s trait:%s", required, traitIter.Label())
					}
				}

				fmt.Printf("%s.%s", dep.PkgName, fieldIter.Selector().String())
				if utils.HasComments(fieldIter.Value()) {
					fmt.Printf("\t%s", utils.GetComments(fieldIter.Value()))
				}
				if len(required) > 0 {
					fmt.Printf(" (requires%s)", required)
				}
				fmt.Println()
				if showDefs {
					fmt.Println(fieldIter.Value())
					fmt.Println()
				}
			}
			fmt.Println()
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
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
				]
			},
		]
	}
}	
	`), 0700)

	return nil
}

func Update(configDir string) error {
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

	packagesValue := cuemodule.LookupPath(cue.ParsePath("packages"))
	if packagesValue.Err() != nil {
		return packagesValue.Err()
	}

	var packages []string
	err = packagesValue.Decode(&packages)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		repoURL, repoRevision, repoPath, err := parsePackage(pkg)
		if err != nil {
			return err
		}

		repo, mfs, err := getRepo(repoURL)
		if err != nil {
			return err
		}

		hash, err := repo.ResolveRevision(plumbing.Revision(repoRevision))

		fmt.Printf("Downloading %s @ %s\n", pkg, hash)

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

func parsePackage(pkg string) (string, string, string, error) {
	parts := strings.SplitN(pkg, "@", 2)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("No revision specified")
	}
	url := "https://" + parts[0]
	parts = strings.SplitN(parts[1], "/", 2)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("No path specified")
	}
	revision := parts[0]
	path := parts[1]
	if !strings.HasPrefix(path, "pkg") {
		return "", "", "", fmt.Errorf("Path must start with '/pkg/'")
	}

	return url, revision, path, nil
}

func getRepo(repoURL string) (*git.Repository, *billy.Filesystem, error) {
	// try without auth
	mfs := memfs.New()
	storer := memory.NewStorage()
	repo, err := git.Clone(storer, mfs, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
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
			return nil, nil, fmt.Errorf("Failed to use git private key %s: %s", gitPrivateKeyFile, err)
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

		contents := fmt.Sprintf(`module: "%s"
packages: [
	"github.com/devopzilla/guku-devx-catalog@main/pkg",
]
		`, module)
		if err := os.WriteFile(modFile, []byte(contents), 0600); err != nil {
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
