//
// https://preview.iiif.io/api/content-state-comments/api/content-state/0.3/
//

package main

// ContentState ...
type ContentState struct {
	Context    string   `json:"@context"`
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Motivation []string `json:"motivation"`
	Target     Target   `json:"target"`
}

// Target ...
type Target struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	PartOf []Parts `json:"partOf"`
}

// Parts ...
type Parts struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
