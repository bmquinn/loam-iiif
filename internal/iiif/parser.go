package iiif

import (
	"encoding/json"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/list"
)

// Label represents the various ways labels can be structured in IIIF
type Label struct {
	None []string          `json:"none,omitempty"`
	En   []string          `json:"en,omitempty"`
	Raw  string            `json:"@value,omitempty"`
	Map  map[string]string `json:"-"`
}

// UnmarshalJSON handles the various label formats
func (l *Label) UnmarshalJSON(data []byte) error {
	// Try structured format first
	type labelAlias Label
	if err := json.Unmarshal(data, (*labelAlias)(l)); err == nil {
		return nil
	}

	// Try string format
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		l.None = []string{str}
		return nil
	}

	// Try map format
	if err := json.Unmarshal(data, &l.Map); err == nil {
		return nil
	}

	return nil // Return nil to skip invalid labels
}

// GetText returns the best available label text
func (l Label) GetText() string {
	if len(l.None) > 0 {
		return l.None[0]
	}
	if len(l.En) > 0 {
		return l.En[0]
	}
	if l.Raw != "" {
		return l.Raw
	}
	if len(l.Map) > 0 {
		// Try common language keys
		for _, lang := range []string{"en", "none", "@value"} {
			if val, ok := l.Map[lang]; ok {
				return val
			}
		}
		// Take first available value
		for _, v := range l.Map {
			return v
		}
	}
	return "Untitled"
}

type IIIFResource struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Label Label  `json:"label"`
}

type IIIFCollection struct {
	IIIFResource
	Items     []IIIFResource    `json:"items,omitempty"`
	Manifests []IIIFResource    `json:"manifests,omitempty"`
	Context   json.RawMessage   `json:"@context,omitempty"`
	Members   []json.RawMessage `json:"members,omitempty"`
	Sequences []json.RawMessage `json:"sequences,omitempty"`
}

// ParseData reads the raw JSON and returns a list of items (Collection + Manifest).
func ParseData(data []byte) []list.Item {
	var collection IIIFCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return []list.Item{
			ui.Item{
				URL:   "Error",
				Title: "Failed to parse IIIF data: " + err.Error(),
			},
		}
	}

	items := []list.Item{}

	// If it's a Collection-type resource
	if isCollectionType(collection.Type) {

		// Process items array
		for _, item := range collection.Items {
			switch {
			case isManifestType(item.Type):
				// It's a manifest
				items = append(items, ui.Item{
					URL:      item.ID,
					Title:    item.Label.GetText(),
					ItemType: "Manifest",
				})

			case isCollectionType(item.Type):
				// It's a child collection
				items = append(items, ui.Item{
					URL:      item.ID,
					Title:    item.Label.GetText(),
					ItemType: "Collection",
				})
			}
		}

		// Also handle older IIIF v2 "manifests" array if present
		for _, manifest := range collection.Manifests {
			items = append(items, ui.Item{
				URL:      manifest.ID,
				Title:    manifest.Label.GetText(),
				ItemType: "Manifest",
			})
		}

	} else if isManifestType(collection.Type) {
		// If it's a single Manifest
		items = append(items, ui.Item{
			URL:      collection.ID,
			Title:    collection.Label.GetText(),
			ItemType: "Manifest",
		})
	}

	// If we didn't find anything
	if len(items) == 0 {
		return []list.Item{
			ui.Item{
				URL:   "Error",
				Title: "No valid manifests or child collections found",
			},
		}
	}

	return items
}

// We consider type "Collection" if it matches these known patterns
func isCollectionType(t string) bool {
	t = strings.ToLower(t)
	return t == "collection" || t == "sc:collection" || strings.HasPrefix(t, "collection")
}

// We consider type "Manifest" if it matches these known patterns
func isManifestType(t string) bool {
	t = strings.ToLower(t)
	return t == "manifest" || t == "sc:manifest" || strings.HasPrefix(t, "manifest")
}
