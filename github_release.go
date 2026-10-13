package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

var releaseClient = &http.Client{Timeout: 30 * time.Second}

func GetGithubRelease(apiURL, fallbackURL string) (*GithubRelease, error) {
	data, err := getReleaseJSON(apiURL)
	if err == nil {
		return data, nil
	}
	if fallbackURL == "" || fallbackURL == apiURL {
		return nil, err
	}
	Log.Warn("Failed to fetch release metadata:", err, "Trying", fallbackURL)
	data, fallbackErr := getReleasePage(fallbackURL)
	if fallbackErr != nil {
		return nil, fmt.Errorf("GitHub API: %v; release page fallback: %w", err, fallbackErr)
	}
	return data, nil
}

func releaseGET(address string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Cache-Control", "no-cache, no-store")
	req.Header.Set("Pragma", "no-cache")
	res, err := releaseClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("%s returned %s", address, res.Status)
	}
	return res, nil
}

func getReleaseJSON(address string) (*GithubRelease, error) {
	res, err := releaseGET(address)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var data GithubRelease
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&data); err != nil {
		return nil, err
	}
	if data.TagName == "" || len(data.Assets) == 0 {
		return nil, fmt.Errorf("release metadata has no tag or assets")
	}
	return &data, nil
}

var htmlTagPattern = regexp.MustCompile(`(?is)<(?:meta|a)\b[^>]*>`)
var htmlAttributePattern = regexp.MustCompile(`([\w:-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

func tagAttributes(tag string) map[string]string {
	attrs := make(map[string]string)
	for _, match := range htmlAttributePattern.FindAllStringSubmatch(tag, -1) {
		value := match[2]
		if value == "" {
			value = match[3]
		}
		attrs[strings.ToLower(match[1])] = html.UnescapeString(value)
	}
	return attrs
}

func getReleasePage(address string) (*GithubRelease, error) {
	base, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.Trim(base.Path, "/"), "/")
	if base.Scheme != "https" || base.Host != "github.com" || len(parts) != 4 || parts[2] != "releases" || parts[3] != "latest" {
		return nil, fmt.Errorf("invalid GitHub release fallback URL")
	}
	repoPath := "/" + parts[0] + "/" + parts[1]
	res, err := releaseGET(address)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	finalURL := res.Request.URL
	tagPrefix := repoPath + "/releases/tag/"
	if finalURL.Scheme != base.Scheme || finalURL.Host != base.Host || !strings.HasPrefix(finalURL.Path, tagPrefix) {
		return nil, fmt.Errorf("latest release did not resolve to this fork's release tag")
	}
	tag := strings.TrimPrefix(finalURL.Path, tagPrefix)
	if tag == "" {
		return nil, fmt.Errorf("release page has no tag")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	data := &GithubRelease{TagName: tag, Name: tag}
	for _, element := range htmlTagPattern.FindAllString(string(body), -1) {
		attrs := tagAttributes(element)
		if attrs["property"] == "og:title" {
			title := strings.TrimSuffix(attrs["content"], " · "+parts[0]+"/"+parts[1])
			data.Name = strings.TrimPrefix(title, "Release ")
			break
		}
	}
	assetsURL := *base
	assetsURL.Path = repoPath + "/releases/expanded_assets/" + tag
	assetsURL.RawPath = ""
	assetsRes, err := releaseGET(assetsURL.String())
	if err != nil {
		return nil, err
	}
	defer assetsRes.Body.Close()
	assetsBody, err := io.ReadAll(io.LimitReader(assetsRes.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	prefix := repoPath + "/releases/download/" + tag + "/"
	seen := make(map[string]bool)
	for _, element := range htmlTagPattern.FindAllString(string(assetsBody), -1) {
		href := tagAttributes(element)["href"]
		link, err := url.Parse(href)
		if err != nil || href == "" {
			continue
		}
		link = base.ResolveReference(link)
		if link.Scheme != base.Scheme || link.Host != base.Host || !strings.HasPrefix(link.Path, prefix) || link.RawQuery != "" || link.Fragment != "" {
			continue
		}
		name := strings.TrimPrefix(link.Path, prefix)
		if name == "" || name == "." || name == ".." || path.Base(name) != name || strings.ContainsAny(name, `/\`) || seen[name] {
			continue
		}
		seen[name] = true
		data.Assets = append(data.Assets, struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{name, link.String()})
	}
	if len(data.Assets) == 0 {
		return nil, fmt.Errorf("release page has no downloadable assets")
	}
	return data, nil
}
