package adaptivecard

import "encoding/json"

// Image displays an image.
// Schema: https://adaptivecards.io/explorer/Image.html
type Image struct {
	ElementBase
	Type                string              `json:"type"`
	URL                 string              `json:"url"`
	AltText             string              `json:"altText,omitempty"`
	BackgroundColor     string              `json:"backgroundColor,omitempty"`
	Height              BlockElementHeight  `json:"height,omitempty"`
	HorizontalAlignment HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	SelectAction        Action              `json:"selectAction,omitempty"`
	Size                ImageSize           `json:"size,omitempty"`
	Style               ImageStyle          `json:"style,omitempty"`
	Width               string              `json:"width,omitempty"`
}

func (*Image) elementType() string { return "Image" }

func NewImage(url string) *Image {
	return &Image{Type: "Image", URL: url}
}

func (i *Image) SetAltText(t string) *Image             { i.AltText = t; return i }
func (i *Image) SetSize(s ImageSize) *Image             { i.Size = s; return i }
func (i *Image) SetImageStyle(s ImageStyle) *Image      { i.Style = s; return i }
func (i *Image) SetWidth(w string) *Image               { i.Width = w; return i }
func (i *Image) SetHAlign(a HorizontalAlignment) *Image { i.HorizontalAlignment = a; return i }
func (i *Image) SetSelectAction(a Action) *Image        { i.SelectAction = a; return i }
func (i *Image) SetBackgroundColor(c string) *Image     { i.BackgroundColor = c; return i }
func (i *Image) SetID(id string) *Image                 { i.ID = id; return i }

func (i *Image) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "selectAction")
	if err != nil {
		return err
	}
	type alias Image
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["selectAction"]; ok {
		a.SelectAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*i = Image(a)
	return nil
}
