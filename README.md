# gh-markdown-viewer

A `gh` CLI extension that renders markdown files changed in a GitHub PR with a
terminal UI, and lets you post line-level review comments directly from your
terminal.

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

## Key bindings

### File picker

| Key   | Action          |
|-------|-----------------|
| j / ↓ | Move down       |
| k / ↑ | Move up         |
| enter | Open file       |
| q     | Quit            |

### Markdown viewer

| Key       | Action                             |
|-----------|------------------------------------|
| j / ↓     | Move cursor down one line          |
| k / ↑     | Move cursor up one line            |
| g         | Jump to top                        |
| G         | Jump to bottom                     |
| c         | Comment on current line (if in diff)|
| esc / ←   | Back to file picker                |
| q         | Quit                               |

### Comment input

| Key        | Action          |
|------------|-----------------|
| ctrl+s     | Submit comment  |
| esc        | Cancel          |

## How it works

Only `.md` files modified by the PR are listed. When you navigate to a line and
press `c`, the extension checks whether that source line falls inside the PR
diff. If it does, you can type a review comment that is posted to GitHub via the
REST API (`POST /repos/{owner}/{repo}/pulls/{n}/comments`).
