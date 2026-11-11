// Package cli parses gen-android's command line and dispatches to its commands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/paulcoding810/android-template/internal/generate"
	"github.com/paulcoding810/android-template/internal/icon"
	"github.com/paulcoding810/android-template/internal/version"
	"github.com/paulcoding810/android-template/templates"
)

// ErrReported marks errors the flag package has already printed.
var ErrReported = errors.New("already reported")

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "icon" {
		return runIcon(args[1:], stdout, stderr)
	}
	return runCreate(args, stdout, stderr)
}

func runCreate(args []string, stdout, stderr io.Writer) error {
	fset := flag.NewFlagSet("gen-android", flag.ContinueOnError)
	fset.SetOutput(stderr)
	single := fset.Bool("single", false, "use the single-module template")
	kmm := fset.Bool("kmm", false, "use the Kotlin Multiplatform template")
	list := fset.Bool("list", false, "list available templates and exit")
	showVersion := fset.Bool("version", false, "print the gen-android version and exit")
	outDir := fset.String("output", ".", "parent directory for the new project")
	teamID := fset.String("team-id", "", "Apple team ID for the iOS app (kmm only)")
	noGit := fset.Bool("no-git", false, "skip git init and the initial commit")
	iconPath := fset.String("icon", "", "SVG file for the launcher icon")
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
		return ErrReported
	}

	if *showVersion {
		fmt.Fprintf(stdout, "gen-android %s\n", version.String())
		return nil
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

	opts := generate.Options{
		Package:  positional[0],
		Project:  positional[1],
		Template: templates.Default,
		OutDir:   *outDir,
		TeamID:   *teamID,
		NoGit:    *noGit,
	}
	if *iconPath != "" {
		iconOpts, err := icon.Load(*iconPath, *iconBg, icon.DefaultScale)
		if err != nil {
			return err
		}
		opts.Icon = &iconOpts
	}
	if *single {
		opts.Template = "single"
	} else if *kmm {
		opts.Template = "kmm"
	}

	return generate.Project(opts, stdout, stderr)
}

func runIcon(args []string, stdout, stderr io.Writer) error {
	fset := flag.NewFlagSet("gen-android icon", flag.ContinueOnError)
	fset.SetOutput(stderr)
	dir := fset.String("dir", ".", "project directory")
	bg := fset.String("bg", "#FFFFFF", "background color (#RGB or #RRGGBB)")
	scale := fset.Float64("scale", icon.DefaultScale, "fraction of the 108dp adaptive canvas the artwork fills")
	fset.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gen-android icon <icon.svg> [flags]\n\nFlags:\n")
		fset.PrintDefaults()
	}

	positional, err := parseInterspersed(fset, args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return ErrReported
	}
	if len(positional) != 1 {
		fset.Usage()
		return fmt.Errorf("expected 1 argument, got %d", len(positional))
	}

	opts, err := icon.Load(positional[0], *bg, *scale)
	if err != nil {
		return err
	}
	if err := icon.Apply(*dir, opts); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Launcher icon updated in %s\n", *dir)
	return nil
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
	for _, t := range templates.All {
		flagName := "(default)"
		if t.Name != templates.Default {
			flagName = "--" + t.Name
		}
		fmt.Fprintf(w, "%-7s %-10s %s\n", t.Name, flagName, t.Description)
	}
}
