package ghpr

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// PostLineComment posts a line-level review comment on the PR.
// line must be present in the diff (RIGHT for head lines, LEFT for base/deleted lines).
func PostLineComment(info PRInfo, path, body string, line int, side string) error {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return err
	}
	if side == "" {
		side = "RIGHT"
	}

	payload := map[string]any{
		"body":      body,
		"commit_id": info.HeadSHA,
		"path":      path,
		"line":      line,
		"side":      side,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var resp struct {
		ID int64 `json:"id"`
	}
	url := fmt.Sprintf("repos/%s/%s/pulls/%d/comments",
		info.Ref.Owner, info.Ref.Repo, info.Ref.Number)
	if err := client.Post(url, bytes.NewReader(raw), &resp); err != nil {
		// Surface GitHub's error message directly (e.g. "line must be part of the diff").
		return fmt.Errorf("posting comment: %w", err)
	}
	return nil
}
