package ui

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ItemDelegate struct {
	Width  int
	Styles struct {
		SelectedTitle, SelectedDesc, NormalTitle, NormalDesc lipgloss.Style
	}
}

func NewItemDelegate(width int) ItemDelegate {
	d := ItemDelegate{Width: width}
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	return d
}

func (d ItemDelegate) Height() int {
	return 2
}

func (d ItemDelegate) Spacing() int {
	return 0
}

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	maxTitleLen := d.Width - 5
	maxDescLen := d.Width - 5

	truncatedTitle := truncateString(i.Title, maxTitleLen)
	truncatedDesc := truncateString(i.URL, maxDescLen)

	var title, desc string
	if index == m.Index() {
		title = d.Styles.SelectedTitle.Render(truncatedTitle)
		desc = d.Styles.SelectedDesc.Render(truncatedDesc)
	} else {
		title = d.Styles.NormalTitle.Render(truncatedTitle)
		desc = d.Styles.NormalDesc.Render(truncatedDesc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func truncateString(s string, length int) string {
	if utf8.RuneCountInString(s) <= length {
		return s
	}
	return string([]rune(s)[:length-3]) + "..."
}
