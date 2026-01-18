package version

import "testing"

func TestVersion(t *testing.T) {
	// Just verify that version variables are set (even if to defaults)
	if VERSION == "" {
		t.Error("VERSION should not be empty")
	}
	if COMMIT == "" {
		t.Error("COMMIT should not be empty")
	}
	if DATE == "" {
		t.Error("DATE should not be empty")
	}
}
