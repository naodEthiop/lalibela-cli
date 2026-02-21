package features

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallFeatureIdempotent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	result, err := InstallFeature(root, "gin", "config", nil)
	if err != nil {
		t.Fatalf("install first pass: %v", err)
	}
	if !result.Installed {
		t.Fatalf("expected Installed=true on first pass")
	}

	result, err = InstallFeature(root, "gin", "config", nil)
	if err != nil {
		t.Fatalf("install second pass: %v", err)
	}
	if !result.AlreadyPresent {
		t.Fatalf("expected AlreadyPresent=true on second pass")
	}
}

func TestInstallFeatureCompatibility(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	result, err := InstallFeature(root, "nethttp", "auth", nil)
	if err != nil {
		t.Fatalf("install result should not error for incompatible feature: %v", err)
	}
	if result.Compatible {
		t.Fatalf("expected feature to be incompatible for nethttp")
	}
}
