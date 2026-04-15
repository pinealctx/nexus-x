package adaptivecard

import "encoding/json"

// Table provides tabular data display.
// Schema: https://adaptivecards.io/explorer/Table.html (version 1.5)
type Table struct {
	ElementBase
	Type                           string              `json:"type"`
	Columns                        []TableColumnDef    `json:"columns,omitempty"`
	Rows                           []TableRow          `json:"rows,omitempty"`
	FirstRowAsHeader               *bool               `json:"firstRowAsHeader,omitempty"`
	ShowGridLines                  *bool               `json:"showGridLines,omitempty"`
	GridStyle                      ContainerStyle      `json:"gridStyle,omitempty"`
	HorizontalCellContentAlignment HorizontalAlignment `json:"horizontalCellContentAlignment,omitempty"`
	VerticalCellContentAlignment   VerticalAlignment   `json:"verticalCellContentAlignment,omitempty"`
}

func (*Table) elementType() string { return "Table" }

func NewTable() *Table {
	h := true
	g := true
	return &Table{
		Type:             "Table",
		FirstRowAsHeader: &h,
		ShowGridLines:    &g,
	}
}

func (t *Table) AddColumn(width any) *Table {
	t.Columns = append(t.Columns, TableColumnDef{Type: "TableColumnDefinition", Width: width})
	return t
}

// AddTextRow adds a row of plain text cells.
func (t *Table) AddTextRow(texts ...string) *Table {
	cells := make([]TableCell, 0, len(texts))
	for _, text := range texts {
		cells = append(cells, TableCell{
			Type:  "TableCell",
			Items: []Element{NewTextBlock(text)},
		})
	}
	t.Rows = append(t.Rows, TableRow{Type: "TableRow", Cells: cells})
	return t
}

// AddRow adds a row with pre-built cells.
func (t *Table) AddRow(cells ...TableCell) *Table {
	t.Rows = append(t.Rows, TableRow{Type: "TableRow", Cells: cells})
	return t
}

func (t *Table) SetFirstRowAsHeader(v bool) *Table    { t.FirstRowAsHeader = &v; return t }
func (t *Table) SetShowGridLines(v bool) *Table       { t.ShowGridLines = &v; return t }
func (t *Table) SetGridStyle(s ContainerStyle) *Table { t.GridStyle = s; return t }
func (t *Table) SetID(id string) *Table               { t.ID = id; return t }
func (t *Table) SetSpacing(s Spacing) *Table          { t.Spacing = s; return t }

func (t *Table) SetHorizontalCellAlignment(a HorizontalAlignment) *Table {
	t.HorizontalCellContentAlignment = a
	return t
}

func (t *Table) SetVerticalCellAlignment(a VerticalAlignment) *Table {
	t.VerticalCellContentAlignment = a
	return t
}

// TableColumnDef defines a column in a Table.
type TableColumnDef struct {
	Type                           string              `json:"type,omitempty"`
	Width                          any                 `json:"width,omitempty"` // number or string like "1", "auto"
	HorizontalCellContentAlignment HorizontalAlignment `json:"horizontalCellContentAlignment,omitempty"`
	VerticalCellContentAlignment   VerticalAlignment   `json:"verticalCellContentAlignment,omitempty"`
}

// TableRow is a row in a Table.
type TableRow struct {
	Type                           string              `json:"type,omitempty"`
	Cells                          []TableCell         `json:"cells,omitempty"`
	Style                          ContainerStyle      `json:"style,omitempty"`
	HorizontalCellContentAlignment HorizontalAlignment `json:"horizontalCellContentAlignment,omitempty"`
	VerticalCellContentAlignment   VerticalAlignment   `json:"verticalCellContentAlignment,omitempty"`
}

// TableCell is a cell in a TableRow.
type TableCell struct {
	Type                     string                   `json:"type,omitempty"`
	Items                    []Element                `json:"items,omitempty"`
	SelectAction             Action                   `json:"selectAction,omitempty"`
	Style                    ContainerStyle           `json:"style,omitempty"`
	VerticalContentAlignment VerticalContentAlignment `json:"verticalContentAlignment,omitempty"`
	Bleed                    *bool                    `json:"bleed,omitempty"`
	MinHeight                string                   `json:"minHeight,omitempty"`
	Rtl                      *bool                    `json:"rtl,omitempty"`
}

// NewTableCell creates a cell with the given elements.
func NewTableCell(items ...Element) TableCell {
	return TableCell{Type: "TableCell", Items: items}
}

// TextCell is a convenience for creating a cell with a single TextBlock.
func TextCell(text string) TableCell {
	return NewTableCell(NewTextBlock(text))
}

func (c *TableCell) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "items", "selectAction")
	if err != nil {
		return err
	}
	type alias TableCell
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
	*c = TableCell(a)
	return nil
}
