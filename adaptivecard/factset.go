package adaptivecard

import "encoding/json"

// FactSet displays a series of key/value pairs.
// Schema: https://adaptivecards.io/explorer/FactSet.html
type FactSet struct {
	ElementBase
	Type  string `json:"type"`
	Facts []Fact `json:"facts"`
}

func (*FactSet) elementType() string { return "FactSet" }

func NewFactSet(facts ...Fact) *FactSet {
	return &FactSet{Type: "FactSet", Facts: facts}
}

func (f *FactSet) AddFact(title, value string) *FactSet {
	f.Facts = append(f.Facts, Fact{Title: title, Value: value})
	return f
}

func (f *FactSet) SetID(id string) *FactSet      { f.ID = id; return f }
func (f *FactSet) SetSpacing(s Spacing) *FactSet { f.Spacing = s; return f }

// Fact is a single key/value pair within a FactSet.
type Fact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// ImageSet displays a collection of images.
// Schema: https://adaptivecards.io/explorer/ImageSet.html
type ImageSet struct {
	ElementBase
	Type      string    `json:"type"`
	Images    []*Image  `json:"images"`
	ImageSize ImageSize `json:"imageSize,omitempty"`
}

func (*ImageSet) elementType() string { return "ImageSet" }

func NewImageSet(images ...*Image) *ImageSet {
	return &ImageSet{Type: "ImageSet", Images: images}
}

func (s *ImageSet) AddImage(img *Image) *ImageSet       { s.Images = append(s.Images, img); return s }
func (s *ImageSet) SetImageSize(sz ImageSize) *ImageSet { s.ImageSize = sz; return s }
func (s *ImageSet) SetID(id string) *ImageSet           { s.ID = id; return s }

// ActionSet displays a set of actions inline within the card body.
// Schema: https://adaptivecards.io/explorer/ActionSet.html
type ActionSet struct {
	ElementBase
	Type    string   `json:"type"`
	Actions []Action `json:"actions"`
}

func (*ActionSet) elementType() string { return "ActionSet" }

func NewActionSet(actions ...Action) *ActionSet {
	return &ActionSet{Type: "ActionSet", Actions: actions}
}

func (a *ActionSet) AddAction(act Action) *ActionSet { a.Actions = append(a.Actions, act); return a }
func (a *ActionSet) SetID(id string) *ActionSet      { a.ID = id; return a }

func (a *ActionSet) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "actions")
	if err != nil {
		return err
	}
	type alias ActionSet
	var al alias
	if err := json.Unmarshal(rest, &al); err != nil {
		return err
	}
	if raw, ok := fields["actions"]; ok {
		al.Actions, err = unmarshalActions(raw)
		if err != nil {
			return err
		}
	}
	*a = ActionSet(al)
	return nil
}
