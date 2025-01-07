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

// IIIFResource is used for either v2 or v3 resources (Manifests, Collections, etc.)
// We implement a custom unmarshaller so we can handle both "id" and "@id" (v3 vs v2).
type IIIFResource struct {
	ID    string `json:"-"` // We’ll fill this from either "id" or "@id"
	Type  string `json:"-"` // We’ll fill this from either "type" or "@type"
	Label Label  `json:"label,omitempty"`
}

// resourceAlias is an internal struct for decoding either v2 or v3 styles.
type resourceAlias struct {
	ID     string `json:"id,omitempty"`
	AtID   string `json:"@id,omitempty"`
	Type   string `json:"type,omitempty"`
	AtType string `json:"@type,omitempty"`
	Label  Label  `json:"label,omitempty"`
}

// UnmarshalJSON tries to decode both v3 ("id", "type") and v2 ("@id", "@type").
func (r *IIIFResource) UnmarshalJSON(data []byte) error {
	var tmp resourceAlias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// ID could come from "id" (v3) or "@id" (v2)
	if tmp.ID != "" {
		r.ID = tmp.ID
	} else {
		r.ID = tmp.AtID
	}

	// Type could come from "type" (v3) or "@type" (v2)
	if tmp.Type != "" {
		r.Type = tmp.Type
	} else {
		r.Type = tmp.AtType
	}

	r.Label = tmp.Label
	return nil
}

// IIIFCollection can hold both v2 and v3 fields: "items" (v3) or "manifests" (v2), etc.
type IIIFCollection struct {
	IIIFResource
	Items     []IIIFResource    `json:"items,omitempty"`     // v3
	Manifests []IIIFResource    `json:"manifests,omitempty"` // v2
	Context   json.RawMessage   `json:"@context,omitempty"`
	Members   []json.RawMessage `json:"members,omitempty"`
	Sequences []json.RawMessage `json:"sequences,omitempty"`
}

// ParseData reads the raw JSON and returns a list of items (Collections + Manifests).
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

		// Process v3 "items"
		for _, item := range collection.Items {
			switch {
			case isManifestType(item.Type):
				items = append(items, ui.Item{
					URL:      item.ID,
					Title:    item.Label.GetText(),
					ItemType: "Manifest",
				})
			case isCollectionType(item.Type):
				items = append(items, ui.Item{
					URL:      item.ID,
					Title:    item.Label.GetText(),
					ItemType: "Collection",
				})
			}
		}

		// Also handle v2 "manifests"
		for _, manifest := range collection.Manifests {
			items = append(items, ui.Item{
				URL:      manifest.ID,
				Title:    manifest.Label.GetText(),
				ItemType: "Manifest",
			})
		}

	} else if isManifestType(collection.Type) {
		// If top-level object is a single Manifest
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
