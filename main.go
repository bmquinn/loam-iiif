package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants for minimum window size
const (
	minWidth  = 60
	minHeight = 20
)

// Item represents a single IIIF item
type Item struct {
	url   string
	title string
}

// ItemDelegate handles rendering of list items
type ItemDelegate struct {
	width  int
	styles struct {
		selectedTitle, selectedDesc, normalTitle, normalDesc lipgloss.Style
	}
}

// Messages
type (
	fetchDataMsg []byte
	errMsg       struct{ error }
)

// Model represents the application state
type Model struct {
	TextArea textarea.Model
	List     list.Model
	Status   string
	Spinner  spinner.Model
	Loading  bool
	mutex    sync.Mutex
	InList   bool
	width    int
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.url }
func (i Item) FilterValue() string { return i.title }

// truncateString truncates a string to a specified length with ellipsis
func truncateString(s string, length int) string {
	if utf8.RuneCountInString(s) <= length {
		return s
	}
	return string([]rune(s)[:length-3]) + "..."
}

func NewItemDelegate(width int) ItemDelegate {
	d := ItemDelegate{width: width}

	d.styles.selectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	d.styles.selectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	d.styles.normalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	d.styles.normalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return d
}

func (d ItemDelegate) Height() int                               { return 2 }
func (d ItemDelegate) Spacing() int                              { return 0 }
func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	// Truncate title and description to fit within the delegate width
	maxTitleLen := d.width - 5 // Leave some space for padding and styling
	maxDescLen := d.width - 5

	truncatedTitle := truncateString(i.Title(), maxTitleLen)
	truncatedDesc := truncateString(i.Description(), maxDescLen)

	var title, desc string
	if index == m.Index() {
		title = d.styles.selectedTitle.Render(truncatedTitle)
		desc = d.styles.selectedDesc.Render(truncatedDesc)
	} else {
		title = d.styles.normalTitle.Render(truncatedTitle)
		desc = d.styles.normalDesc.Render(truncatedDesc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

func initialModel() *Model {
	ta := textarea.New()
	ta.Placeholder = "Enter IIIF URL..."
	ta.Focus()
	ta.CharLimit = 256
	ta.SetWidth(40)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.KeyMap.DeleteWordBackward.SetEnabled(true)
	ta.KeyMap.DeleteWordForward.SetEnabled(true)
	ta.KeyMap.DeleteAfterCursor.SetEnabled(true)
	ta.KeyMap.DeleteBeforeCursor.SetEnabled(true)

	delegate := NewItemDelegate(40) // Initial width, will be updated
	l := list.New([]list.Item{}, delegate, 40, 10)
	l.Title = ""
	lipgloss.NewStyle().Height(0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.NoItems = lipgloss.NewStyle().Margin(1, 0)

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &Model{
		TextArea: ta,
		List:     l,
		Status:   "Ready",
		Spinner:  s,
		Loading:  false,
		InList:   false,
		width:    40,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.Spinner.Tick)
}

func fetchData(urlStr string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(urlStr)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errMsg{fmt.Errorf("failed to fetch data: %s", resp.Status)}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg{err}
		}
		return fetchDataMsg(body)
	}
}

func parseData(data []byte) []list.Item {
	type IIIFManifest struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Label struct {
			None []string `json:"none"`
		} `json:"label"`
		Summary struct {
			None []string `json:"none"`
		} `json:"summary"`
	}

	type IIIFManifestResponse struct {
		Items []IIIFManifest `json:"items"`
	}

	type IIIFCollectionManifest struct {
		ID    string `json:"@id"`
		Type  string `json:"@type"`
		Label string `json:"label"`
	}

	type IIIFCollectionResponse struct {
		Context   interface{}              `json:"@context"`
		ID        string                   `json:"@id"`
		Type      string                   `json:"@type"`
		Label     string                   `json:"label"`
		Manifests []IIIFCollectionManifest `json:"manifests"`
		Items     []IIIFManifest           `json:"items"`
	}

	var collectionResponse IIIFCollectionResponse
	if err := json.Unmarshal(data, &collectionResponse); err == nil {
		items := []list.Item{}
		if collectionResponse.Type == "sc:Collection" && len(collectionResponse.Manifests) > 0 {
			// Handle NLS-style collections
			for _, manifest := range collectionResponse.Manifests {
				items = append(items, Item{
					url:   manifest.ID,
					title: manifest.Label,
				})
			}
			return items
		} else if (strings.HasPrefix(collectionResponse.Type, "Collection") || collectionResponse.Type == "sc:Collection") && len(collectionResponse.Items) > 0 {
			// Handle Northwestern-style collections
			for _, item := range collectionResponse.Items {
				if item.Type == "Manifest" {
					var label string
					if len(item.Label.None) > 0 {
						label = item.Label.None[0]
					}
					items = append(items, Item{
						url:   item.ID,
						title: label,
					})
				}
			}
			return items
		}
	}

	var manifestResponse IIIFManifestResponse
	if err := json.Unmarshal(data, &manifestResponse); err == nil && len(manifestResponse.Items) > 0 {
		items := []list.Item{}
		for _, entry := range manifestResponse.Items {
			if entry.Type == "Collection" {
				continue
			}

			var label string
			if len(entry.Label.None) > 0 {
				label = entry.Label.None[0]
			}

			items = append(items, Item{
				url:   entry.ID,
				title: label,
			})
		}
		return items
	}

	return []list.Item{Item{url: "Error", title: "Failed to parse data as Manifest or Collection"}}
}

