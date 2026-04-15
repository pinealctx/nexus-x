package adaptivecard_test

import (
	"encoding/json"
	"testing"

	ac "github.com/pinealctx/nexus-x/adaptivecard"
)

func TestCardJSON(t *testing.T) {
	card := ac.NewCard().
		AddBody(ac.NewTextBlock("Hello World").SetWeight(ac.WeightBolder).SetSize(ac.SizeLarge)).
		AddBody(ac.NewTextBlock("This is a test card.").SetWrap(true)).
		AddAction(ac.NewActionSubmit("Submit", map[string]any{"verb": "test"}))

	data, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify key fields.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if raw["type"] != "AdaptiveCard" {
		t.Errorf("type = %v, want AdaptiveCard", raw["type"])
	}
	if raw["version"] != "1.5" {
		t.Errorf("version = %v, want 1.5", raw["version"])
	}
	body, ok := raw["body"].([]any)
	if !ok || len(body) != 2 {
		t.Fatalf("body length = %d, want 2", len(body))
	}
	actions, ok := raw["actions"].([]any)
	if !ok || len(actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(actions))
	}
}

func TestCardRoundTrip(t *testing.T) {
	original := ac.NewCard().
		AddBody(ac.NewTextBlock("Title").SetWeight(ac.WeightBolder)).
		AddBody(ac.NewFactSet(ac.Fact{Title: "Key", Value: "Value"})).
		AddBody(ac.NewInputText("name").SetLabel("Name").SetPlaceholder("Enter name")).
		AddAction(ac.NewActionSubmit("OK", map[string]any{"verb": "submit"})).
		AddAction(ac.NewActionOpenURL("Docs", "https://example.com"))

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ac.Card
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Type != "AdaptiveCard" {
		t.Errorf("type = %q", decoded.Type)
	}
	if decoded.Version != "1.5" {
		t.Errorf("version = %q", decoded.Version)
	}
	if len(decoded.Body) != 3 {
		t.Errorf("body len = %d, want 3", len(decoded.Body))
	}
	if len(decoded.Actions) != 2 {
		t.Errorf("actions len = %d, want 2", len(decoded.Actions))
	}
}

func TestConnectFormCard(t *testing.T) {
	// Simulate the OpenProject connect form card.
	card := ac.NewCard().
		AddBody(ac.NewTextBlock("Connect to OpenProject").SetWeight(ac.WeightBolder).SetSize(ac.SizeMedium)).
		AddBody(ac.NewTextBlock("Enter your OpenProject instance URL and API key.")).
		AddBody(ac.NewInputText("url").SetLabel("OpenProject URL").SetPlaceholder("https://your-instance.openproject.com")).
		AddBody(ac.NewInputText("apiKey").SetLabel("API Key").SetPlaceholder("Your API key").SetStyle(ac.InputStylePassword)).
		AddAction(ac.NewActionSubmit("Connect", map[string]any{"verb": "bind"}))

	s, err := card.JSON()
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(s) == 0 {
		t.Fatal("empty json")
	}

	// Verify it round-trips.
	var decoded ac.Card
	if err := json.Unmarshal([]byte(s), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Body) != 4 {
		t.Errorf("body len = %d, want 4", len(decoded.Body))
	}
}

func TestTableCard(t *testing.T) {
	card := ac.NewCard().
		AddBody(ac.NewTextBlock("Work Packages").SetWeight(ac.WeightBolder)).
		AddBody(
			ac.NewTable().
				AddColumn(1).
				AddColumn(3).
				AddColumn(1).
				AddTextRow("ID", "Subject", "Status").
				AddTextRow("#1", "Fix login bug", "Open").
				AddTextRow("#2", "Add dark mode", "In Progress"),
		)

	data, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	t.Logf("Table card:\n%s", data)
}

func TestAllElementTypes(t *testing.T) {
	// Verify all element types can be added to a card and serialized.
	card := ac.NewCard().
		AddBody(ac.NewTextBlock("text")).
		AddBody(ac.NewImage("https://example.com/img.png")).
		AddBody(ac.NewRichTextBlock(ac.NewTextRun("rich"))).
		AddBody(ac.NewMedia(ac.NewMediaSource("video/mp4", "https://example.com/v.mp4"))).
		AddBody(ac.NewContainer(ac.NewTextBlock("inside container"))).
		AddBody(ac.NewColumnSet(
			ac.NewColumn(ac.NewTextBlock("col1")),
			ac.NewColumn(ac.NewTextBlock("col2")),
		)).
		AddBody(ac.NewFactSet(ac.Fact{Title: "k", Value: "v"})).
		AddBody(ac.NewImageSet(ac.NewImage("https://example.com/a.png"))).
		AddBody(ac.NewActionSet(ac.NewActionSubmit("btn", nil))).
		AddBody(ac.NewTable().AddColumn(1).AddTextRow("cell")).
		AddBody(ac.NewInputText("t").SetLabel("Text")).
		AddBody(ac.NewInputNumber("n").SetLabel("Number")).
		AddBody(ac.NewInputDate("d").SetLabel("Date")).
		AddBody(ac.NewInputTime("tm").SetLabel("Time")).
		AddBody(ac.NewInputToggle("tg", "Toggle")).
		AddBody(ac.NewInputChoiceSet("cs").AddChoice("A", "a").AddChoice("B", "b"))

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty")
	}
}

