// File: /loam/internal/app/update.go

package app

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/iiif"
	"github.com/bmquinn/loam-iiif/internal/types"
	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	MinWidth  = 80
	MinHeight = 24
)

// Update is the main update loop for the Bubble Tea program.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// If the Chat panel is open, let the chat sub-update handle most inputs first.
	if m.ShowChat {
		newModel, subCmd := m.updateChat(msg)
		return newModel, subCmd
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
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
			modelsHeight = 0 // Fixed height for foundation models
			sectionGap   = 1
		)

		totalOverhead := titleHeight + textareaHeight + statusHeight + modelsHeight + helpHeight + (sectionGap * 4)
		listHeight := contentHeight - totalOverhead
		if listHeight < 1 {
			listHeight = 1
		}
		m.List.SetWidth(contentWidth - 2)
		m.List.SetHeight(listHeight - 2)

		// Update foundation models viewport size
		m.ModelViewport.Width = contentWidth - 2
		m.ModelViewport.Height = modelsHeight
		m.ModelViewport.SetContent("Loading models...") // Reset content on resize

		// Also update chat sub-model to match new window size
		m.Chat.Viewport.Width = contentWidth - 2
		m.Chat.TextArea.SetWidth(contentWidth - 2)

		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Toggle chat with "c"
		if key == "c" || key == "C" {
			m.ShowChat = !m.ShowChat
			if m.ShowChat {
				m.Status = "Opened chat panel."
				// Focus the chat text area
				m.Chat.TextArea.Focus()
			} else {
				m.Status = "Closed chat panel."
			}
			return m, nil
		}

		// If detail pane is open, check if user wants to close it with "esc"
		if m.ShowDetail {
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				// Close the detail pane
				m.ShowDetail = false
				m.Status = "Closed detail pane."
				return m, nil
			}
			// If the detail pane is open, ignore other keys
			return m, nil
		}

		// If we are NOT in the list, handle text input or switching
		if !m.InList {
			switch key {
			case "ctrl+c":
				return m, tea.Quit

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
					return m, tea.Batch(cmds...)
				}
			}
			var cmd tea.Cmd
			m.TextArea, cmd = m.TextArea.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// If we ARE in the list:
		switch key {
		case "ctrl+c":
			// Only Ctrl+C quits.
			return m, tea.Quit

		case "esc":
			// Instead of quitting, let's go "back" if possible.
			if len(m.PrevItemsStack) > 0 {
				// Pop the last item slice from the stack
				lastIndex := len(m.PrevItemsStack) - 1
				prevItems := m.PrevItemsStack[lastIndex]
				m.PrevItemsStack = m.PrevItemsStack[:lastIndex]

				// Restore the previous list
				m.List.SetItems(prevItems)
				m.Status = "Went back to previous list."
			} else {
				m.Status = "No previous items to go back to."
			}
			return m, nil

		case "tab":
			m.TextArea.Focus()
			m.InList = false
			m.Status = "Ready"
			return m, nil

		case "up", "down", "k", "j":
			if m.Status == "Opened in browser" {
				m.Status = "Ready"
			}

		case "enter":
			// Show detail or fetch nested collection
			if item, ok := m.List.SelectedItem().(ui.Item); ok {
				if strings.EqualFold(item.ItemType, "collection") {
					// Push the CURRENT list onto the stack
					currentItems := m.List.Items()
					m.PrevItemsStack = append(m.PrevItemsStack, currentItems)

					// Fetch the new collection
					m.Status = "Fetching nested collection..."
					m.Loading = true
					return m, tea.Batch(
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
			if item, ok := m.List.SelectedItem().(ui.Item); ok && item.URL != "Error" {
				if err := iiif.OpenURL(item.URL); err != nil {
					m.Status = "Failed to open URL"
				} else {
					m.Status = "Opened in browser"
				}
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case types.FetchDataMsg:
		// We got new data back from the IIIF API
		newItems := iiif.ParseData(msg)
		var listItems []list.Item
		for _, item := range newItems {
			listItems = append(listItems, item)
		}

		m.Mutex.Lock()
		m.List.SetItems(listItems)
		m.Mutex.Unlock()

		m.Status = fmt.Sprintf("Fetched %d items", len(newItems))
		m.Loading = false

		// Extract relevant context from newItems and store it
		// For example, concatenate titles or descriptions
		var contextBuilder strings.Builder
		for _, item := range newItems {
			contextBuilder.WriteString(fmt.Sprintf("Title: %s\nURL: %s\n\n", item.Title, item.URL))
		}
		m.Chat.Context = contextBuilder.String()

		return m, nil

	case types.ErrMsg:
		m.Status = "Error: " + msg.Error.Error()
		m.Loading = false
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		if m.Loading {
			cmds = append(cmds, m.Spinner.Tick)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case FoundationModelsMsg:
		// Handle the received foundation models
		if len(msg.Models) == 0 {
			m.ModelViewport.SetContent("No foundation models available.")
		} else {
			modelList := "Available Foundation Models:\n\n"
			for _, modelID := range msg.Models {
				modelList += "- " + modelID + "\n"
			}
			m.ModelViewport.SetContent(modelList)
		}
		m.Status = "Loaded foundation models."
		return m, nil

		// You can add more cases here if needed.

	}

	return m, tea.Batch(cmds...)
}

// Init sets up any initial commands for the Bubble Tea program.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.Spinner.Tick,
		GetModels(), // Fetch foundation models at startup
	)
}

// updateChat handles messages for the chat panel when it's open.
func (m *Model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd   tea.Cmd
		vpCmd   tea.Cmd
		chatCmd tea.Cmd
	)

	// Update text area and viewport
	m.Chat.TextArea, tiCmd = m.Chat.TextArea.Update(msg)
	m.Chat.Viewport, vpCmd = m.Chat.Viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Resize chat sub-model
		m.Chat.Viewport.Width = msg.Width - 4
		m.Chat.Viewport.Height = msg.Height / 3
		m.Chat.TextArea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			// Close chat if user presses Esc or Ctrl+C
			m.ShowChat = false
			m.Status = "Closed chat panel."
			return m, nil

		case tea.KeyEnter:
			// On Enter, send the message to Bedrock
			userInput := strings.TrimSpace(m.Chat.TextArea.Value())
			if userInput == "" {
				// Ignore empty messages
				return m, nil
			}

			// Append user's message
			userMessage := m.Chat.SenderStyle.Render("You: ") + userInput
			m.Chat.Messages = append(m.Chat.Messages, userMessage)

			// Update chat viewport
			m.Chat.Viewport.SetContent(strings.Join(m.Chat.Messages, "\n\n"))
			m.Chat.Viewport.GotoBottom()

			// Clear the text area
			m.Chat.TextArea.Reset()

			// Send the message to Bedrock with context
			chatCmd = SendChat(userInput, m.Chat.Context)
			return m, tea.Batch(tiCmd, vpCmd, chatCmd)
		}

	case ChatResponseMsg:
		// Append the assistant's response to messages
		assistantResponse := strings.TrimSpace(msg.Response)
		if assistantResponse != "" {
			assistantMessage := AssistantStyle.Render("Assistant: ") + assistantResponse
			m.Chat.Messages = append(m.Chat.Messages, assistantMessage)
			// log.Printf("Assistant message appended: %s", assistantResponse)

			// Update chat viewport
			m.Chat.Viewport.SetContent(strings.Join(m.Chat.Messages, "\n\n"))
			m.Chat.Viewport.GotoBottom()
		}
		return m, nil

	case ChatErrorMsg:
		// Append the error message to messages
		errorMessage := m.Chat.SenderStyle.Render("Error: ") + msg.Error.Error()
		m.Chat.Messages = append(m.Chat.Messages, errorMessage)

		// Update chat viewport
		m.Chat.Viewport.SetContent(strings.Join(m.Chat.Messages, "\n\n"))
		m.Chat.Viewport.GotoBottom()
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, chatCmd)
}
