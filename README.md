# gh-markdown-viewer

A `gh` CLI extension that renders markdown files changed in a GitHub PR in your
browser and lets you post line-level review comments directly from the UI.

## Install

```sh
gh extension install xdm67x/gh-markdown-viewer
```

## Usage

```sh
# Auto-detect the PR for the current branch
gh markdown-viewer

# By PR number (current repo)
gh markdown-viewer 42

# By owner/repo#number
gh markdown-viewer owner/repo#42

# By URL
gh markdown-viewer https://github.com/owner/repo/pull/42

# Specify a different repo
gh markdown-viewer 42 --repo owner/repo
```

Running the command starts a local HTTP server on a random port, prints the URL,
and auto-opens it in your default browser. Press **Ctrl+C** to shut down.

## Web UI

The interface has two panels:

- **Sidebar** — lists all `.md` files changed by the PR with their status badge
  (`added`, `modified`, `renamed`) and `+additions / -deletions` counts.
- **Main panel** — two tabs per file:
  - **Rendered** — GitHub-flavored markdown rendered pixel-perfectly via the
    GitHub Markdown API (`POST /markdown`), including `#issue` and `@mention`
    links resolved relative to the PR's repository.
  - **Source** — raw markdown with numbered lines. Lines that are part of the PR
    diff are highlighted; clicking one opens an inline textarea for a review
    comment.

## Posting comments

1. Select a file in the sidebar, then switch to the **Source** tab.
2. Hover over a line — commentable lines (those in the diff) show a `+` marker.
3. Click the line or the `+` marker to open an inline comment form.
4. Type your comment and click **Submit**. The comment is posted via
   `POST /repos/{owner}/{repo}/pulls/{n}/comments` and a confirmation appears.
5. If you click a non-commentable line and try to submit, GitHub's own error
   message is surfaced (e.g. "line must be part of the diff").

## How it works

File content is fetched at the PR's head SHA via the GitHub Contents API.
Commentable lines are derived from the unified diff patch in the PR files
response: any `+` (added) or ` ` (context) line inside a diff hunk is
considered commentable. Deleted lines (`-`) are not present in the rendered
source view and are not offered for commenting.
