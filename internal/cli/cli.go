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

const createUsage = `Usage: gen-android <package.name> <ProjectName> [flags]
       gen-android icon <icon.svg> [--dir <project>] [--bg #RRGGBB] [--scale 0.61]

  package.name  application ID and Kotlin package, e.g. com.acme.notes: at least two
                lowercase segments of letters, digits or _, no Kotlin keywords
  ProjectName   letters and digits, used in code and folder names; set the name users
                see with --app-name

Generated projects record the gen-android version and template in gradle.properties.

Flags:
`

const iconUsage = `Usage: gen-android icon <icon.svg> [flags]

Replaces the launcher icon of an existing project (Android, plus iOS for kmm).
The SVG may not use text, images, masks, clip paths, filters or patterns.

Flags:
`

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "icon" {
		return runIcon(args[1:], stdout, stderr)
	}
	return runCreate(args, stdout, stderr)
}

func runCreate(args []string, stdout, stderr io.Writer) error {
	fset := flag.NewFlagSet("gen-android", flag.ContinueOnError)
	var template string
	templateUsage := "project template: " + templates.Names() + " (see --list)"
	fset.StringVar(&template, "template", templates.Default, templateUsage)
	fset.StringVar(&template, "t", templates.Default, "shorthand for --template")
	list := fset.Bool("list", false, "list templates with their minSdk and exit")
	showVersion := fset.Bool("version", false, "print the gen-android version and exit")
	var outDir string
	fset.StringVar(&outDir, "output", ".", "parent directory for the new project")
	fset.StringVar(&outDir, "o", ".", "shorthand for --output")
	appName := fset.String("app-name", "", "name shown under the app icon, may contain spaces (default: ProjectName)")
	minSDK := fset.Int("min-sdk", 0, "Android minSdk (default: the template's, see --list)")
	teamID := fset.String("team-id", "", "Apple team ID, written to iosApp/Configuration/Config.xcconfig; the iOS bundle ID becomes <package.name>.<ProjectName><team ID> (kmm only)")
	here := fset.Bool("here", false, "generate into the --output directory itself (e.g. a cloned repo) instead of a new ProjectName folder")
	noGit := fset.Bool("no-git", false, "skip git init and the initial commit")
	quiet := fset.Bool("quiet", false, "print only warnings and errors")
	iconPath := fset.String("icon", "", "SVG file for the launcher icon")
	iconBg := fset.String("icon-bg", "#FFFFFF", "launcher icon background color (#RGB or #RRGGBB)")
	iconScale := fset.Float64("icon-scale", icon.DefaultScale, "fraction of the 108dp adaptive icon canvas the artwork fills")

	positional, err := parseInterspersed(fset, args)
	if errors.Is(err, flag.ErrHelp) {
		printUsage(stdout, fset, createUsage)
		return nil
	}
	if err != nil {
		return usageError(err)
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
		return usageError(fmt.Errorf("expected <package.name> <ProjectName>, got %d argument(s)", len(positional)))
	}
	if !templates.Exists(template) {
		return fmt.Errorf("unknown template %q (available: %s)", template, templates.Names())
	}

	opts := generate.Options{
		Package:  positional[0],
		Project:  positional[1],
		AppName:  *appName,
		Template: template,
		OutDir:   outDir,
		TeamID:   *teamID,
		MinSDK:   *minSDK,
		Here:     *here,
		NoGit:    *noGit,
	}
	if *iconPath != "" {
		iconOpts, err := icon.Load(*iconPath, *iconBg, *iconScale)
		if err != nil {
			return err
		}
		opts.Icon = &iconOpts
	}
	if *quiet {
		stdout = io.Discard
	}
	return generate.Project(opts, stdout, stderr)
}

func runIcon(args []string, stdout, stderr io.Writer) error {
	fset := flag.NewFlagSet("gen-android icon", flag.ContinueOnError)
	dir := fset.String("dir", ".", "project directory")
	var bg string
	var scale float64
	// The --icon-* spellings match the create command's flags.
	fset.StringVar(&bg, "bg", "#FFFFFF", "background color (#RGB or #RRGGBB)")
	fset.StringVar(&bg, "icon-bg", "#FFFFFF", "same as --bg")
	fset.Float64Var(&scale, "scale", icon.DefaultScale, "fraction of the 108dp adaptive icon canvas the artwork fills")
	fset.Float64Var(&scale, "icon-scale", icon.DefaultScale, "same as --scale")

	positional, err := parseInterspersed(fset, args)
	if errors.Is(err, flag.ErrHelp) {
		printUsage(stdout, fset, iconUsage)
		return nil
	}
	if err != nil {
		return usageError(err)
	}
	if len(positional) != 1 {
		return usageError(fmt.Errorf("expected <icon.svg>, got %d argument(s)", len(positional)))
	}

	opts, err := icon.Load(positional[0], bg, scale)
	if err != nil {
		return err
	}
	if err := icon.Apply(*dir, opts); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Launcher icon updated in %s\n", *dir)
	return nil
}

func usageError(err error) error {
	return fmt.Errorf("%w\nRun 'gen-android -h' for usage.", err)
}

func printUsage(w io.Writer, fset *flag.FlagSet, header string) {
	fmt.Fprint(w, header)
	fset.SetOutput(w)
	fset.PrintDefaults()
}

// parseInterspersed lets flags appear before, between or after positional arguments.
func parseInterspersed(fset *flag.FlagSet, args []string) ([]string, error) {
	// Errors are returned to the caller instead of being printed with the full usage.
	fset.SetOutput(io.Discard)
	fset.Usage = func() {}
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
		marker := ""
		if t.Name == templates.Default {
			marker = "(default)"
		}
		fmt.Fprintf(w, "%-7s %-10s minSdk %-3s %s\n", t.Name, marker, t.MinSDK(), t.Description)
	}
}
