package cuemods

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	// FS contains the filesystem of the stdlib.
	//go:embed guku.io
	FS embed.FS
)

// var (
// 	GukuModule  = "guku.io"
// 	GukuPackage = fmt.Sprintf(path.Join(GukuModule, "devx"))
// )

func InstallCore(cueModPath string) error {
	pkgDir := path.Join(cueModPath, "cue.mod", "pkg")
	return extractModules(pkgDir)
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

		contents := fmt.Sprintf(`module: "%s"`, module)
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

func GetCueModParent(args ...string) (string, bool) {
	cwd, _ := os.Getwd()
	parentDir := cwd

	if len(args) == 1 {
		parentDir = args[0]
	}

	found := false

	for {
		if _, err := os.Stat(path.Join(parentDir, "cue.mod")); !errors.Is(err, os.ErrNotExist) {
			found = true
			break // found it!
		}

		parentDir = filepath.Dir(parentDir)

		if parentDir == fmt.Sprintf("%s%s", filepath.VolumeName(parentDir), string(os.PathSeparator)) {
			// reached the root
			parentDir = cwd // reset to working directory
			break
		}
	}

	return parentDir, found
}

func extractModules(dest string) error {
	return fs.WalkDir(FS, ".", func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		// Do not vendor the package's `cue.mod/pkg`
		if strings.Contains(p, "cue.mod/pkg") {
			return nil
		}

		contents, err := fs.ReadFile(FS, p)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}

		overlayPath := path.Join(dest, p)

		if err := os.MkdirAll(filepath.Dir(overlayPath), 0755); err != nil {
			return err
		}

		// Give exec permission on embedded file to freely use shell script
		// Exclude permission linter
		//nolint
		return os.WriteFile(overlayPath, contents, 0700)
	})
}
