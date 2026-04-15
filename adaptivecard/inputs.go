package adaptivecard

import "encoding/json"

// InputText lets a user enter text.
// Schema: https://adaptivecards.io/explorer/Input.Text.html
type InputText struct {
	InputBase
	Type         string         `json:"type"`
	IsMultiline  *bool          `json:"isMultiline,omitempty"`
	MaxLength    *int           `json:"maxLength,omitempty"`
	Placeholder  string         `json:"placeholder,omitempty"`
	Regex        string         `json:"regex,omitempty"`
	Style        TextInputStyle `json:"style,omitempty"`
	InlineAction Action         `json:"inlineAction,omitempty"`
	Value        string         `json:"value,omitempty"`
}

func (*InputText) elementType() string { return "Input.Text" }

func NewInputText(id string) *InputText {
	return &InputText{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.Text",
	}
}

func (i *InputText) SetLabel(l string) *InputText         { i.Label = l; return i }
func (i *InputText) SetPlaceholder(p string) *InputText   { i.Placeholder = p; return i }
func (i *InputText) SetStyle(s TextInputStyle) *InputText { i.Style = s; return i }
func (i *InputText) SetMultiline(v bool) *InputText       { i.IsMultiline = &v; return i }
func (i *InputText) SetMaxLength(n int) *InputText        { i.MaxLength = &n; return i }
func (i *InputText) SetRegex(r string) *InputText         { i.Regex = r; return i }
func (i *InputText) SetRequired(v bool) *InputText        { i.IsRequired = &v; return i }
func (i *InputText) SetValue(v string) *InputText         { i.Value = v; return i }
func (i *InputText) SetErrorMessage(m string) *InputText  { i.ErrorMessage = m; return i }
func (i *InputText) SetInlineAction(a Action) *InputText  { i.InlineAction = a; return i }
func (i *InputText) SetSpacing(s Spacing) *InputText      { i.Spacing = s; return i }

func (i *InputText) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "inlineAction")
	if err != nil {
		return err
	}
	type alias InputText
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["inlineAction"]; ok {
		a.InlineAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*i = InputText(a)
	return nil
}

// InputNumber lets a user enter a number.
// Schema: https://adaptivecards.io/explorer/Input.Number.html
type InputNumber struct {
	InputBase
	Type        string   `json:"type"`
	Max         *float64 `json:"max,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Value       *float64 `json:"value,omitempty"`
}

func (*InputNumber) elementType() string { return "Input.Number" }

func NewInputNumber(id string) *InputNumber {
	return &InputNumber{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.Number",
	}
}

func (i *InputNumber) SetLabel(l string) *InputNumber        { i.Label = l; return i }
func (i *InputNumber) SetPlaceholder(p string) *InputNumber  { i.Placeholder = p; return i }
func (i *InputNumber) SetMin(v float64) *InputNumber         { i.Min = &v; return i }
func (i *InputNumber) SetMax(v float64) *InputNumber         { i.Max = &v; return i }
func (i *InputNumber) SetValue(v float64) *InputNumber       { i.Value = &v; return i }
func (i *InputNumber) SetRequired(v bool) *InputNumber       { i.IsRequired = &v; return i }
func (i *InputNumber) SetErrorMessage(m string) *InputNumber { i.ErrorMessage = m; return i }

// InputDate lets a user choose a date.
// Schema: https://adaptivecards.io/explorer/Input.Date.html
type InputDate struct {
	InputBase
	Type        string `json:"type"`
	Max         string `json:"max,omitempty"`
	Min         string `json:"min,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Value       string `json:"value,omitempty"`
}

func (*InputDate) elementType() string { return "Input.Date" }

func NewInputDate(id string) *InputDate {
	return &InputDate{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.Date",
	}
}

func (i *InputDate) SetLabel(l string) *InputDate       { i.Label = l; return i }
func (i *InputDate) SetPlaceholder(p string) *InputDate { i.Placeholder = p; return i }
func (i *InputDate) SetMin(v string) *InputDate         { i.Min = v; return i }
func (i *InputDate) SetMax(v string) *InputDate         { i.Max = v; return i }
func (i *InputDate) SetValue(v string) *InputDate       { i.Value = v; return i }
func (i *InputDate) SetRequired(v bool) *InputDate      { i.IsRequired = &v; return i }

