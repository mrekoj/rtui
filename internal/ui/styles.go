package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorCyan    = lipgloss.Color("6")
	colorMagenta = lipgloss.Color("5")
	colorGreen   = lipgloss.Color("2")
	colorYellow  = lipgloss.Color("3")
	colorRed     = lipgloss.Color("1")
	colorGray    = lipgloss.Color("8")
	colorWhite   = lipgloss.Color("15")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Reverse(true)

	cleanRepoStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	dirtyRepoStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	stagedStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	modifiedStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	untrackedStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	conflictStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	aheadStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	behindStyle = lipgloss.NewStyle().
			Foreground(colorMagenta)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)
)
