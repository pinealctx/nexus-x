package adaptivecard

import "encoding/json"

// RichTextBlock displays rich text with inline formatting.
// Schema: https://adaptivecards.io/explorer/RichTextBlock.html
type RichTextBlock struct {
	ElementBase
	Type                string              `json:"type"`
	Inlines             []Inline            `json:"inlines"`
	HorizontalAlignment HorizontalAlignment `json:"horizontalAlignment,omitempty"`
}

func (*RichTextBlock) elementType() string { return "RichTextBlock" }

func NewRichTextBlock(inlines ...Inline) *RichTextBlock {
	return &RichTextBlock{Type: "RichTextBlock", Inlines: inlines}
}

func (r *RichTextBlock) AddInline(i Inline) *RichTextBlock {
	r.Inlines = append(r.Inlines, i)
	return r
}

func (r *RichTextBlock) SetHAlign(a HorizontalAlignment) *RichTextBlock {
	r.HorizontalAlignment = a
	return r
}

func (r *RichTextBlock) SetID(id string) *RichTextBlock { r.ID = id; return r }

func (r *RichTextBlock) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "inlines")
	if err != nil {
		return err
	}
	type alias RichTextBlock
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["inlines"]; ok {
		a.Inlines, err = unmarshalSlice(raw, unmarshalInline)
		if err != nil {
			return err
		}
	}
	*r = RichTextBlock(a)
	return nil
}

// unmarshalInline decodes a single Inline element (currently only TextRun).
func unmarshalInline(raw json.RawMessage) (Inline, error) {
	var tr TextRun
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// Inline is the interface for inline elements within a RichTextBlock.
// Currently only TextRun is defined in the schema.
type Inline interface {
	inlineType() string
}

// TextRun is an inline text element with formatting.
// Schema: https://adaptivecards.io/explorer/TextRun.html
type TextRun struct {
	Type          string     `json:"type"`
	Text          string     `json:"text"`
	Color         TextColor  `json:"color,omitempty"`
	FontType      FontType   `json:"fontType,omitempty"`
	Highlight     *bool      `json:"highlight,omitempty"`
	IsSubtle      *bool      `json:"isSubtle,omitempty"`
	Italic        *bool      `json:"italic,omitempty"`
	SelectAction  Action     `json:"selectAction,omitempty"`
	Size          FontSize   `json:"size,omitempty"`
	Strikethrough *bool      `json:"strikethrough,omitempty"`
	Underline     *bool      `json:"underline,omitempty"`
	Weight        FontWeight `json:"weight,omitempty"`
}

func (*TextRun) inlineType() string { return "TextRun" }

func NewTextRun(text string) *TextRun {
	return &TextRun{Type: "TextRun", Text: text}
}

func (t *TextRun) SetColor(c TextColor) *TextRun     { t.Color = c; return t }
func (t *TextRun) SetSize(s FontSize) *TextRun       { t.Size = s; return t }
func (t *TextRun) SetWeight(w FontWeight) *TextRun   { t.Weight = w; return t }
func (t *TextRun) SetItalic(v bool) *TextRun         { t.Italic = &v; return t }
func (t *TextRun) SetBold() *TextRun                 { t.Weight = WeightBolder; return t }
func (t *TextRun) SetStrikethrough(v bool) *TextRun  { t.Strikethrough = &v; return t }
func (t *TextRun) SetUnderline(v bool) *TextRun      { t.Underline = &v; return t }
func (t *TextRun) SetHighlight(v bool) *TextRun      { t.Highlight = &v; return t }
func (t *TextRun) SetSubtle(v bool) *TextRun         { t.IsSubtle = &v; return t }
func (t *TextRun) SetSelectAction(a Action) *TextRun { t.SelectAction = a; return t }

func (t *TextRun) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "selectAction")
	if err != nil {
		return err
	}
	type alias TextRun
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
	*t = TextRun(a)
	return nil
}
