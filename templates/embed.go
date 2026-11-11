// Package templates embeds the project templates shipped with gen-android.
package templates

import (
	"embed"
	"strings"
)

//go:embed all:main all:single all:empty all:kmm
var FS embed.FS

// Default is the template used when no template flag is given.
const Default = "main"

type Template struct {
	Name        string
	Description string
}

// All lists the templates in display order; each Name is also its directory in FS.
var All = []Template{
	{"main", "multi-module Android app (app, core/*, feature/*)"},
	{"single", "single-module Android app"},
	{"empty", "minimal Compose app, no extra libraries"},
	{"kmm", "Kotlin Multiplatform app (Android, iOS, JVM)"},
}

func Exists(name string) bool {
	for _, t := range All {
		if t.Name == name {
			return true
		}
	}
	return false
}

// Names returns the template names as a comma-separated list.
func Names() string {
	names := make([]string, len(All))
	for i, t := range All {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}
