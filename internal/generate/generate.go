// Package generate creates a new project from an embedded template.
package generate

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/paulcoding810/android-template/internal/icon"
	"github.com/paulcoding810/android-template/internal/version"
	"github.com/paulcoding810/android-template/templates"
)

const (
	oldPkg       = "com.paulcoding.androidtemplate"
	oldProjName  = "AndroidTemplate"
	oldProjLower = "androidtemplate"
	// Spellings used for display text and file names (window titles, database paths).
	oldDisplayName = "Android Template"
	oldProjKebab   = "android-template"
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

// Bounds for --min-sdk; Gradle reports libraries that need a higher value.
const (
	minSupportedSDK = 23
	maxSupportedSDK = 36
)

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
	AppName  string // user-visible name; defaults to Project
	Template string
	OutDir   string
	TeamID   string
	MinSDK   int  // 0 keeps the template default
	Here     bool // write into OutDir itself, which may already exist, instead of OutDir/Project
	NoGit    bool
	Icon     *icon.Options
}

// Project validates opts and writes the new project to OutDir/Project (or OutDir with Here).
func Project(opts Options, stdout, stderr io.Writer) error {
	if opts.AppName == "" {
		opts.AppName = opts.Project
	}
	if err := validate(opts); err != nil {
		return err
	}

	files, err := render(opts)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	dir := filepath.Join(opts.OutDir, opts.Project)
	if opts.Here {
		dir = opts.OutDir
	}
	w, err := newWriter(dir, opts.Here)
	if err != nil {
		return err
	}
	if err := w.writeAll(files); err != nil {
		w.rollback()
		return err
	}

	if opts.Icon != nil {
		if err := icon.Apply(dir, *opts.Icon); err != nil {
			w.rollback()
			return fmt.Errorf("generating icon: %w", err)
		}
	}

	if !opts.NoGit {
		if insideWorkTree(dir) {
			fmt.Fprintf(stdout, "%s is inside an existing git repository; the files were left uncommitted.\n", dir)
		} else if err := gitInit(dir); err != nil {
			fmt.Fprintf(stderr, "Warning: git init failed: %v\n", err)
		}
	}

	fmt.Fprintf(stdout, "\nDone! Project created at: %s\n", dir)
	return nil
}

func validate(opts Options) error {
	if strings.TrimSpace(opts.AppName) == "" || strings.IndexFunc(opts.AppName, unicode.IsControl) >= 0 {
		return fmt.Errorf("invalid app name %q: must be non-empty text on one line", opts.AppName)
	}
	if !templates.Exists(opts.Template) {
		return fmt.Errorf("unknown template %q (available: %s)", opts.Template, templates.Names())
	}
	if err := validatePackage(opts.Package); err != nil {
		return err
	}
	if !projNameRe.MatchString(opts.Project) {
		return fmt.Errorf("invalid project name %q: use letters and digits only, starting with a letter", opts.Project)
	}
	if opts.MinSDK != 0 && (opts.MinSDK < minSupportedSDK || opts.MinSDK > maxSupportedSDK) {
		return fmt.Errorf("invalid minSdk %d: must be between %d and %d", opts.MinSDK, minSupportedSDK, maxSupportedSDK)
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

// render produces every output file in memory, so conflicts are found before anything is written.
func render(opts Options) ([]file, error) {
	newPkgSlash := strings.ReplaceAll(opts.Package, ".", "/")
	oldPkgSlash := strings.ReplaceAll(oldPkg, ".", "/")
	// The display name needs file-specific escaping, so keep one replacer per extension.
	replacers := map[string]*strings.Replacer{}
	replacerFor := func(p string) *strings.Replacer {
		ext := path.Ext(p)
		if r, ok := replacers[ext]; ok {
			return r
		}
		r := strings.NewReplacer(
			oldPkg, opts.Package,
			oldProjName, opts.Project,
			oldProjLower, strings.ToLower(opts.Project),
			oldDisplayName, escapeDisplayName(ext, opts.AppName),
			oldProjKebab, strings.ToLower(opts.Project),
		)
		replacers[ext] = r
		return r
	}

	var files []file
	err := fs.WalkDir(templates.FS, opts.Template, func(p string, d fs.DirEntry, err error) error {
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
		rel = replacerFor("").Replace(strings.ReplaceAll(rel, oldPkgSlash, newPkgSlash))

		data, err := templates.FS.ReadFile(p)
		if err != nil {
			return err
		}

		content := replacerFor(p).Replace(string(data))
		if rel == "gradle.properties" {
			content = generatedBy(opts.Template) + content
		}
		if rel == templates.Catalog && opts.MinSDK != 0 {
			if !templates.MinSDKPattern.MatchString(content) {
				return fmt.Errorf("template %q has no android-minSdk in %s", opts.Template, templates.Catalog)
			}
			content = templates.MinSDKPattern.ReplaceAllString(content, fmt.Sprintf(`android-minSdk = "%d"`, opts.MinSDK))
		}
		if opts.TeamID != "" && path.Base(p) == "Config.xcconfig" {
			content = strings.Replace(content, "TEAM_ID=\n", "TEAM_ID="+opts.TeamID+"\n", 1)
		}

		perm := fs.FileMode(0644)
		if d.Name() == "gradlew" {
			perm = 0755
		}
		files = append(files, file{rel: rel, data: []byte(content), perm: perm})
		return nil
	})
	return files, err
}

// escapeDisplayName makes the app name a valid literal in the file types that show it.
func escapeDisplayName(ext, name string) string {
	switch ext {
	case ".xml": // Android string resource
		name = strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(name)
		if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "?") {
			name = `\` + name
		}
		return name
	case ".plist":
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(name)
	case ".kt", ".kts":
		return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "$", `\$`).Replace(name)
	}
	return name
}

// generatedBy records the generator build, so a project can later be compared with newer templates.
func generatedBy(template string) string {
	return fmt.Sprintf("# Generated by gen-android %s (%s template)\n", version.String(), template)
}

func insideWorkTree(dir string) bool {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func gitInit(dir string) error {
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "."},
		{"commit", "-q", "-m", "Initial commit"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %w\n%s", args[0], err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}
