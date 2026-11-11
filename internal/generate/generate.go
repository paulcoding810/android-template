// Package generate creates a new project from an embedded template.
package generate

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/paulcoding810/android-template/internal/icon"
	"github.com/paulcoding810/android-template/templates"
)

const (
	oldPkg       = "com.paulcoding.androidtemplate"
	oldProjName  = "AndroidTemplate"
	oldProjLower = "androidtemplate"
)

// Local build output and IDE state that must never be copied into a new project.
var skippedNames = map[string]bool{
	"build":            true,
	".gradle":          true,
	".kotlin":          true,
	".idea":            true,
	"local.properties": true,
	"xcuserdata":       true,
	"DerivedData":      true,
	".DS_Store":        true,
}

var (
	pkgSegment  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	projNameRe  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	teamIDRe    = regexp.MustCompile(`^[A-Z0-9]{10}$`)
	kotlinWords = map[string]bool{
		"as": true, "break": true, "class": true, "continue": true, "do": true, "else": true,
		"false": true, "for": true, "fun": true, "if": true, "in": true, "interface": true,
		"is": true, "null": true, "object": true, "package": true, "return": true, "super": true,
		"this": true, "throw": true, "true": true, "try": true, "typealias": true, "typeof": true,
		"val": true, "var": true, "when": true, "while": true,
	}
)

type Options struct {
	Package  string
	Project  string
	Template string
	OutDir   string
	TeamID   string
	NoGit    bool
	Icon     *icon.Options
}

// Project validates opts and writes the new project to OutDir/Project.
func Project(opts Options, stdout, stderr io.Writer) error {
	if err := validate(opts); err != nil {
		return err
	}

	dir := filepath.Join(opts.OutDir, opts.Project)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory %q already exists", dir)
	}

	if err := extract(dir, opts); err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("extracting template: %w", err)
	}

	if opts.Icon != nil {
		if err := icon.Apply(dir, *opts.Icon); err != nil {
			os.RemoveAll(dir)
			return fmt.Errorf("generating icon: %w", err)
		}
	}

	if !opts.NoGit {
		if err := gitInit(dir, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "Warning: git init failed: %v\n", err)
		}
	}

	fmt.Fprintf(stdout, "\nDone! Project created at: %s\n", dir)
	return nil
}

func validate(opts Options) error {
	if !templateExists(opts.Template) {
		return fmt.Errorf("unknown template %q", opts.Template)
	}
	if err := validatePackage(opts.Package); err != nil {
		return err
	}
	if !projNameRe.MatchString(opts.Project) {
		return fmt.Errorf("invalid project name %q: use letters and digits only, starting with a letter", opts.Project)
	}
	if opts.TeamID != "" {
		if opts.Template != "kmm" {
			return errors.New("--team-id only applies to the kmm template")
		}
		if !teamIDRe.MatchString(opts.TeamID) {
			return fmt.Errorf("invalid team ID %q: expected 10 uppercase letters or digits", opts.TeamID)
		}
	}
	return nil
}

func templateExists(name string) bool {
	for _, t := range templates.All {
		if t.Name == name {
			return true
		}
	}
	return false
}

func validatePackage(pkg string) error {
	segments := strings.Split(pkg, ".")
	if len(segments) < 2 {
		return fmt.Errorf("invalid package %q: needs at least two segments, e.g. com.example.app", pkg)
	}
	for _, s := range segments {
		if !pkgSegment.MatchString(s) {
			return fmt.Errorf("invalid package %q: segment %q must be lowercase letters, digits or _, starting with a letter", pkg, s)
		}
		if kotlinWords[s] {
			return fmt.Errorf("invalid package %q: segment %q is a Kotlin keyword", pkg, s)
		}
	}
	return nil
}

func extract(outDir string, opts Options) error {
	newPkgSlash := strings.ReplaceAll(opts.Package, ".", "/")
	oldPkgSlash := strings.ReplaceAll(oldPkg, ".", "/")
	replacer := strings.NewReplacer(
		oldPkg, opts.Package,
		oldProjName, opts.Project,
		oldProjLower, strings.ToLower(opts.Project),
	)

	return fs.WalkDir(templates.FS, opts.Template, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skippedNames[d.Name()] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		rel := strings.TrimPrefix(strings.TrimPrefix(p, opts.Template), "/")
		rel = replacer.Replace(strings.ReplaceAll(rel, oldPkgSlash, newPkgSlash))
		outPath := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		data, err := templates.FS.ReadFile(p)
		if err != nil {
			return err
		}

		content := replacer.Replace(string(data))
		if opts.TeamID != "" && path.Base(p) == "Config.xcconfig" {
			content = strings.Replace(content, "TEAM_ID=\n", "TEAM_ID="+opts.TeamID+"\n", 1)
		}

		perm := fs.FileMode(0644)
		if d.Name() == "gradlew" {
			perm = 0755
		}
		return os.WriteFile(outPath, []byte(content), perm)
	})
}

func gitInit(dir string, stdout, stderr io.Writer) error {
	for _, args := range [][]string{
		{"init"},
		{"add", "."},
		{"commit", "-m", "Initial commit"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Stdout = stdout
		c.Stderr = stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("git %s: %w", args[0], err)
		}
	}
	return nil
}
