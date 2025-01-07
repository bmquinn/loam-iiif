package app

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/iiif"
	"github.com/bmquinn/loam-iiif/internal/types"
	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	bubbletea "github.com/charmbracelet/bubbletea"
)

const (
	MinWidth  = 60
	MinHeight = 20
)

// Init sets up any initial commands for the Bubble Tea program.
func (m *Model) Init() bubbletea.Cmd {
	return bubbletea.Batch(
		textarea.Blink,
		m.Spinner.Tick,
	)
}

// Update is the main update loop for the Bubble Tea program.
func (m *Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var cmds []bubbletea.Cmd

	switch msg := msg.(type) {

	case bubbletea.WindowSizeMsg:
		if msg.Width < MinWidth || msg.Height < MinHeight {
			m.Status = fmt.Sprintf("Window too small. Minimum size is %dx%d", MinWidth, MinHeight)
			return m, nil
		}
		m.Width = msg.Width - 6

		contentWidth := msg.Width - 4
		contentHeight := msg.Height - 2

		m.List.SetDelegate(ui.NewItemDelegate(contentWidth - 4))

		textareaWidth := contentWidth
		textareaHeight := 3
		m.TextArea.SetWidth(textareaWidth - 2)
		m.TextArea.SetHeight(textareaHeight - 2)

		const (
			titleHeight  = 1
			statusHeight = 3
			helpHeight   = 1
			sectionGap   = 1
		)

		totalOverhead := titleHeight + textareaHeight + statusHeight + helpHeight + (sectionGap * 3)
		listHeight := contentHeight - totalOverhead
		if listHeight < 1 {
			listHeight = 1
		}
		m.List.SetWidth(contentWidth - 2)
		m.List.SetHeight(listHeight - 2)
		return m, nil

	case bubbletea.KeyMsg:
		key := msg.String()

		// If detail pane is open, check if user wants to close it with "esc"
		if m.ShowDetail {
			switch key {
			case "ctrl+c":
				return m, bubbletea.Quit
			case "esc":
				// Close the detail pane
				m.ShowDetail = false
				m.Status = "Closed detail pane."
				return m, nil
			}
			// If the detail pane is open, we ignore other keys
			return m, nil
		}

		// If we are NOT in the list, handle text input or switching
		if !m.InList {
			switch key {
			case "ctrl+c":
				// (NEW) We remove "esc" from here so it doesn't quit.
				return m, bubbletea.Quit

			// case "esc": do nothing here, or implement "go back" if you want
			// to return from the input form to something else. Usually not needed.

			case "tab":
				m.TextArea.Blur()
				m.InList = true
				m.Status = "Ready"
				if len(m.List.Items()) > 0 {
					m.List.Select(0)
				}
				return m, nil

			case "enter":
				urlInput := m.TextArea.Value()
				if urlInput != "" {
					parsedURL, err := url.ParseRequestURI(urlInput)
					if err != nil {
						m.Status = "Invalid URL format"
						return m, nil
					}
					if parsedURL.Scheme == "" {
						m.Status = "URL must include http:// or https://"
						return m, nil
					}
					m.Status = "Fetching data..."
					m.Loading = true
					cmds = append(cmds, iiif.FetchData(urlInput), m.Spinner.Tick)
					return m, bubbletea.Batch(cmds...)
				}
			}
			var cmd bubbletea.Cmd
			m.TextArea, cmd = m.TextArea.Update(msg)
			cmds = append(cmds, cmd)
			return m, bubbletea.Batch(cmds...)
		}

		// If we ARE in the list:
		switch key {
		case "ctrl+c":
			// Only Ctrl+C quits. (Removed "esc" here!)
			return m, bubbletea.Quit

		case "esc":
			// (NEW) Instead of quitting, let's go "back" if possible.
			if len(m.PrevItemsStack) > 0 {
				// Pop the last item slice from the stack
				lastIndex := len(m.PrevItemsStack) - 1
				prevItems := m.PrevItemsStack[lastIndex]
				m.PrevItemsStack = m.PrevItemsStack[:lastIndex]

				// Restore the previous list
				m.List.SetItems(prevItems)
				m.Status = "Went back to previous list."
			} else {
				// No previous list: do nothing or set a status
				m.Status = "No previous items to go back to."
			}
			return m, nil

		case "tab":
			m.TextArea.Focus()
			m.InList = false
			m.Status = "Ready"
			return m, nil

		case "up", "down", "k", "j":
			// If the status was "Opened in browser", reset it
			if m.Status == "Opened in browser" {
				m.Status = "Ready"
			}

		case "enter":
			// Show detail or fetch nested collection
			if item, ok := m.List.SelectedItem().(ui.Item); ok {
				if strings.EqualFold(item.ItemType, "collection") {
					// (NEW) Before we fetch nested items, push the CURRENT list onto the stack
					currentItems := m.List.Items()
					m.PrevItemsStack = append(m.PrevItemsStack, currentItems)

					// Now fetch the new collection
					m.Status = "Fetching nested collection..."
					m.Loading = true
					return m, bubbletea.Batch(
						iiif.FetchData(item.URL),
						m.Spinner.Tick,
					)

				} else {
					// It's a manifest (or something else)
					m.SelectedItem = item
					m.ShowDetail = true
					m.Status = fmt.Sprintf("Viewing detail: %s", item.Title)
				}
			}
			return m, nil

		case "o", "O":
			// 'Open in browser'
			if item, ok := m.List.SelectedItem().(ui.Item); ok && item.URL != "Error" {
				if err := iiif.OpenURL(item.URL); err != nil {
					m.Status = "Failed to open URL"
				} else {
					m.Status = "Opened in browser"
				}
				return m, nil
			}
		}

		var cmd bubbletea.Cmd
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)

	case types.FetchDataMsg:
		// We got new data back from the IIIF API
		newItems := iiif.ParseData(msg)

		m.Mutex.Lock()
		m.List.SetItems(newItems)
		m.Mutex.Unlock()

		m.Status = fmt.Sprintf("Fetched %d items", len(newItems))
		m.Loading = false
		return m, nil

	case types.ErrMsg:
		m.Status = "Error: " + msg.Error.Error()
		m.Loading = false
		return m, nil

	case spinner.TickMsg:
		var cmd bubbletea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		if m.Loading {
			cmds = append(cmds, m.Spinner.Tick)
		}
		cmds = append(cmds, cmd)
		return m, bubbletea.Batch(cmds...)
	}

	return m, bubbletea.Batch(cmds...)
}
