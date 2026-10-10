# gh-markdown-viewer

A `gh` CLI extension that makes reading and reviewing Markdown diffs effortless for developers directly in your terminal.

With AI tools and agents writing and rewriting plans, specifications, harnesses, and documentation, reviewing raw Git diffs can be noisy and hard to parse. `gh-markdown-viewer` renders Markdown diffs with human-friendly styling, word-level change highlights, and semantic intent summaries — and lets you post line-level review comments, reply to discussions, and resolve threads without leaving the terminal.

## Features

- **Styled Unified Diff (`1`)**: Markdown-aware unified diff with intra-line / word-level diff highlights, distinct heading colors, checklist indicators, and review comment badges (`💬`).
- **Side-by-Side Split Diff (`2`)**: Synchronized Base (before) and Head (after) columns side-by-side with intra-line word highlights.
- **Rendered Document View (`3`)**: Full post-change Markdown document rendered in your terminal via Glamour.
- **Section Graph & Hierarchy Navigator (`4` or `s`)**: Interactive graph/tree representation of document sections (`#`, `##`, `###`), showing hierarchy, diff additions/deletions per section, task list progress (`[X/Y ✓]`), review comments (`💬`), and internal anchor cross-links (`↗`). Press **Enter** on any section node to jump directly to it in the diff view!
- **AI Intent & Changes Summary (`i`)**: Instant overview of what the changes actually mean — outlines modified sections with addition/deletion stats and tracks checklist transitions (`[ ]` → `[x]`).
- **Review Comments & Conversations**:
  - Post line-level review comments on any line in the diff (`c`).
  - View existing comment threads directly from diff lines (`t` / `Enter`).
  - Reply to discussion threads (`r`).
  - Toggle thread resolved status (`e`).
- **Changed Files Switcher (`f` or `b`)**: Switch between `.md` files changed in the PR with addition/deletion counts and comment indicators.

## Install

```sh
gh extension install xdm67x/gh-markdown-viewer
```

Or install locally from source:

```sh
gh extension install .
```

## Usage

```sh
# Auto-detect the PR for the current branch
gh markdown-viewer

# By PR number (in current repo)
gh markdown-viewer 42

# By owner/repo#number
gh markdown-viewer owner/repo#42

# By full PR URL
gh markdown-viewer https://github.com/owner/repo/pull/42

# Specify a different repo
gh markdown-viewer 42 --repo owner/repo
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move cursor down one line |
| `k` / `↑` | Move cursor up one line |
| `d` / `ctrl+d` | Scroll half page down |
| `u` / `ctrl+u` | Scroll half page up |
| `g` / `G` | Jump to top / bottom |
| `n` / `]` | Jump to next diff hunk |
| `p` / `[` | Jump to previous diff hunk |

### Views & Overlays

| Key | Action |
|---|---|
| `1` | Switch to Styled Unified Diff |
| `2` | Switch to Side-by-Side Split Diff |
| `3` | Switch to Full Rendered Document |
| `4` / `s` | Switch to Interactive Section Graph & Outline Navigator |
| `Tab` | Cycle through view modes |
| `Enter` | (In Graph) Jump directly to selected section in Diff view |
| `i` | Open AI Intent & Changes Summary |
| `f` / `b` | Open File Switcher |
| `?` | Toggle Help overlay |
| `q` / `ctrl+c` | Quit |

### Reviews & Comments

| Key | Action |
|---|---|
| `c` | Post review comment on selected diff line |
| `t` / `Enter` | View comment thread on selected line |
| `r` | Reply to thread (inside thread viewer) |
| `e` | Toggle Resolve / Unresolve thread |
| `ctrl+s` | Submit comment / reply in editor |
| `esc` | Close modal / cancel comment |

## How It Works

- Fetches PR metadata, changed Markdown files, and unified patches using the GitHub REST and GraphQL APIs.
- Analyzes diff hunks to compute word-level changes between modified lines, extracts section headings, and tracks task list checkbox transitions.
- Line comments are posted directly to GitHub via `POST /repos/{owner}/{repo}/pulls/{number}/comments` with appropriate side (`RIGHT` for additions/context, `LEFT` for deletions).
- Review threads, replies, and resolutions synchronize directly with GitHub's Pull Request Review Threads API.
