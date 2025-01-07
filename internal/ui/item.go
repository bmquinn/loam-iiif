package ui

// ItemType can be "Manifest", "Collection", or something else if needed.
type Item struct {
	URL      string
	Title    string
	ItemType string
}

func (i Item) TitleText() string   { return i.Title }
func (i Item) Description() string { return i.URL }
func (i Item) FilterValue() string { return i.Title }
