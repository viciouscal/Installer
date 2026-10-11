package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type releaseTransport func(*http.Request) (*http.Response, error)

func (f releaseTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)), Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}
}

func TestReleaseFallback(t *testing.T) {
	for _, code := range []int{401, 403, 429, 500} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			old := releaseClient.Transport
			t.Cleanup(func() { releaseClient.Transport = old })
			calls := 0
			releaseClient.Transport = releaseTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("User-Agent") != UserAgent {
					t.Fatal("missing user agent")
				}
				switch r.URL.String() {
				case ReleaseUrl:
					return testResponse(r, code, ""), nil
				case ReleaseUrlFallback:
					res := testResponse(r, 302, "")
					res.Header.Set("Location", "https://github.com/viciouscal/Vencord/releases/tag/devbuild")
					return res, nil
				case "https://github.com/viciouscal/Vencord/releases/tag/devbuild":
					return testResponse(r, 200, `<meta property="og:title" content="Release DevBuild abc1234 · viciouscal/Vencord">`), nil
				case "https://github.com/viciouscal/Vencord/releases/expanded_assets/devbuild":
					return testResponse(r, 200, `<a href="/viciouscal/Vencord/releases/download/devbuild/patcher.js">Patcher</a>
<a href='/viciouscal/Vencord/releases/download/devbuild/renderer.js'>Renderer</a>
<a href="/viciouscal/Vencord/releases/download/devbuild/patcher.js">Duplicate</a>
<a href="https://github.com/Vendicated/Vencord/releases/download/devbuild/preload.js">Wrong fork</a>
<a href="https://example.com/viciouscal/Vencord/releases/download/devbuild/renderer.css">Wrong host</a>
<a href="/viciouscal/Vencord/releases/download/devbuild/..%2Foutside.js">Invalid name</a>`), nil
				default:
					t.Fatalf("unexpected request: %s", r.URL)
					return nil, nil
				}
			})
			data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 4 || data.TagName != "devbuild" || data.Name != "DevBuild abc1234" || len(data.Assets) != 2 {
				t.Fatalf("incorrect release: %#v, requests=%d", data, calls)
			}
		})
	}
}

func TestReleaseAPISuccess(t *testing.T) {
	old := releaseClient.Transport
	t.Cleanup(func() { releaseClient.Transport = old })
	releaseClient.Transport = releaseTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != ReleaseUrl {
			t.Fatal("unexpected fallback")
		}
		return testResponse(r, 200, `{"name":"DevBuild abc1234","tag_name":"devbuild","assets":[{"name":"patcher.js","browser_download_url":"https://github.com/viciouscal/Vencord/releases/download/devbuild/patcher.js"}]}`), nil
	})
	data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
	if err != nil || data.Name != "DevBuild abc1234" {
		t.Fatalf("%#v %v", data, err)
	}
}

func TestReleaseFallbackErrors(t *testing.T) {
	for _, scenario := range []string{"network", "invalid-json", "empty-assets", "wrong-repo", "page-blocked"} {
		t.Run(scenario, func(t *testing.T) {
			old := releaseClient.Transport
			t.Cleanup(func() { releaseClient.Transport = old })
			releaseClient.Transport = releaseTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() == ReleaseUrl {
					if scenario == "network" {
						return nil, fmt.Errorf("network unavailable")
					}
					if scenario == "invalid-json" {
						return testResponse(r, 200, "invalid"), nil
					}
					return testResponse(r, 403, ""), nil
				}
				if scenario == "page-blocked" {
					return testResponse(r, 403, ""), nil
				}
				if r.URL.String() == ReleaseUrlFallback {
					res := testResponse(r, 302, "")
					location := "https://github.com/viciouscal/Vencord/releases/tag/devbuild"
					if scenario == "wrong-repo" {
						location = "https://github.com/Vendicated/Vencord/releases/tag/devbuild"
					}
					res.Header.Set("Location", location)
					return res, nil
				}
				return testResponse(r, 200, "<html>No assets</html>"), nil
			})
			if _, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback); err == nil {
				t.Fatal("expected failure")
			}
		})
	}
}

func TestLiveReleaseFallback(t *testing.T) {
	if os.Getenv("INSTALLER_LIVE_TEST") != "1" {
		t.Skip("opt-in network check")
	}
	for _, address := range []string{ReleaseUrlFallback, InstallerReleaseUrlFallback} {
		data, err := getReleasePage(address)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: %s, tag=%s, assets=%d", address, data.Name, data.TagName, len(data.Assets))
	}
}
