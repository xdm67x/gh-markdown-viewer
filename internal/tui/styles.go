package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base chrome
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	headerSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statusToastStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	statusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("9")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	// Diff lines
	hunkHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	sectionTagStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220")).
			Background(lipgloss.Color("236"))

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(4).
			Align(lipgloss.Right)

	gutterSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237"))

	// Added lines
	addedLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78"))

	addedWordStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("120")).
			Background(lipgloss.Color("22"))

	addedPrefixStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("78"))

	// Deleted lines
	deletedLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("167"))

	deletedWordStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("203")).
			Background(lipgloss.Color("52"))

	deletedPrefixStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("167"))

	// Context line
	contextLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// Badges
	commentBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("220")).
				Background(lipgloss.Color("236"))

	badgeAddStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("2")).
			SetString("[A]")

	badgeModStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("3")).
			SetString("[M]")

	badgeRenStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")).
			SetString("[R]")

	badgeDelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("1")).
			SetString("[D]")

	// Modal boxes
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Background(lipgloss.Color("235"))

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	modalFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244"))

	// Split view
	splitDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("239")).
				SetString("│")

	// Graph view
	graphBranchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	graphLevelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	graphTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255"))

	graphLinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213"))

	graphTaskStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	graphInspectorPanelStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("63")).
					Padding(1, 2).
					Background(lipgloss.Color("235"))
)
