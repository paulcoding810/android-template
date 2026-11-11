// Package version reports which gen-android build is running.
package version

import "runtime/debug"

// Version is set at build time via -ldflags "-X github.com/paulcoding810/android-template/internal/version.Version=v1.2.3".
var Version = ""

// String returns the stamped version, the module version from `go install`, or the VCS revision of a local build.
func String() string {
	if Version != "" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if revision == "" {
		return "(devel)"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if modified == "true" {
		revision += "-dirty"
	}
	return "(devel) " + revision
}
