/*
 * SPDX-License-Identifier: GPL-3.0
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	path "path/filepath"
)

const OpenAsarDownloadLink = "https://github.com/GooseMod/OpenAsar/releases/download/nightly/app.asar"

func FindAsarFile(dir string) (*os.File, error) {
	for _, file := range []string{"_app.asar", "app.asar"} {
		f, err := os.Open(path.Join(dir, file))
		if err != nil {
			continue
		}
		stats, err := f.Stat()
		if err == nil && !stats.IsDir() {
			return f, nil
		}
		_ = f.Close()
	}
	return nil, errors.New("Install at " + dir + " has no asar file")
}

func (di *DiscordInstall) IsOpenAsar() (retBool bool) {
	if di.isOpenAsar != nil {
		return *di.isOpenAsar
	}

	defer func() {
		Log.Debug("Checking if", di.path, "is using OpenAsar:", retBool)
		di.isOpenAsar = &retBool
	}()

	asarFile, err := FindAsarFile(path.Join(di.appPath, ".."))
	if err != nil {
		Log.Error(err.Error())
		return false
	}

	b, err := io.ReadAll(asarFile)
	_ = asarFile.Close()
	if err != nil {
		Log.Error(err.Error())
		return false
	}

	if bytes.Contains(b, []byte("OpenAsar")) {
		return true
	}

	return false
}

func (di *DiscordInstall) InstallOpenAsar() error {
	PreparePatch(di)

	dir := path.Join(di.appPath, "..")
	asarFile, err := FindAsarFile(dir)
	if err != nil {
		return err
	}
	_ = asarFile.Close()

	stage, err := os.MkdirTemp(dir, ".OpenAsar-download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	stagedAsar := path.Join(stage, "app.asar")
	if err = downloadFile(OpenAsarDownloadLink, stagedAsar, 0644); err != nil {
		return err
	}
	backup := path.Join(dir, "app.asar.backup")
	if _, err = os.Lstat(backup); err == nil {
		return errors.New("An app.asar backup already exists. Restore it before installing OpenAsar")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(asarFile.Name(), backup); err != nil {
		return err
	}
	if err = os.Rename(stagedAsar, asarFile.Name()); err != nil {
		if restoreErr := os.Rename(backup, asarFile.Name()); restoreErr != nil {
			return fmt.Errorf("Install failed: %w; restore failed: %v; backup: %s", err, restoreErr, backup)
		}
		return err
	}

	di.isOpenAsar = Ptr(true)
	return nil
}

func (di *DiscordInstall) UninstallOpenAsar() error {
	PreparePatch(di)

	dir := path.Join(di.appPath, "..")

	for _, file := range []string{path.Join(dir, "app.asar.backup"), path.Join(dir, "app.asar.original")} {
		if !ExistsFile(file) {
			continue
		}

		asarFile, err := FindAsarFile(dir)
		if err != nil {
			return err
		}
		_ = asarFile.Close()

		stage, err := os.MkdirTemp(dir, ".OpenAsar-restore-")
		if err != nil {
			return err
		}
		defer func() {
			if ExistsFile(path.Join(stage, "app.asar")) {
				return
			}
			os.RemoveAll(stage)
		}()
		saved := path.Join(stage, "app.asar")
		if err = os.Rename(asarFile.Name(), saved); err != nil {
			return err
		}
		if err = os.Rename(file, asarFile.Name()); err != nil {
			if restoreErr := os.Rename(saved, asarFile.Name()); restoreErr != nil {
				return fmt.Errorf("Restore failed: %w; rollback failed: %v; backup: %s", err, restoreErr, saved)
			}
			return err
		}

		_ = os.Remove(saved)
		di.isOpenAsar = Ptr(false)
		return nil
	}

	return errors.New("No app.asar.backup. Reinstall Discord")
}
