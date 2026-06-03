// Package ghpr handles GitHub PR resolution and API calls.
package ghpr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	goghr "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// Ref uniquely identifies a pull request.
type Ref struct {
	Owner  string
	Repo   string
	Number int
}

var (
	reOwnerRepoNum = regexp.MustCompile(`^([^/]+)/([^#]+)#(\d+)$`)
	rePRURL        = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
)

// ParseRef parses a pr-ref argument which may be:
//   - a bare number ("42") — owner/repo detected from cwd or defaultOwner/defaultRepo
//   - "owner/repo#42"
//   - a full PR URL
func ParseRef(arg, defaultOwner, defaultRepo string) (Ref, error) {
	if m := rePRURL.FindStringSubmatch(arg); m != nil {
		n, _ := strconv.Atoi(m[3])
		return Ref{Owner: m[1], Repo: m[2], Number: n}, nil
	}
	if m := reOwnerRepoNum.FindStringSubmatch(arg); m != nil {
		n, _ := strconv.Atoi(m[3])
		return Ref{Owner: m[1], Repo: m[2], Number: n}, nil
	}
	if n, err := strconv.Atoi(arg); err == nil {
		owner, repo, err := resolveRepo(defaultOwner, defaultRepo)
		if err != nil {
			return Ref{}, err
		}
		return Ref{Owner: owner, Repo: repo, Number: n}, nil
	}
	return Ref{}, fmt.Errorf("cannot parse PR reference %q", arg)
}

// CurrentBranchPR resolves the open PR for the current branch using gh pr view.
// ownerRepo may be "owner/repo" or empty to let gh detect from cwd.
func CurrentBranchPR(ownerRepo string) (Ref, error) {
	args := []string{"pr", "view", "--json", "number,headRepository"}
	if ownerRepo != "" {
		args = append(args, "--repo", ownerRepo)
	}
	out, _, err := goghr.Exec(args...)
	if err != nil {
		return Ref{}, fmt.Errorf("could not detect current PR: %w", err)
	}

	var v struct {
		Number         int `json:"number"`
		HeadRepository struct {
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			Name string `json:"name"`
		} `json:"headRepository"`
	}
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		return Ref{}, fmt.Errorf("parsing pr view output: %w", err)
	}
	return Ref{
		Owner:  v.HeadRepository.Owner.Login,
		Repo:   v.HeadRepository.Name,
		Number: v.Number,
	}, nil
}

// resolveRepo returns (owner, repo) from explicit values or cwd detection.
func resolveRepo(owner, repo string) (string, string, error) {
	if owner != "" && repo != "" {
		return owner, repo, nil
	}
	r, err := repository.Current()
	if err != nil {
		return "", "", fmt.Errorf("could not detect repository from cwd: %w", err)
	}
	return r.Owner, r.Name, nil
}

// SplitOwnerRepo splits "owner/repo" into its two parts.
func SplitOwnerRepo(ownerRepo string) (owner, repo string, err error) {
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected owner/repo, got %q", ownerRepo)
	}
	return parts[0], parts[1], nil
}
