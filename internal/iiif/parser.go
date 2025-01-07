package iiif

import (
	"encoding/json"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/ui"
)

func ParseData(data []byte) []ui.Item {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return []ui.Item{
			{URL: "Error", Title: "JSON parse error", ItemType: "Error"},
		}
	}
	return parseIIIF(raw)
}

// parseIIIF handles both v3 ("type": "Collection"/"Manifest") and v2 ("@type": "sc:Collection"/"sc:Manifest").
func parseIIIF(val interface{}) []ui.Item {
	var out []ui.Item

	m, ok := val.(map[string]interface{})
	if !ok {
		return out
	}

	v3Type, _ := m["type"].(string)
	v2Type, _ := m["@type"].(string)

	switch {
	case strings.Contains(v2Type, "Collection") || v3Type == "Collection":
		label := fetchLabel(m)
		id := fetchID(m)
		out = append(out, ui.Item{URL: id, Title: label, ItemType: "Collection"})

		// Recurse for v3 "items" array
		if v3Type == "Collection" {
			if arr, ok := m["items"].([]interface{}); ok {
				for _, child := range arr {
					out = append(out, parseIIIF(child)...)
				}
			}
		}

		// Recurse for v2 "manifests" / "collections" arrays
		if manifests, ok := m["manifests"].([]interface{}); ok {
			for _, child := range manifests {
				out = append(out, parseIIIF(child)...)
			}
		}
		if collections, ok := m["collections"].([]interface{}); ok {
			for _, child := range collections {
				out = append(out, parseIIIF(child)...)
			}
		}

	case strings.Contains(v2Type, "Manifest") || v3Type == "Manifest":
		label := fetchLabel(m)
		id := fetchID(m)
		out = append(out, ui.Item{URL: id, Title: label, ItemType: "Manifest"})
	}

	return out
}

func fetchID(m map[string]interface{}) string {
	if id, ok := m["id"].(string); ok {
		return id
	}
	if id, ok := m["@id"].(string); ok {
		return id
	}
	return ""
}

func fetchLabel(m map[string]interface{}) string {
	// Try IIIF v3 style label
	if lab, ok := m["label"].(map[string]interface{}); ok {
		for _, key := range []string{"en", "none"} {
			if arr, ok := lab[key].([]interface{}); ok && len(arr) > 0 {
				if str, ok := arr[0].(string); ok {
					return str
				}
			}
		}
		// Possibly a single string "en"
		if s, ok := lab["en"].(string); ok {
			return s
		}
	}
	// Try IIIF v2 style
	if lbl, ok := m["label"].(string); ok {
		return lbl
	}
	return ""
}
