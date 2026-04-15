package adaptivecard

import "encoding/json"

// Container groups items together.
// Schema: https://adaptivecards.io/explorer/Container.html
type Container struct {
	ElementBase
	Type                     string                   `json:"type"`
	Items                    []Element                `json:"items,omitempty"`
	SelectAction             Action                   `json:"selectAction,omitempty"`
	Style                    ContainerStyle           `json:"style,omitempty"`
	VerticalContentAlignment VerticalContentAlignment `json:"verticalContentAlignment,omitempty"`
	Bleed                    *bool                    `json:"bleed,omitempty"`
	BackgroundImage          *BackgroundImage         `json:"backgroundImage,omitempty"`
	MinHeight                string                   `json:"minHeight,omitempty"`
	Rtl                      *bool                    `json:"rtl,omitempty"`
}

func (*Container) elementType() string { return "Container" }

func NewContainer(items ...Element) *Container {
	return &Container{Type: "Container", Items: items}
}

func (c *Container) AddItem(el Element) *Container        { c.Items = append(c.Items, el); return c }
func (c *Container) SetStyle(s ContainerStyle) *Container { c.Style = s; return c }
func (c *Container) SetBleed(v bool) *Container           { c.Bleed = &v; return c }
func (c *Container) SetMinHeight(h string) *Container     { c.MinHeight = h; return c }
func (c *Container) SetSelectAction(a Action) *Container  { c.SelectAction = a; return c }
func (c *Container) SetID(id string) *Container           { c.ID = id; return c }
func (c *Container) SetSpacing(s Spacing) *Container      { c.Spacing = s; return c }
func (c *Container) SetRtl(v bool) *Container             { c.Rtl = &v; return c }

func (c *Container) SetVerticalContentAlignment(v VerticalContentAlignment) *Container {
	c.VerticalContentAlignment = v
	return c
}

func (c *Container) MarshalJSON() ([]byte, error) {
	type alias Container
	cc := (*alias)(c)
	if cc.Type == "" {
		cc.Type = "Container"
	}
	return json.Marshal(cc)
}

func (c *Container) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "items", "selectAction")
	if err != nil {
		return err
	}
	type alias Container
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["items"]; ok {
		a.Items, err = unmarshalElements(raw)
		if err != nil {
			return err
		}
	}
	if raw, ok := fields["selectAction"]; ok {
		a.SelectAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*c = Container(a)
	return nil
}

// ColumnSet displays a set of columns.
// Schema: https://adaptivecards.io/explorer/ColumnSet.html
type ColumnSet struct {
	ElementBase
	Type                string              `json:"type"`
	Columns             []*Column           `json:"columns,omitempty"`
	SelectAction        Action              `json:"selectAction,omitempty"`
	Style               ContainerStyle      `json:"style,omitempty"`
	Bleed               *bool               `json:"bleed,omitempty"`
	MinHeight           string              `json:"minHeight,omitempty"`
	HorizontalAlignment HorizontalAlignment `json:"horizontalAlignment,omitempty"`
}

func (*ColumnSet) elementType() string { return "ColumnSet" }

func NewColumnSet(cols ...*Column) *ColumnSet {
	return &ColumnSet{Type: "ColumnSet", Columns: cols}
}

func (cs *ColumnSet) AddColumn(col *Column) *ColumnSet {
	cs.Columns = append(cs.Columns, col)
	return cs
}
func (cs *ColumnSet) SetStyle(s ContainerStyle) *ColumnSet { cs.Style = s; return cs }
func (cs *ColumnSet) SetBleed(v bool) *ColumnSet           { cs.Bleed = &v; return cs }
func (cs *ColumnSet) SetSelectAction(a Action) *ColumnSet  { cs.SelectAction = a; return cs }
func (cs *ColumnSet) SetID(id string) *ColumnSet           { cs.ID = id; return cs }
func (cs *ColumnSet) SetSpacing(s Spacing) *ColumnSet      { cs.Spacing = s; return cs }

func (cs *ColumnSet) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "columns", "selectAction")
	if err != nil {
		return err
	}
	type alias ColumnSet
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["columns"]; ok && len(raw) > 0 {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return err
		}
		a.Columns = make([]*Column, 0, len(arr))
		for _, r := range arr {
			var col Column
			if err := json.Unmarshal(r, &col); err != nil {
				return err
			}
			a.Columns = append(a.Columns, &col)
		}
	}
	if raw, ok := fields["selectAction"]; ok {
		a.SelectAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*cs = ColumnSet(a)
	return nil
}

// Column is a single column within a ColumnSet.
// Schema: https://adaptivecards.io/explorer/Column.html
type Column struct {
	ElementBase
	Type                     string                   `json:"type,omitempty"`
	Items                    []Element                `json:"items,omitempty"`
	BackgroundImage          *BackgroundImage         `json:"backgroundImage,omitempty"`
	Bleed                    *bool                    `json:"bleed,omitempty"`
	MinHeight                string                   `json:"minHeight,omitempty"`
	Rtl                      *bool                    `json:"rtl,omitempty"`
	SelectAction             Action                   `json:"selectAction,omitempty"`
	Style                    ContainerStyle           `json:"style,omitempty"`
	VerticalContentAlignment VerticalContentAlignment `json:"verticalContentAlignment,omitempty"`
	Width                    string                   `json:"width,omitempty"` // "auto", "stretch", "1", "2", "50px"
}

func NewColumn(items ...Element) *Column {
	return &Column{Type: "Column", Items: items}
}

func (c *Column) AddItem(el Element) *Column        { c.Items = append(c.Items, el); return c }
func (c *Column) SetWidth(w string) *Column         { c.Width = w; return c }
func (c *Column) SetStyle(s ContainerStyle) *Column { c.Style = s; return c }
func (c *Column) SetMinHeight(h string) *Column     { c.MinHeight = h; return c }
func (c *Column) SetSelectAction(a Action) *Column  { c.SelectAction = a; return c }
func (c *Column) SetID(id string) *Column           { c.ID = id; return c }
func (c *Column) SetSpacing(s Spacing) *Column      { c.Spacing = s; return c }

func (c *Column) SetVerticalContentAlignment(v VerticalContentAlignment) *Column {
	c.VerticalContentAlignment = v
	return c
}

func (c *Column) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "items", "selectAction")
	if err != nil {
		return err
	}
	type alias Column
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["items"]; ok {
		a.Items, err = unmarshalElements(raw)
		if err != nil {
			return err
		}
	}
	if raw, ok := fields["selectAction"]; ok {
		a.SelectAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*c = Column(a)
	return nil
}

// BackgroundImage specifies a background image for a container or card.
// Schema: https://adaptivecards.io/explorer/BackgroundImage.html
type BackgroundImage struct {
	URL                 string              `json:"url"`
	FillMode            ImageFillMode       `json:"fillMode,omitempty"`
	HorizontalAlignment HorizontalAlignment `json:"horizontalAlignment,omitempty"`
	VerticalAlignment   VerticalAlignment   `json:"verticalAlignment,omitempty"`
}
