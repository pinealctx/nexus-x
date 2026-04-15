package adaptivecard

// Media displays a media player for audio or video content.
// Schema: https://adaptivecards.io/explorer/Media.html (version 1.1+)
type Media struct {
	ElementBase
	Type    string        `json:"type"`
	Sources []MediaSource `json:"sources"`
	Poster  string        `json:"poster,omitempty"`
	AltText string        `json:"altText,omitempty"`
}

func (*Media) elementType() string { return "Media" }

func NewMedia(sources ...MediaSource) *Media {
	return &Media{Type: "Media", Sources: sources}
}

func (m *Media) AddSource(s MediaSource) *Media { m.Sources = append(m.Sources, s); return m }
func (m *Media) SetPoster(url string) *Media    { m.Poster = url; return m }
func (m *Media) SetAltText(t string) *Media     { m.AltText = t; return m }
func (m *Media) SetID(id string) *Media         { m.ID = id; return m }

// MediaSource defines a source for a Media element.
type MediaSource struct {
	MimeType string `json:"mimeType"`
	URL      string `json:"url"`
}

func NewMediaSource(mimeType, url string) MediaSource {
	return MediaSource{MimeType: mimeType, URL: url}
}
