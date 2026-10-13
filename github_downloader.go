/*
 * SPDX-License-Identifier: GPL-3.0
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bufio"

	"fmt"
	"os"
	path "path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var ReleaseData GithubRelease
var GithubError error
var GithubDoneChan chan bool
var GithubFinished atomic.Bool

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("VENCORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubFinished.Store(true)
		GithubDoneChan <- true
		return
	}

	readInstalledHash()

	go func() {

		defer func() {
			GithubFinished.Store(true)
			GithubDoneChan <- GithubError == nil
		}()

		data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
		if err != nil {
			GithubError = err
			return
		}

		ReleaseData = *data

		i := strings.LastIndex(data.Name, " ") + 1
		LatestHash = data.Name[i:]
		Log.Debug("Finished fetching GitHub Data")
		Log.Debug("Latest hash is", LatestHash, "Local Install is", Ternary(LatestHash == InstalledHash, "up to date!", "outdated!"))
	}()

}

func readInstalledHash() {
	f, err := os.Open(Patcher)
	if err != nil {
		return
	}

	defer f.Close()

	Log.Debug("Found existing Vencord Arab Install. Checking for hash...")
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "// Vencord ") {
			InstalledHash = line[11:]
			Log.Debug("Existing hash is", InstalledHash)
		} else {
			Log.Debug("Didn't find hash")
		}
	}

}

func refreshAndInstallLatestBuilds() error {
	data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
	if err != nil {
		return err
	}
	return installReleaseBuilds(*data)
}

func installLatestBuilds() error {
	return installReleaseBuilds(ReleaseData)
}

func installReleaseBuilds(release GithubRelease) error {

	required := map[string]bool{"patcher.js": false, "preload.js": false, "renderer.js": false, "renderer.css": false}
	assets := make(map[string]string)
	for _, asset := range release.Assets {
		_, needed := required[asset.Name]
		_, sourceMap := required[strings.TrimSuffix(asset.Name, ".map")]
		if !needed && !(sourceMap && strings.HasSuffix(asset.Name, ".map")) {
			continue
		}
		if _, duplicate := assets[asset.Name]; duplicate {
			return fmt.Errorf("Duplicate release asset: %s", asset.Name)
		}
		assets[asset.Name] = asset.DownloadURL
		if needed {
			required[asset.Name] = true
		}
	}
	for name, found := range required {
		if !found {
			return fmt.Errorf("Missing required release asset: %s", name)
		}
	}
	stage, err := os.MkdirTemp(FilesDir, ".download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = os.WriteFile(path.Join(stage, "package.json"), []byte("{}"), 0644); err != nil {
		return err
	}
	var wg sync.WaitGroup
	failures := make(chan error, len(assets))
	for name, url := range assets {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			if err := downloadFile(url, path.Join(stage, name), 0644); err != nil {
				failures <- fmt.Errorf("%s: %w", name, err)
			}
		}(name, url)
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		return err
	}
	names := []string{"package.json"}
	for name := range assets {
		names = append(names, name)
	}
	if err = replaceStagedFiles(stage, FilesDir, names); err != nil {
		return err
	}
	_ = FixOwnership(FilesDir)
	InstalledHash = release.Name[strings.LastIndex(release.Name, " ")+1:]
	return nil
}
