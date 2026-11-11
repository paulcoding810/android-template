// Package templates embeds the project templates shipped with gen-android.
package templates

import "embed"

//go:embed all:main all:single all:kmm
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
	{"kmm", "Kotlin Multiplatform app (Android, iOS, JVM)"},
}