func TestAllActionTypes(t *testing.T) {
	card := ac.NewCard().
		AddAction(ac.NewActionOpenURL("Open", "https://example.com")).
		AddAction(ac.NewActionSubmit("Submit", map[string]any{"key": "val"})).
		AddAction(ac.NewActionShowCard("Show", ac.NewCard().AddBody(ac.NewTextBlock("sub")))).
		AddAction(ac.NewActionToggleVisibility("Toggle", ac.TargetElement{ElementID: "el1"})).
		AddAction(ac.NewActionExecute("Execute").SetVerb("doSomething").SetData(map[string]any{"x": 1})).
		AddAction(ac.NewActionOpenMiniApp("Mini App", "https://example.com/app"))

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ac.Card
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// 5 standard + 1 nexus extension (now all decoded as typed actions)
	if len(decoded.Actions) != 6 {
		t.Errorf("actions len = %d, want 6", len(decoded.Actions))
	}
}

func TestNestedUnmarshalRoundTrip(t *testing.T) {
	// Build a card with deeply nested interface fields.
	card := ac.NewCard().
		AddBody(
			ac.NewContainer(
				ac.NewTextBlock("inside"),
			).SetSelectAction(ac.NewActionOpenURL("Go", "https://example.com")),
		).
		AddBody(
			ac.NewColumnSet(
				ac.NewColumn(
					ac.NewImage("https://example.com/img.png").
						SetSelectAction(ac.NewActionSubmit("Click", map[string]any{"v": 1})),
				).SetSelectAction(ac.NewActionOpenURL("Col", "https://example.com")),
			).SetSelectAction(ac.NewActionOpenURL("CS", "https://example.com")),
		).
		AddBody(
			ac.NewActionSet(
				ac.NewActionExecute("Exec").SetVerb("test"),
				ac.NewActionOpenMiniApp("App", "https://example.com/app"),
			),
		).
		AddBody(
			ac.NewTable().
				AddColumn(1).AddColumn(2).
				AddRow(
					ac.NewTableCell(ac.NewTextBlock("cell1")),
					ac.NewTableCell(ac.NewInputText("inp").SetInlineAction(ac.NewActionSubmit("Go", nil))),
				),
		).
		AddBody(
			ac.NewRichTextBlock(
				ac.NewTextRun("bold").SetBold().SetSelectAction(ac.NewActionOpenURL("Link", "https://example.com")),
			),
		).
		AddBody(
			ac.NewInputText("search").SetInlineAction(ac.NewActionSubmit("Search", nil)),
		)

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ac.Card
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Body) != 6 {
		t.Fatalf("body len = %d, want 6", len(decoded.Body))
	}

	// Container with selectAction.
	container, ok := decoded.Body[0].(*ac.Container)
	if !ok {
		t.Fatalf("body[0] type = %T, want *Container", decoded.Body[0])
	}
	if len(container.Items) != 1 {
		t.Errorf("container items = %d, want 1", len(container.Items))
	}
	if container.SelectAction == nil {
		t.Error("container selectAction is nil")
	}

	// ColumnSet → Column → Image with selectAction.
	cs, ok := decoded.Body[1].(*ac.ColumnSet)
	if !ok {
		t.Fatalf("body[1] type = %T, want *ColumnSet", decoded.Body[1])
	}
	if cs.SelectAction == nil {
		t.Error("columnset selectAction is nil")
	}
	if len(cs.Columns) != 1 {
		t.Fatalf("columns = %d, want 1", len(cs.Columns))
	}
	if cs.Columns[0].SelectAction == nil {
		t.Error("column selectAction is nil")
	}
	if len(cs.Columns[0].Items) != 1 {
		t.Fatalf("column items = %d, want 1", len(cs.Columns[0].Items))
	}
	img, ok := cs.Columns[0].Items[0].(*ac.Image)
	if !ok {
		t.Fatalf("column item type = %T, want *Image", cs.Columns[0].Items[0])
	}
	if img.SelectAction == nil {
		t.Error("image selectAction is nil")
	}

	// ActionSet with typed actions.
	as, ok := decoded.Body[2].(*ac.ActionSet)
	if !ok {
		t.Fatalf("body[2] type = %T, want *ActionSet", decoded.Body[2])
	}
	if len(as.Actions) != 2 {
		t.Errorf("actionset actions = %d, want 2", len(as.Actions))
	}

	// RichTextBlock → TextRun with selectAction.
	rtb, ok := decoded.Body[4].(*ac.RichTextBlock)
	if !ok {
		t.Fatalf("body[4] type = %T, want *RichTextBlock", decoded.Body[4])
	}
	if len(rtb.Inlines) != 1 {
		t.Fatalf("inlines = %d, want 1", len(rtb.Inlines))
	}

	// InputText with inlineAction.
	inp, ok := decoded.Body[5].(*ac.InputText)
	if !ok {
		t.Fatalf("body[5] type = %T, want *InputText", decoded.Body[5])
	}
	if inp.InlineAction == nil {
		t.Error("inputtext inlineAction is nil")
	}

	// Re-marshal and verify it's valid JSON.
	data2, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if len(data2) == 0 {
		t.Fatal("re-marshal produced empty output")
	}
}
