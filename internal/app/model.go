// File: /loam/internal/app/model.go

package app

import (
	"sync"

	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// ChatModel holds data for the chat feature.
type ChatModel struct {
	Viewport    viewport.Model
	Messages    []string
	TextArea    textarea.Model
	SenderStyle lipgloss.Style
	Err         error

	// New Field for Context
	Context string
}

// Model is the main application model.
type Model struct {
	// Existing fields
	TextArea     textarea.Model
	List         list.Model
	Status       string
	Spinner      spinner.Model
	Loading      bool
	Mutex        sync.Mutex
	InList       bool
	Width        int
	ShowDetail   bool
	SelectedItem ui.Item

	// Stack for item slices (so you can go back)
	PrevItemsStack [][]list.Item

	// --- New Chat Fields ---
	ShowChat        bool // Are we currently showing the chat panel?
	Chat            ChatModel
	AvailableModels []string       // List of foundation models
	ModelViewport   viewport.Model // Scrollable viewport for foundation models
	Err             error
}

// InitialChatModel creates an initialized ChatModel.
func InitialChatModel() ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(50) // Increased width for better usability
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle() // remove cursor line styling
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(50, 10) // Increased height for more messages
	vp.SetContent(`Welcome to LoamIIIF Chat!
Press 'esc' or 'c' to close the chat panel.`)

	return ChatModel{
		Viewport:    vp,
		Messages:    []string{},
		TextArea:    ta,
		SenderStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		Err:         nil,
		Context:     "", // Initialize context as empty
	}
}

// InitialModel initializes the main Model.
func InitialModel() *Model {
	// Existing initialization of the text input
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

	delegate := ui.NewItemDelegate(40)
	l := list.New([]list.Item{}, delegate, 40, 10)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle
	l.Styles.NoItems = NoItemsStyle

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = SpinnerStyle

	// Initialize foundation models viewport with default size; will be updated on WindowSizeMsg
	foundationModelsViewport := viewport.New(40, 10)
	foundationModelsViewport.SetContent("Loading models...")

	return &Model{
		TextArea:        ta,
		List:            l,
		Status:          "Ready",
		Spinner:         s,
		Loading:         false,
		InList:          false,
		Width:           40,
		ShowDetail:      false,
		SelectedItem:    ui.Item{},
		PrevItemsStack:  make([][]list.Item, 0),
		ShowChat:        false,
		Chat:            InitialChatModel(),
		AvailableModels: []string{},
		ModelViewport:   foundationModelsViewport,
		Err:             nil,
	}
}
