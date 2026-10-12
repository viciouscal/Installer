package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var downloadClient = &http.Client{Timeout: 2 * time.Minute}

func downloadFile(url, destination string, mode os.FileMode) (errOut error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Download failed: %s", res.Status)
	}
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if errOut != nil {
			os.Remove(destination)
		}
	}()
	count, err := io.Copy(out, res.Body)
	if err != nil {
		return err
	}
	if count == 0 || (res.ContentLength >= 0 && count != res.ContentLength) {
		return fmt.Errorf("Incomplete download: received %d bytes", count)
	}
	if err = out.Sync(); err != nil {
		return err
	}
	return out.Close()
}

func replaceStagedFiles(stage, destination string, names []string) (errOut error) {
	backup, err := os.MkdirTemp(destination, ".backup-")
	if err != nil {
		return err
	}
	type change struct {
		name             string
		saved, installed bool
	}
	changes := make([]change, 0, len(names))
	defer func() {
		if errOut != nil {
			for i := len(changes) - 1; i >= 0; i-- {
				c := changes[i]
				target := filepath.Join(destination, c.name)
				if c.installed {
					if err := os.Remove(target); err != nil {
						errOut = fmt.Errorf("%w; rollback removal failed: %v; backup: %s", errOut, err, backup)
						return
					}
				}
				if c.saved {
					if err := os.Rename(filepath.Join(backup, c.name), target); err != nil {
						errOut = fmt.Errorf("%w; rollback failed: %v; backup: %s", errOut, err, backup)
						return
					}
				}
			}
		}
		os.RemoveAll(backup)
	}()
	for _, name := range names {
		if filepath.Base(name) != name || name == "." || name == ".." {
			return fmt.Errorf("Invalid filename: %s", name)
		}
		changes = append(changes, change{name: name})
		c := &changes[len(changes)-1]
		target := filepath.Join(destination, name)
		if _, err = os.Lstat(target); err == nil {
			if err = os.Rename(target, filepath.Join(backup, name)); err != nil {
				return err
			}
			c.saved = true
		} else if !os.IsNotExist(err) {
			return err
		}
		if err = os.Rename(filepath.Join(stage, name), target); err != nil {
			return err
		}
		c.installed = true
	}
	return nil
}
