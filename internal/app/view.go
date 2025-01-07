package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var sections []string

	// Title and TextArea
	sections = append(sections,
		TitleStyle.Render("LoamIIIF"),
		FocusedBorderStyle.Render(m.TextArea.View()),
	)

	// Status
	statusContent := m.Status
	if m.Loading {
		statusContent = fmt.Sprintf("%s %s", m.Spinner.View(), m.Status)
	}
	sections = append(sections,
		TitleStyle.Render("Status"),
		BorderStyle.Render(statusContent),
	)

	// If showing detail, render detail pane; otherwise, render list
	if m.ShowDetail {
		// Show selected record detail
		detailString := fmt.Sprintf(
			"Title: %s\nURL:   %s",
			m.SelectedItem.Title,
			m.SelectedItem.URL,
		)

		sections = append(sections,
			TitleStyle.Render("Record Detail"),
			BorderStyle.Render(detailString),
		)
	} else {
		// Conditionally apply focused style to Results
		var resultsView string
		if m.InList {
			// Apply FocusedBorderStyle when "Results" is focused
			resultsView = FocusedBorderStyle.Render(m.List.View())
		} else {
			// Use default BorderStyle otherwise
			resultsView = BorderStyle.Render(m.List.View())
		}

		sections = append(sections,
			TitleStyle.Render("Results"),
			resultsView,
		)
	}

	// Footer help
	sections = append(sections,
		HelpStyle.Render("Tab: Switch Focus | Enter: Open Detail | O: Open URL in browser | Esc: Quit/Close Detail"),
	)

	// Join all sections vertically with left alignment
	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().Padding(1, 2).Render(mainContent)
}
