package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var sections []string

	sections = append(sections,
		TitleStyle.Render("LoamIIIF"),
		FocusedBorderStyle.Render(m.TextArea.View()),
	)

	statusContent := m.Status
	if m.Loading {
		statusContent = fmt.Sprintf("%s %s", m.Spinner.View(), m.Status)
	}
	sections = append(sections,
		TitleStyle.Render("Status"),
		BorderStyle.Render(statusContent),
	)

	sections = append(sections,
		TitleStyle.Render("Results"),
		BorderStyle.Render(m.List.View()),
	)

	sections = append(sections,
		HelpStyle.Render("Tab: Switch Focus | Enter: Submit URL | O: Open URL | Ctrl+C/Esc: Quit"),
	)

	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().Padding(1, 2).Render(mainContent)
}
