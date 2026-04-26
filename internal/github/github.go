// Package github provides a thin client for the github/gitignore repository,
// exposing the data the CLI needs to list and fetch .gitignore templates.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	contentsURL     = "https://api.github.com/repos/github/gitignore/contents"
	rawTemplateURL  = "https://raw.githubusercontent.com/github/gitignore/main"
	gitignoreSuffix = ".gitignore"
	userAgent       = "gitignore-cli"
)

// ErrTemplateNotFound is returned by FetchTemplate when the requested
// template name does not exist in github/gitignore. Callers can compare
// with errors.Is to provide a helpful "did you mean…" message.
var ErrTemplateNotFound = errors.New("template not found")

// contentEntry mirrors the subset of fields the GitHub contents API returns
// that this package actually uses.
type contentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// ListTemplates fetches the names of all root-level .gitignore templates
// from the github/gitignore repository. The ".gitignore" suffix is stripped
// from each returned name (e.g. "Python", not "Python.gitignore"). Entries
// that are directories (Global, community, .github) are excluded.
func ListTemplates(ctx context.Context) ([]string, error) {
	entries, err := fetchContents(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Type != "file" || !strings.HasSuffix(e.Name, gitignoreSuffix) {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name, gitignoreSuffix))
	}
	return names, nil
}

// FetchTemplate downloads the raw contents of the named .gitignore template
// from github/gitignore. The name is matched case-sensitively and must not
// include the ".gitignore" suffix (e.g. "Python", "C++"). When the template
// does not exist the returned error wraps ErrTemplateNotFound.
func FetchTemplate(ctx context.Context, name string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s%s", rawTemplateURL, name, gitignoreSuffix)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %q", ErrTemplateNotFound, name)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("github raw returned %s for %s", resp.Status, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}

// fetchContents calls the github/gitignore contents API and decodes the
// response. Non-2xx responses are surfaced as errors.
func fetchContents(ctx context.Context) ([]contentEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, contentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", contentsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("github contents API returned %s", resp.Status)
	}

	var entries []contentEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return entries, nil
}
