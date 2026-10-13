//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverySkipsInvalidNewerAppFolder(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "app-1.0.1", "resources")
	invalid := filepath.Join(root, "app-1.0.9", "resources")
	for _, dir := range []string{valid, invalid} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"app.asar", "_app.asar"} {
		if err := os.WriteFile(filepath.Join(valid, name), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	install := ParseDiscord(root, "stable")
	if install == nil || install.resourcesPath != valid || !install.isPatched {
		t.Fatalf("selected invalid app folder: %#v", install)
	}
	if err := os.Remove(filepath.Join(valid, "app.asar")); err != nil {
		t.Fatal(err)
	}
	if install := ParseDiscord(root, "stable"); install != nil {
		t.Fatalf("invalid installation accepted: %#v", install)
	}
}
