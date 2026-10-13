/*
 * SPDX-License-Identifier: GPL-3.0
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"errors"
	"fmt"
	"os"
	path "path/filepath"
	"runtime"
	"sync/atomic"
	"time"
	"vencordinstaller/buildinfo"
)

var IsSelfOutdated = false
var SelfUpdateFinished atomic.Bool
var SelfUpdateCheckDoneChan = make(chan bool, 1)

func init() {

	if buildinfo.InstallerTag == buildinfo.VersionUnknown {
		Log.Debug("Disabling self updater as this is not a release build")
		SelfUpdateFinished.Store(true)
		SelfUpdateCheckDoneChan <- false
		return
	}

	go DeleteOldExecutable()

	go func() {
		Log.Debug("Checking for Installer Updates...")

		res, err := GetGithubRelease(InstallerReleaseUrl, InstallerReleaseUrlFallback)
		if err != nil {
			Log.Warn("Failed to check for self updates:", err)
			SelfUpdateFinished.Store(true)
			SelfUpdateCheckDoneChan <- false
		} else {
			IsSelfOutdated = res.TagName != buildinfo.InstallerTag
			Log.Debug("Is self outdated?", IsSelfOutdated)
			SelfUpdateFinished.Store(true)
			SelfUpdateCheckDoneChan <- true
		}
	}()
}

func GetInstallerDownloadLink() string {
	const BaseUrl = "https://github.com/viciouscal/Installer/releases/latest/download/"
	switch runtime.GOOS {
	case "windows":
		filename := Ternary(buildinfo.UiType == buildinfo.UiTypeCli, "VencordInstallerCli.exe", "VencordArInstaller.exe")
		return BaseUrl + filename
	case "darwin":
		return BaseUrl + "VencordInstaller.dmg"
	case "linux":
		return BaseUrl + "VencordInstallerCli-linux"
	default:
		return ""
	}
}

func CanUpdateSelf() bool {

	return SelfUpdateFinished.Load() && IsSelfOutdated && runtime.GOOS != "darwin"
}

func UpdateSelf() error {
	if !CanUpdateSelf() {
		return errors.New("Cannot update self. Either no update available or macos")
	}

	url := GetInstallerDownloadLink()
	if url == "" {
		return errors.New("Failed to get installer download link")
	}

	Log.Debug("Updating self from", url)

	ownExePath, err := os.Executable()
	if err != nil {
		return err
	}

	ownExeDir := path.Dir(ownExePath)

	stage, err := os.MkdirTemp(ownExeDir, ".VencordAR-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	stagedExe := path.Join(stage, path.Base(ownExePath))
	if err = downloadFile(url, stagedExe, 0755); err != nil {
		return err
	}
	if _, err = os.Lstat(ownExePath + ".old"); err == nil {
		return errors.New("Previous installer backup still exists. Close the previous installer and try again")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(ownExePath, ownExePath+".old"); err != nil {
		return err
	}
	if err = os.Rename(stagedExe, ownExePath); err != nil {
		if restoreErr := os.Rename(ownExePath+".old", ownExePath); restoreErr != nil {
			return fmt.Errorf("Update failed: %w; restore failed: %v; original executable: %s.old", err, restoreErr, ownExePath)
		}
		return err
	}

	return nil
}

func DeleteOldExecutable() {
	ownExePath, err := os.Executable()
	if err != nil {
		return
	}

	for attempts := 0; attempts < 10; attempts += 1 {
		err = os.Remove(ownExePath + ".old")

		if err == nil || errors.Is(err, os.ErrNotExist) {
			break
		}

		Log.Warn("Failed to remove old executable. Retrying in 1 second.", err)
		time.Sleep(1 * time.Second)
	}
}

func RelaunchSelf() error {
	attr := new(os.ProcAttr)
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	ownExePath, err := os.Executable()
	if err != nil {
		return err
	}
	argv := append([]string{ownExePath}, os.Args[1:]...)
	Log.Debug("Restarting self with exe", ownExePath, "and args", argv)
	proc, err := os.StartProcess(ownExePath, argv, attr)
	if err != nil {
		return fmt.Errorf("Failed to start new process: %w", err)
	}

	if err = proc.Release(); err != nil {
		return fmt.Errorf("Failed to release new process: %w", err)
	}

	os.Exit(0)
	return nil
}