func openURL(urlStr string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{urlStr}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", urlStr}
	default:
		cmd = "xdg-open"
		args = []string{urlStr}
	}

	return exec.Command(cmd, args...).Start()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// First check minimum size
		if msg.Width < minWidth || msg.Height < minHeight {
			m.Status = fmt.Sprintf("Window too small. Minimum size is %dx%d", minWidth, minHeight)
			return m, nil
		}

		m.width = msg.Width - 6 // Account for main padding and borders

		// Calculate usable area (accounting for main padding)
		contentWidth := msg.Width - 4   // 2 cells padding on each side
		contentHeight := msg.Height - 2 // 1 cell padding top and bottom

		// Update delegate width
		m.List.SetDelegate(NewItemDelegate(contentWidth - 4)) // Account for list borders and padding

		// Set textarea width/height (accounting for its border)
		textareaWidth := contentWidth
		textareaHeight := 3                      // Fixed height for URL input
		m.TextArea.SetWidth(textareaWidth - 2)   // Account for border
		m.TextArea.SetHeight(textareaHeight - 2) // Account for border

		// Fixed heights for other elements
		const (
			titleHeight  = 1
			statusHeight = 3 // Title + bordered content
			helpHeight   = 1
			sectionGap   = 1
		)

		// Calculate total overhead and remaining height for list
		totalOverhead := titleHeight + textareaHeight + statusHeight + helpHeight + (sectionGap * 3)
		listHeight := contentHeight - totalOverhead
		if listHeight < 1 {
			listHeight = 1
		}
		m.List.SetWidth(contentWidth - 2) // Account for border
		m.List.SetHeight(listHeight - 2)  // Account for border and title

		return m, nil

	case tea.KeyMsg:
		if !m.InList {
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "tab":
				m.TextArea.Blur()
				m.InList = true
				m.Status = "Ready" // Reset status when switching to list
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
					cmds = append(cmds, fetchData(urlInput), m.Spinner.Tick)
					return m, tea.Batch(cmds...)
				}
			}
			var cmd tea.Cmd
			m.TextArea, cmd = m.TextArea.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.TextArea.Focus()
			m.InList = false
			m.Status = "Ready" // Reset status when switching to textarea
			return m, nil
		case "up", "down", "k", "j":
			// Reset status when moving in the list
			if m.Status == "Opened in browser" {
				m.Status = "Ready"
			}
		case "o", "O":
			if item, ok := m.List.SelectedItem().(Item); ok && item.url != "Error" {
				if err := openURL(item.url); err != nil {
					m.Status = "Failed to open URL"
				} else {
					m.Status = "Opened in browser"
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)

	case fetchDataMsg:
		items := parseData(msg)
		m.mutex.Lock()
		m.List.SetItems(items)
		m.mutex.Unlock()
		m.Status = fmt.Sprintf("Fetched %d items", len(items))
		m.Loading = false
		return m, nil

	case errMsg:
		m.Status = "Error: " + msg.error.Error()
		m.Loading = false
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		if m.Loading {
			cmds = append(cmds, m.Spinner.Tick)
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var sections []string

	// Input Section
	sections = append(sections,
		titleStyle.Render("LoamIIIF"),
		focusedBorderStyle.Render(m.TextArea.View()))

	// Status Section with Spinner
	statusContent := m.Status
	if m.Loading {
		statusContent = fmt.Sprintf("%s %s", m.Spinner.View(), m.Status)
	}
	sections = append(sections,
		titleStyle.Render("Status"),
		borderStyle.Render(statusContent))

	// Results Section
	sections = append(sections,
		titleStyle.Render("Results"),
		borderStyle.Render(m.List.View()))

	// Help Section
	sections = append(sections,
		helpStyle.Render("Tab: Switch Focus | Enter: Submit URL | O: Open URL | Ctrl+C/Esc: Quit"))

	// Join all sections with a gap between them
	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add padding around everything
	return lipgloss.NewStyle().Padding(1, 2).Render(mainContent)
}

func main() {
	p := tea.NewProgram(initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
