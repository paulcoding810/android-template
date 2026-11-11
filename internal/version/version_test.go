package version

import "testing"

func TestStringPrefersStampedVersion(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })

	Version = ""
	if String() == "" {
		t.Error("String must never be empty")
	}
	Version = "v1.2.3"
	if got := String(); got != "v1.2.3" {
		t.Errorf("String() = %q, want v1.2.3", got)
	}
}
