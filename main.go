package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
	"github.com/xdm67x/gh-markdown-viewer/internal/tui"
)

func main() {
	repoFlag := flag.String("repo", "", "Repository in owner/repo format (overrides cwd detection)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gh markdown-viewer [<pr-ref>] [--repo owner/repo]\n\n")
		fmt.Fprintf(os.Stderr, "  <pr-ref>  PR number, owner/repo#N, or full PR URL\n")
		fmt.Fprintf(os.Stderr, "            (omit to auto-detect from current branch)\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	ref, err := resolveRef(flag.Arg(0), *repoFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	info, err := ghpr.FetchPRInfo(ref)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching PR: %s\n", err)
		os.Exit(1)
	}

	files, err := ghpr.ListMDFiles(info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing files: %s\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No markdown files changed in PR #%d.\n", ref.Number)
		os.Exit(0)
	}

	p := tea.NewProgram(
		tui.New(info, files),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func resolveRef(arg, repoFlag string) (ghpr.Ref, error) {
	var defaultOwner, defaultRepo string
	if repoFlag != "" {
		var err error
		defaultOwner, defaultRepo, err = ghpr.SplitOwnerRepo(repoFlag)
		if err != nil {
			return ghpr.Ref{}, err
		}
	}

	if arg == "" {
		return ghpr.CurrentBranchPR(repoFlag)
	}
	return ghpr.ParseRef(arg, defaultOwner, defaultRepo)
}
