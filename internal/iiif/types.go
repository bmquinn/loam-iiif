package iiif

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
