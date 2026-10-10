package ghpr

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ReviewComment is a single comment in a PR review thread.
type ReviewComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
	URL       string `json:"url"`
}

// ReviewThread groups comments anchored to a file path and source line.
type ReviewThread struct {
	NodeID   string          `json:"nodeId"`
	Path     string          `json:"path"`
	Line     int             `json:"line"`
	Resolved bool            `json:"resolved"`
	Comments []ReviewComment `json:"comments"`
}

// FetchThreads returns all review threads for the PR via GraphQL.
func FetchThreads(info PRInfo) ([]ReviewThread, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}

	query := `
query($owner:String!, $repo:String!, $number:Int!) {
  repository(owner:$owner, name:$repo) {
    pullRequest(number:$number) {
      reviewThreads(first:100) {
        nodes {
          id
          isResolved
          path
          line
          originalLine
          comments(first:50) {
            nodes {
              databaseId
              body
              author { login }
              createdAt
              url
            }
          }
        }
      }
    }
  }
}`

	vars := map[string]any{
		"owner":  info.Ref.Owner,
		"repo":   info.Ref.Repo,
		"number": info.Ref.Number,
	}

	var resp struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						ID           string `json:"id"`
						IsResolved   bool   `json:"isResolved"`
						Path         string `json:"path"`
						Line         *int   `json:"line"`
						OriginalLine *int   `json:"originalLine"`
						Comments     struct {
							Nodes []struct {
								DatabaseID int64  `json:"databaseId"`
								Body       string `json:"body"`
								Author     struct {
									Login string `json:"login"`
								} `json:"author"`
								CreatedAt string `json:"createdAt"`
								URL       string `json:"url"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := client.Do(query, vars, &resp); err != nil {
		return nil, fmt.Errorf("fetching review threads: %w", err)
	}

	var threads []ReviewThread
	for _, n := range resp.Repository.PullRequest.ReviewThreads.Nodes {
		line := 0
		if n.Line != nil {
			line = *n.Line
		} else if n.OriginalLine != nil {
			// outdated threads use originalLine
			line = *n.OriginalLine
		}

		var comments []ReviewComment
		for _, c := range n.Comments.Nodes {
			comments = append(comments, ReviewComment{
				ID:        c.DatabaseID,
				Body:      c.Body,
				Author:    c.Author.Login,
				CreatedAt: c.CreatedAt,
				URL:       c.URL,
			})
		}
		threads = append(threads, ReviewThread{
			NodeID:   n.ID,
			Path:     n.Path,
			Line:     line,
			Resolved: n.IsResolved,
			Comments: comments,
		})
	}
	return threads, nil
}

// ResolveThread marks a review thread as resolved via GraphQL.
func ResolveThread(nodeID string) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return err
	}
	mutation := `
mutation($threadId:ID!) {
  resolveReviewThread(input:{threadId:$threadId}) {
    thread { isResolved }
  }
}`
	var result map[string]any
	return client.Do(mutation, map[string]any{"threadId": nodeID}, &result)
}

// UnresolveThread marks a review thread as un-resolved via GraphQL.
func UnresolveThread(nodeID string) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return err
	}
	mutation := `
mutation($threadId:ID!) {
  unresolveReviewThread(input:{threadId:$threadId}) {
    thread { isResolved }
  }
}`
	var result map[string]any
	return client.Do(mutation, map[string]any{"threadId": nodeID}, &result)
}

// ReplyToComment posts a reply to an existing PR review comment.
func ReplyToComment(info PRInfo, inReplyTo int64, body string) error {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(map[string]any{
		"body":         body,
		"in_reply_to":  inReplyTo,
	})
	if err != nil {
		return err
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	url := fmt.Sprintf("repos/%s/%s/pulls/%d/comments",
		info.Ref.Owner, info.Ref.Repo, info.Ref.Number)
	err = client.Post(url, bytes.NewReader(raw), &resp)
	if err != nil {
		return fmt.Errorf("posting reply: %w", err)
	}
	return nil
}