// InputTime lets a user select a time.
// Schema: https://adaptivecards.io/explorer/Input.Time.html
type InputTime struct {
	InputBase
	Type        string `json:"type"`
	Max         string `json:"max,omitempty"`
	Min         string `json:"min,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Value       string `json:"value,omitempty"`
}

func (*InputTime) elementType() string { return "Input.Time" }

func NewInputTime(id string) *InputTime {
	return &InputTime{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.Time",
	}
}

func (i *InputTime) SetLabel(l string) *InputTime       { i.Label = l; return i }
func (i *InputTime) SetPlaceholder(p string) *InputTime { i.Placeholder = p; return i }
func (i *InputTime) SetMin(v string) *InputTime         { i.Min = v; return i }
func (i *InputTime) SetMax(v string) *InputTime         { i.Max = v; return i }
func (i *InputTime) SetValue(v string) *InputTime       { i.Value = v; return i }
func (i *InputTime) SetRequired(v bool) *InputTime      { i.IsRequired = &v; return i }

// InputToggle lets a user choose between two options (on/off).
// Schema: https://adaptivecards.io/explorer/Input.Toggle.html
type InputToggle struct {
	InputBase
	Type     string `json:"type"`
	Title    string `json:"title"`
	Value    string `json:"value,omitempty"`
	ValueOff string `json:"valueOff,omitempty"`
	ValueOn  string `json:"valueOn,omitempty"`
	Wrap     *bool  `json:"wrap,omitempty"`
}

func (*InputToggle) elementType() string { return "Input.Toggle" }

func NewInputToggle(id, title string) *InputToggle {
	return &InputToggle{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.Toggle",
		Title:     title,
	}
}

func (i *InputToggle) SetLabel(l string) *InputToggle    { i.Label = l; return i }
func (i *InputToggle) SetValue(v string) *InputToggle    { i.Value = v; return i }
func (i *InputToggle) SetValueOn(v string) *InputToggle  { i.ValueOn = v; return i }
func (i *InputToggle) SetValueOff(v string) *InputToggle { i.ValueOff = v; return i }
func (i *InputToggle) SetWrap(v bool) *InputToggle       { i.Wrap = &v; return i }
func (i *InputToggle) SetRequired(v bool) *InputToggle   { i.IsRequired = &v; return i }

// InputChoiceSet lets a user choose from a set of options.
// Schema: https://adaptivecards.io/explorer/Input.ChoiceSet.html
type InputChoiceSet struct {
	InputBase
	Type          string           `json:"type"`
	Choices       []Choice         `json:"choices,omitempty"`
	IsMultiSelect *bool            `json:"isMultiSelect,omitempty"`
	Style         ChoiceInputStyle `json:"style,omitempty"`
	Value         string           `json:"value,omitempty"`
	Placeholder   string           `json:"placeholder,omitempty"`
	Wrap          *bool            `json:"wrap,omitempty"`
}

func (*InputChoiceSet) elementType() string { return "Input.ChoiceSet" }

func NewInputChoiceSet(id string) *InputChoiceSet {
	return &InputChoiceSet{
		InputBase: InputBase{ElementBase: ElementBase{ID: id}},
		Type:      "Input.ChoiceSet",
	}
}

func (i *InputChoiceSet) AddChoice(title, value string) *InputChoiceSet {
	i.Choices = append(i.Choices, Choice{Title: title, Value: value})
	return i
}

func (i *InputChoiceSet) SetLabel(l string) *InputChoiceSet           { i.Label = l; return i }
func (i *InputChoiceSet) SetMultiSelect(v bool) *InputChoiceSet       { i.IsMultiSelect = &v; return i }
func (i *InputChoiceSet) SetStyle(s ChoiceInputStyle) *InputChoiceSet { i.Style = s; return i }
func (i *InputChoiceSet) SetValue(v string) *InputChoiceSet           { i.Value = v; return i }
func (i *InputChoiceSet) SetPlaceholder(p string) *InputChoiceSet     { i.Placeholder = p; return i }
func (i *InputChoiceSet) SetWrap(v bool) *InputChoiceSet              { i.Wrap = &v; return i }
func (i *InputChoiceSet) SetRequired(v bool) *InputChoiceSet          { i.IsRequired = &v; return i }

// Choice is a single option within an Input.ChoiceSet.
type Choice struct {
	Title string `json:"title"`
	Value string `json:"value"`
}
