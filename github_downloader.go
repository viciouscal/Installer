/*
 * SPDX-License-Identifier: GPL-3.0
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bufio"

	"errors"

	"io"
	"net/http"
	"os"
	path "path/filepath"
	"strconv"
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

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("VENCORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	go func() {

		defer func() {
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

func installLatestBuilds() (retErr error) {

	Log.Debug("Installing latest builds...")

	pkgJsonFile := path.Join(FilesDir, "package.json")
	err := os.WriteFile(pkgJsonFile, []byte("{}"), 0644)
	if err != nil {
		Log.Warn("Failed to create", pkgJsonFile, err)
	}

	var wg sync.WaitGroup
	var downloadedFiles atomic.Uint64

	for _, ass := range ReleaseData.Assets {
		if strings.HasPrefix(ass.Name, "patcher.js") ||
			strings.HasPrefix(ass.Name, "preload.js") ||
			strings.HasPrefix(ass.Name, "renderer.js") ||
			strings.HasPrefix(ass.Name, "renderer.css") {
			wg.Add(1)
			ass := ass
			go func() {
				defer wg.Done()
				Log.Debug("Downloading file", ass.Name)

				res, err := http.Get(ass.DownloadURL)
				if err == nil && res.StatusCode >= 300 {
					err = errors.New(res.Status)
				}
				if err != nil {
					Log.Error("Failed to download", ass.Name+":", err)
					retErr = err
					return
				}
				outFile := path.Join(FilesDir, ass.Name)
				out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					Log.Error("Failed to create", outFile+":", err)
					retErr = err
					return
				}
				read, err := io.Copy(out, res.Body)
				if err != nil {
					Log.Error("Failed to download to", outFile+":", err)
					retErr = err
					return
				}
				contentLength := res.Header.Get("Content-Length")
				expected := strconv.FormatInt(read, 10)
				if expected != contentLength {
					err = errors.New("Unexpected end of input. Content-Length was " + contentLength + ", but I only read " + expected)
					Log.Error(err.Error())
					retErr = err
					return
				}
				downloadedFiles.Add(1)
			}()
		}
	}

	wg.Wait()

	if retErr != nil {
		return retErr
	}
	if downloadedFiles.Load() < 4 {
		return errors.New("Couldn't find all required files")
	}

	Log.Debug("Done!")
	_ = FixOwnership(FilesDir)

	InstalledHash = LatestHash
	return
}
