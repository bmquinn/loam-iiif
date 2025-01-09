// File: /loam/internal/app/view.go

package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var sections []string

	// Title
	var title string
	if !m.InList {
		title = FocusedTitleStyle.Render("LoamIIIF")
	} else {
		title = TitleStyle.Render("LoamIIIF")
	}

	// Conditionally apply BorderStyle or FocusedBorderStyle to TextArea
	var textAreaView string
	if !m.InList {
		textAreaView = FocusedBorderStyle.Render(m.TextArea.View())
	} else {
		textAreaView = BorderStyle.Render(m.TextArea.View())
	}

	// Construct top sections
	sections = append(sections,
		title,
		textAreaView,
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

	// Add Foundation Models section using ModelViewport
	// foundationModelsSection := TitleStyle.Render("Foundation Models")
	// modelsContent := m.ModelViewport.View()
	// sections = append(sections,
	// 	foundationModelsSection,
	// 	BorderStyle.Render(modelsContent),
	// )

	// Main Section (Results or Detail)
	mainSection := m.renderMainSection()
	sections = append(sections, mainSection)

	// If the chat panel is open, render the chat at the bottom
	if m.ShowChat {
		chatSection := m.renderChatSection()
		sections = append(sections, chatSection)
	}

	// Footer help
	helpMsg := "Tab: Switch Focus | Enter: Open Detail | O: Open URL in browser | Esc: Close Detail/Back | c: Toggle Chat"
	sections = append(sections, HelpStyle.Render(helpMsg))

	// Join all sections vertically
	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().Padding(1, 2).Render(mainContent)
}

// renderMainSection handles either the detail view or the list view.
func (m *Model) renderMainSection() string {
	if m.ShowDetail {
		// Show selected record detail
		detailString := fmt.Sprintf(
			"Title: %s\nURL:   %s",
			m.SelectedItem.Title,
			m.SelectedItem.URL,
		)
		return lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("Record Detail"),
			BorderStyle.Render(detailString),
		)
	}

	// Otherwise, render the list
	var resultsView string
	if m.InList {
		resultsView = FocusedBorderStyle.Render(m.List.View())
	} else {
		resultsView = BorderStyle.Render(m.List.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Results"),
		resultsView,
	)
}

// renderChatSection shows the chat viewport and text area.
func (m *Model) renderChatSection() string {
	// Chat panel with a distinct border and title
	chatContent := lipgloss.JoinVertical(lipgloss.Left,
		m.Chat.Viewport.View(),
		m.Chat.TextArea.View(),
	)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		FocusedTitleStyle.Render("Chat Panel"),
		FocusedBorderStyle.Render(chatContent),
	)
}
