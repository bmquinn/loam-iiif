package iiif

import (
	"encoding/json"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/ui"
	"github.com/charmbracelet/bubbles/list"
)

func ParseData(data []byte) []list.Item {
	var collectionResponse IIIFCollectionResponse
	if err := json.Unmarshal(data, &collectionResponse); err == nil {
		items := []list.Item{}
		if collectionResponse.Type == "sc:Collection" && len(collectionResponse.Manifests) > 0 {
			for _, manifest := range collectionResponse.Manifests {
				items = append(items, ui.Item{
					URL:   manifest.ID,
					Title: manifest.Label,
				})
			}
			return items
		} else if (strings.HasPrefix(collectionResponse.Type, "Collection") || collectionResponse.Type == "sc:Collection") && len(collectionResponse.Items) > 0 {
			for _, item := range collectionResponse.Items {
				if item.Type == "Manifest" {
					var label string
					if len(item.Label.None) > 0 {
						label = item.Label.None[0]
					}
					items = append(items, ui.Item{
						URL:   item.ID,
						Title: label,
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
			items = append(items, ui.Item{
				URL:   entry.ID,
				Title: label,
			})
		}
		return items
	}

	return []list.Item{
		ui.Item{
			URL:   "Error",
			Title: "Failed to parse data as Manifest or Collection",
		},
	}
}
