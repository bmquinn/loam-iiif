package ui

type Item struct {
	URL   string
	Title string
}

func (i Item) TitleText() string   { return i.Title }
func (i Item) Description() string { return i.URL }
func (i Item) FilterValue() string { return i.Title }
