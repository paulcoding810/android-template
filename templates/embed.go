// Package templates embeds the project templates shipped with gen-android.
package templates

import (
	"embed"
	"regexp"
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

// Catalog is each template's version catalog; its android-minSdk entry is the single source of minSdk.
const Catalog = "gradle/libs.versions.toml"

var MinSDKPattern = regexp.MustCompile(`(?m)^android-minSdk = "(\d+)"`)

// MinSDK returns the template's default minSdk, or "" if it can't be read.
func (t Template) MinSDK() string {
	data, err := FS.ReadFile(t.Name + "/" + Catalog)
	if err != nil {
		return ""
	}
	m := MinSDKPattern.FindSubmatch(data)
	if m == nil {
		return ""
	}
	return string(m[1])
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
