package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed all:template-main all:template-single all:template-kmm
var templates embed.FS

const (
	oldPkg       = "com.paulcoding.androidtemplate"
	oldProjName  = "AndroidTemplate"
	oldProjLower = "androidtemplate"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintf(os.Stderr, "Usage: %s <package.name> <ProjectName> [--single|--kmm]\n", os.Args[0])
		os.Exit(1)
	}

	newPkg := os.Args[1]
	newProj := os.Args[2]
	single := len(os.Args) == 4 && os.Args[3] == "--single"
	kmm := len(os.Args) == 4 && os.Args[3] == "--kmm"

	templateDir := "template-main"
	if single {
		templateDir = "template-single"
	} else if kmm {
		templateDir = "template-kmm"
	}

	dir := filepath.Join(".", newProj)
	if _, err := os.Stat(dir); err == nil {
		fmt.Fprintf(os.Stderr, "Error: directory %q already exists\n", dir)
		os.Exit(1)
	}

	extractTemplate(templateDir, dir, newPkg, newProj)

	if err := gitInit(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: git init failed: %v\n", err)
	}

	fmt.Printf("\nDone! Project created at: %s\n", dir)
}

func extractTemplate(templateDir, outDir, newPkg, newProj string) {
	newPkgSlash := strings.ReplaceAll(newPkg, ".", "/")
	oldPkgSlash := strings.ReplaceAll(oldPkg, ".", "/")
	newProjLower := strings.ToLower(newProj)

	fs.WalkDir(templates, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(templateDir, path)
		outRel := strings.ReplaceAll(rel, oldPkgSlash, newPkgSlash)
		outPath := filepath.Join(outDir, outRel)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		data, err := templates.ReadFile(path)
		if err != nil {
			return err
		}

		content := strings.ReplaceAll(string(data), oldPkg, newPkg)
		content = strings.ReplaceAll(content, oldProjName, newProj)
		content = strings.ReplaceAll(content, oldProjLower, newProjLower)

		perm := fs.FileMode(0644)
		if filepath.Base(outPath) == "gradlew" {
			perm = 0755
		}
		return os.WriteFile(outPath, []byte(content), perm)
	})
}

func gitInit(dir string) error {
	for _, cmd := range []struct {
		args []string
		msg  string
	}{
		{[]string{"init"}, "git init"},
		{[]string{"add", "."}, "git add"},
		{[]string{"commit", "-m", "Initial commit"}, "git commit"},
	} {
		c := exec.Command("git", cmd.args...)
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("%s: %w", cmd.msg, err)
		}
	}
	return nil
}
