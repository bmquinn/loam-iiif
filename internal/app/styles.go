// File: /loam/internal/app/styles.go

package app

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	NoItemsStyle = lipgloss.NewStyle().Margin(1, 0)

	SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// New Style for Focused "LoamIIIF" Title
	FocusedTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("206")).
				Underline(true) // Optional: Adds underline to indicate focus

	// AssistantStyle for Assistant messages
	AssistantStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42")) // Choose a distinct color for Assistant
)
