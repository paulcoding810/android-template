package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed all:template-main all:template-single all:template-kmm
var templates embed.FS

const (
	oldPkg       = "com.paulcoding.androidtemplate"
	oldProjName  = "AndroidTemplate"
	oldProjLower = "androidtemplate"
)

var templateDirs = map[string]string{
	"main":   "template-main",
	"single": "template-single",
	"kmm":    "template-kmm",
}

var templateDescriptions = map[string]string{
	"main":   "multi-module Android app (app, core/*, feature/*)",
	"single": "single-module Android app",
	"kmm":    "Kotlin Multiplatform app (Android, iOS, JVM)",
}

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

// errReported marks errors the flag package has already printed.
var errReported = errors.New("already reported")

type options struct {
	pkg      string
	project  string
	template string
	outDir   string
	teamID   string
	noGit    bool
	icon     *iconOptions
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errReported) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "icon" {
		return runIcon(args[1:], stdout, stderr)
	}

	fset := flag.NewFlagSet("gen-android", flag.ContinueOnError)
	fset.SetOutput(stderr)
	single := fset.Bool("single", false, "use the single-module template")
	kmm := fset.Bool("kmm", false, "use the Kotlin Multiplatform template")
	list := fset.Bool("list", false, "list available templates and exit")
	outDir := fset.String("output", ".", "parent directory for the new project")
	teamID := fset.String("team-id", "", "Apple team ID for the iOS app (kmm only)")
	noGit := fset.Bool("no-git", false, "skip git init and the initial commit")
	icon := fset.String("icon", "", "SVG file for the launcher icon")
	iconBg := fset.String("icon-bg", "#FFFFFF", "launcher icon background color")
	fset.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gen-android <package.name> <ProjectName> [flags]\n")
		fmt.Fprintf(stderr, "       gen-android icon <icon.svg> [--dir <project>] [--bg #RRGGBB]\n\nFlags:\n")
		fset.PrintDefaults()
	}

	positional, err := parseInterspersed(fset, args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return errReported
	}

	if *list {
		printTemplates(stdout)
		return nil
	}

	if len(positional) != 2 {
		fset.Usage()
		return fmt.Errorf("expected 2 arguments, got %d", len(positional))
	}
	if *single && *kmm {
		return errors.New("--single and --kmm cannot be used together")
	}

	opts := options{
		pkg:      positional[0],
		project:  positional[1],
		template: "main",
		outDir:   *outDir,
		teamID:   *teamID,
		noGit:    *noGit,
	}
	if *icon != "" {
		iconOpts, err := loadIconOptions(*icon, *iconBg, defaultScale)
		if err != nil {
			return err
		}
		opts.icon = &iconOpts
	}
	if *single {
		opts.template = "single"
	} else if *kmm {
		opts.template = "kmm"
	}

	return generate(opts, stdout, stderr)
}

// parseInterspersed lets flags appear before, between or after positional arguments.
func parseInterspersed(fset *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fset.Parse(args); err != nil {
			return nil, err
		}
		if fset.NArg() == 0 {
			return positional, nil
		}
		positional = append(positional, fset.Arg(0))
		args = fset.Args()[1:]
	}
}

func printTemplates(w io.Writer) {
	names := make([]string, 0, len(templateDirs))
	for name := range templateDirs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		flagName := "(default)"
		if name != "main" {
			flagName = "--" + name
		}
		fmt.Fprintf(w, "%-7s %-10s %s\n", name, flagName, templateDescriptions[name])
	}
}

func generate(opts options, stdout, stderr io.Writer) error {
	if err := validatePackage(opts.pkg); err != nil {
		return err
	}
	if !projNameRe.MatchString(opts.project) {
		return fmt.Errorf("invalid project name %q: use letters and digits only, starting with a letter", opts.project)
	}
	if opts.teamID != "" {
		if opts.template != "kmm" {
			return errors.New("--team-id only applies to the kmm template")
		}
		if !teamIDRe.MatchString(opts.teamID) {
			return fmt.Errorf("invalid team ID %q: expected 10 uppercase letters or digits", opts.teamID)
		}
	}

	dir := filepath.Join(opts.outDir, opts.project)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory %q already exists", dir)
	}

	if err := extractTemplate(templateDirs[opts.template], dir, opts); err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("extracting template: %w", err)
	}

	if opts.icon != nil {
		if err := applyIcon(dir, *opts.icon); err != nil {
			os.RemoveAll(dir)
			return fmt.Errorf("generating icon: %w", err)
		}
	}

	if !opts.noGit {
		if err := gitInit(dir, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "Warning: git init failed: %v\n", err)
		}
	}

	fmt.Fprintf(stdout, "\nDone! Project created at: %s\n", dir)
	return nil
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

func extractTemplate(templateDir, outDir string, opts options) error {
	newPkgSlash := strings.ReplaceAll(opts.pkg, ".", "/")
	oldPkgSlash := strings.ReplaceAll(oldPkg, ".", "/")
	replacer := strings.NewReplacer(
		oldPkg, opts.pkg,
		oldProjName, opts.project,
		oldProjLower, strings.ToLower(opts.project),
	)

	return fs.WalkDir(templates, templateDir, func(p string, d fs.DirEntry, err error) error {
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

		rel := strings.TrimPrefix(strings.TrimPrefix(p, templateDir), "/")
		rel = replacer.Replace(strings.ReplaceAll(rel, oldPkgSlash, newPkgSlash))
		outPath := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		data, err := templates.ReadFile(p)
		if err != nil {
			return err
		}

		content := replacer.Replace(string(data))
		if opts.teamID != "" && path.Base(p) == "Config.xcconfig" {
			content = strings.Replace(content, "TEAM_ID=\n", "TEAM_ID="+opts.teamID+"\n", 1)
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
