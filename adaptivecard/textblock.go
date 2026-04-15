package adaptivecard

// TextBlock displays text, allowing control over font sizes, weight, and color.
// Schema: https://adaptivecards.io/explorer/TextBlock.html
type TextBlock struct {
	ElementBase
	Type                string              `json:"type"`
	Text                string              `json:"text"`
	Color               TextColor           `json:"color,omitempty"`
	FontType            FontType            `json:"fontType,omitempty"`
	HorizontalAlignment HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	IsSubtle            *bool               `json:"isSubtle,omitempty"`
	MaxLines            *int                `json:"maxLines,omitempty"`
	Size                FontSize            `json:"size,omitempty"`
	Weight              FontWeight          `json:"weight,omitempty"`
	Wrap                *bool               `json:"wrap,omitempty"`
	Style               TextBlockStyle      `json:"style,omitempty"`
}

func (*TextBlock) elementType() string { return "TextBlock" }

// NewTextBlock creates a TextBlock with the given text.
func NewTextBlock(text string) *TextBlock {
	return &TextBlock{Type: "TextBlock", Text: text}
}

func (t *TextBlock) SetColor(c TextColor) *TextBlock            { t.Color = c; return t }
func (t *TextBlock) SetFontType(f FontType) *TextBlock          { t.FontType = f; return t }
func (t *TextBlock) SetHAlign(a HorizontalAlignment) *TextBlock { t.HorizontalAlignment = a; return t }
func (t *TextBlock) SetSubtle(v bool) *TextBlock                { t.IsSubtle = &v; return t }
func (t *TextBlock) SetMaxLines(n int) *TextBlock               { t.MaxLines = &n; return t }
func (t *TextBlock) SetSize(s FontSize) *TextBlock              { t.Size = s; return t }
func (t *TextBlock) SetWeight(w FontWeight) *TextBlock          { t.Weight = w; return t }
func (t *TextBlock) SetWrap(w bool) *TextBlock                  { t.Wrap = &w; return t }
func (t *TextBlock) SetStyle(s TextBlockStyle) *TextBlock       { t.Style = s; return t }
func (t *TextBlock) SetID(id string) *TextBlock                 { t.ID = id; return t }
func (t *TextBlock) SetSpacing(s Spacing) *TextBlock            { t.Spacing = s; return t }
func (t *TextBlock) SetSeparator(v bool) *TextBlock             { t.Separator = v; return t }
func (t *TextBlock) SetHeight(h BlockElementHeight) *TextBlock  { t.Height = h; return t }
func (t *TextBlock) SetVisible(v bool) *TextBlock               { t.IsVisible = &v; return t }
