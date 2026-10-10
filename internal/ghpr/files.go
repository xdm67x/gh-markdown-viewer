package ghpr

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// File represents a markdown file changed in a PR.
type File struct {
	Filename string
	Patch    string // unified diff patch from the PR files endpoint
	Status   string // added, modified, renamed, etc.
	SHA      string // blob SHA at head
	// Additions / Deletions for display in the file picker
	Additions int
	Deletions int
}

// PRInfo holds the top-level data fetched once per PR.
type PRInfo struct {
	Ref     Ref
	HeadSHA string
	BaseSHA string
	Title   string
}

// FetchPRInfo fetches the PR metadata (head SHA, base SHA, title).
func FetchPRInfo(ref Ref) (PRInfo, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return PRInfo{}, err
	}
	var pr struct {
		Title string `json:"title"`
		Head  struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
	}
	url := fmt.Sprintf("repos/%s/%s/pulls/%d", ref.Owner, ref.Repo, ref.Number)
	if err := client.Get(url, &pr); err != nil {
		return PRInfo{}, fmt.Errorf("fetching PR: %w", err)
	}
	return PRInfo{
		Ref:     ref,
		HeadSHA: pr.Head.SHA,
		BaseSHA: pr.Base.SHA,
		Title:   pr.Title,
	}, nil
}

// ListMDFiles returns all non-removed markdown files changed in the PR.
func ListMDFiles(info PRInfo) ([]File, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}

	var all []File
	page := 1
	for {
		var batch []struct {
			Filename  string `json:"filename"`
			Patch     string `json:"patch"`
			Status    string `json:"status"`
			SHA       string `json:"sha"`
			Additions int    `json:"additions"`
			Deletions int    `json:"deletions"`
		}
		url := fmt.Sprintf(
			"repos/%s/%s/pulls/%d/files?per_page=100&page=%d",
			info.Ref.Owner, info.Ref.Repo, info.Ref.Number, page,
		)
		if err := client.Get(url, &batch); err != nil {
			return nil, fmt.Errorf("listing PR files (page %d): %w", page, err)
		}
		for _, f := range batch {
			if f.Status == "removed" {
				continue
			}
			if !strings.HasSuffix(f.Filename, ".md") {
				continue
			}
			all = append(all, File{
				Filename:  f.Filename,
				Patch:     f.Patch,
				Status:    f.Status,
				SHA:       f.SHA,
				Additions: f.Additions,
				Deletions: f.Deletions,
			})
		}
		if len(batch) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// FileContent fetches the raw content of path at the PR head SHA.
func FileContent(info PRInfo, path string) (string, error) {
	return FileContentAt(info, path, info.HeadSHA)
}

// FileBaseContent fetches the raw content of path at the PR base SHA.
// If the file did not exist in base (e.g. newly added), it returns empty string with no error.
func FileBaseContent(info PRInfo, path string) (string, error) {
	if info.BaseSHA == "" {
		return "", nil
	}
	content, err := FileContentAt(info, path, info.BaseSHA)
	if err != nil {
		// Treat 404 (file not found in base) as empty file
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return "", nil
		}
		return "", err
	}
	return content, nil
}

// FileContentAt fetches the raw content of path at a specific commit or ref.
func FileContentAt(info PRInfo, path string, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}
	var resp struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	url := fmt.Sprintf(
		"repos/%s/%s/contents/%s?ref=%s",
		info.Ref.Owner, info.Ref.Repo, path, ref,
	)
	if err := client.Get(url, &resp); err != nil {
		return "", fmt.Errorf("fetching content of %s at %s: %w", path, ref, err)
	}
	if resp.Encoding != "base64" {
		return resp.Content, nil
	}
	// GitHub returns base64 with line-break wrapping; strip them.
	clean := strings.ReplaceAll(resp.Content, "\n", "")
	raw, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return "", fmt.Errorf("decoding content of %s: %w", path, err)
	}
	return string(raw), nil
}
