package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	for _, scenario := range []string{"ok", "chunked", "forbidden", "short", "empty"} {
		t.Run(scenario, func(t *testing.T) {
			oldClient := downloadClient
			t.Cleanup(func() { downloadClient = oldClient })
			downloadClient = &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
				code, body, length := 200, "complete", int64(8)
				switch scenario {
				case "forbidden":
					code = 403
				case "empty":
					body, length = "", 0
				case "short":
					body, length = "partial", 100
				case "chunked":
					length = -1
				}
				res := testResponse(r, code, body)
				res.ContentLength = length
				return res, nil
			})}
			target := filepath.Join(t.TempDir(), "download")
			err := downloadFile("https://downloads.example/test", target, 0644)
			valid := scenario == "ok" || scenario == "chunked"
			if valid {
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(target)
				if err != nil || string(data) != "complete" {
					t.Fatalf("invalid download: %q %v", data, err)
				}
			} else {
				if err == nil {
					t.Fatal("invalid download accepted")
				}
				if _, err := os.Stat(target); !os.IsNotExist(err) {
					t.Fatal("failed download retained")
				}
			}
		})
	}
}

func TestReplacementRollback(t *testing.T) {
	dir := t.TempDir()
	stage := t.TempDir()
	for _, name := range []string{"first", "second"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(stage, "first"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := replaceStagedFiles(stage, dir, []string{"first", "second"}); err == nil {
		t.Fatal("missing staged file accepted")
	}
	for _, name := range []string{"first", "second"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(data) != "original" {
			t.Fatalf("original lost: %s %q %v", name, data, err)
		}
	}
}

func TestInstallDownloadFailurePreservesExistingFiles(t *testing.T) {
	oldDir, oldRelease, oldHash := FilesDir, ReleaseData, InstalledHash
	t.Cleanup(func() { FilesDir, ReleaseData, InstalledHash = oldDir, oldRelease, oldHash })
	FilesDir = t.TempDir()
	InstalledHash = "original"
	oldClient := downloadClient
	t.Cleanup(func() { downloadClient = oldClient })
	downloadClient = &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		status := 200
		if strings.Contains(r.URL.Path, "preload") {
			status = 403
		}
		res := testResponse(r, status, "new")
		res.ContentLength = 3
		return res, nil
	})}
	ReleaseData = GithubRelease{}
	for _, name := range []string{"patcher.js", "preload.js", "renderer.js", "renderer.css", "package.json"} {
		if err := os.WriteFile(filepath.Join(FilesDir, name), []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}
		if name != "package.json" {
			ReleaseData.Assets = append(ReleaseData.Assets, struct {
				Name        string `json:"name"`
				DownloadURL string `json:"browser_download_url"`
			}{name, "https://downloads.example/" + name})
		}
	}
	if err := installLatestBuilds(); err == nil {
		t.Fatal("failed download accepted")
	}
	if InstalledHash != "original" {
		t.Fatal("installed hash changed")
	}
	for _, name := range []string{"patcher.js", "preload.js", "renderer.js", "renderer.css", "package.json"} {
		data, err := os.ReadFile(filepath.Join(FilesDir, name))
		if err != nil || string(data) != "original" {
			t.Fatalf("original lost: %s %q %v", name, data, err)
		}
	}
	ReleaseData.Assets = ReleaseData.Assets[:3]
	if err := installLatestBuilds(); err == nil {
		t.Fatal("missing required file accepted")
	}
}

func TestUnpatchRollback(t *testing.T) {
	dir := t.TempDir()
	for name, value := range map[string]string{"app.asar": "wrapper", "_app.asar": "original"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := unpatchAppAsar(dir, true); err == nil {
		t.Fatal("missing unpacked directory accepted")
	}
	for name, value := range map[string]string{"app.asar": "wrapper", "_app.asar": "original"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(data) != value {
			t.Fatalf("rollback lost %s: %q %v", name, data, err)
		}
	}
}

func TestStartupGlyphCoverage(t *testing.T) {
	seed := uiGlyphSeed()
	for key, text := range arabicTranslations {
		for _, r := range visualArabic(text) {
			if r != '\n' && r != '\r' && !strings.ContainsRune(seed, r) {
				t.Fatalf("missing startup glyph %U for %s", r, key)
			}
		}
	}
}

func TestPatchPreservesOriginalBackup(t *testing.T) {
	dir := t.TempDir()
	for name, value := range map[string]string{"app.asar": "current", "_app.asar": "original"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := patchAppAsar(dir, false); err == nil {
		t.Fatal("existing original backup overwritten")
	}
	for name, value := range map[string]string{"app.asar": "current", "_app.asar": "original"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(data) != value {
			t.Fatalf("lost %s: %q %v", name, data, err)
		}
	}
}

func TestPatchRollbackOnMissingUnpacked(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "app.asar")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := patchAppAsar(dir, true); err == nil {
		t.Fatal("missing unpacked directory accepted")
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Fatalf("original lost: %q %v", data, err)
	}
}

func TestReinstallFetchesFreshReleaseEveryTime(t *testing.T) {
	oldReleaseClient, oldDownloadClient := releaseClient, downloadClient
	oldDir, oldRelease, oldHash := FilesDir, ReleaseData, InstalledHash
	t.Cleanup(func() {
		releaseClient, downloadClient = oldReleaseClient, oldDownloadClient
		FilesDir, ReleaseData, InstalledHash = oldDir, oldRelease, oldHash
	})
	FilesDir = t.TempDir()
	ReleaseData = GithubRelease{Name: "DevBuild stale"}
	generation := 0
	releaseClient = &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		generation++
		if r.Header.Get("Cache-Control") == "" {
			t.Fatal("metadata request may reuse cached response")
		}
		release := GithubRelease{Name: fmt.Sprintf("DevBuild fresh%d", generation), TagName: "devbuild"}
		for _, name := range []string{"patcher.js", "preload.js", "renderer.js", "renderer.css"} {
			release.Assets = append(release.Assets, struct {
				Name        string `json:"name"`
				DownloadURL string `json:"browser_download_url"`
			}{name, "https://downloads.example/" + name})
		}
		body, _ := json.Marshal(release)
		return testResponse(r, 200, string(body)), nil
	})}
	downloadClient = &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("_vencordar") == "" {
			t.Fatal("download URL has no cache bypass")
		}
		body := fmt.Sprintf("fresh%d", generation)
		res := testResponse(r, 200, body)
		res.ContentLength = int64(len(body))
		return res, nil
	})}
	for attempt := 1; attempt <= 2; attempt++ {
		if err := refreshAndInstallLatestBuilds(); err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("fresh%d", attempt)
		data, err := os.ReadFile(filepath.Join(FilesDir, "patcher.js"))
		if err != nil || string(data) != expected || InstalledHash != expected {
			t.Fatalf("reinstall reused old build: %q %s %v", data, InstalledHash, err)
		}
	}
	if generation != 2 {
		t.Fatalf("release fetched %d times", generation)
	}
}
