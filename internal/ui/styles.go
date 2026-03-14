package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Title style for section headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			PaddingLeft(1)

	// Success indicator.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	// Error indicator.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Warning indicator.
	WarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	// Dim text for secondary info.
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	// Key-value label.
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Width(20)

	// Key-value value.
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Status icons.
	IconOK   = SuccessStyle.Render("✓")
	IconFail = ErrorStyle.Render("✗")
	IconWarn = WarnStyle.Render("!")
	IconInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render("ℹ")
)
