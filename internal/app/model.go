package app

import (
	"sync"

	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
)

type Model struct {
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
}

func InitialModel() *Model {
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

	return &Model{
		TextArea:     ta,
		List:         l,
		Status:       "Ready",
		Spinner:      s,
		Loading:      false,
		InList:       false,
		Width:        40,
		ShowDetail:   false,
		SelectedItem: ui.Item{},
	}
}
