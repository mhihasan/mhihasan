package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary compiles the binary into a temp dir and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "metrics-collector")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/")
	cmd.Dir = filepath.Join("..")  // module root = tools/metrics-collector/
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestMissingGHTokenExitsNonZero(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "GH_TOKEN=", "USERNAME=testuser")
	// Remove any real GH_TOKEN from env
	filtered := []string{"USERNAME=testuser"}
	for _, e := range os.Environ() {
		if len(e) >= 8 && e[:8] != "GH_TOKEN" {
			filtered = append(filtered, e)
		}
	}
	cmd.Env = filtered
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit when GH_TOKEN is missing")
	}
}

func TestMissingUsernameExitsNonZero(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	filtered := []string{"GH_TOKEN=fake-token"}
	for _, e := range os.Environ() {
		if len(e) >= 8 && e[:8] != "USERNAME" && e[:8] != "GH_TOKEN" {
			filtered = append(filtered, e)
		}
	}
	cmd.Env = filtered
	err := cmd.Run()
	if err == nil {
		t.Error("expected non-zero exit when USERNAME is missing")
	}
}
