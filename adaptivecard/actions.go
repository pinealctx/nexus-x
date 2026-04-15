package adaptivecard

// ActionOpenURL opens a URL when invoked.
// Schema: https://adaptivecards.io/explorer/Action.OpenUrl.html
type ActionOpenURL struct {
	ActionBase
	Type string `json:"type"`
	URL  string `json:"url"`
}

func (*ActionOpenURL) actionType() string { return "Action.OpenUrl" }

func NewActionOpenURL(title, url string) *ActionOpenURL {
	return &ActionOpenURL{
		ActionBase: ActionBase{Title: title},
		Type:       "Action.OpenUrl",
		URL:        url,
	}
}

func (a *ActionOpenURL) SetID(id string) *ActionOpenURL        { a.ID = id; return a }
func (a *ActionOpenURL) SetStyle(s ActionStyle) *ActionOpenURL { a.Style = s; return a }
func (a *ActionOpenURL) SetTooltip(t string) *ActionOpenURL    { a.Tooltip = t; return a }
func (a *ActionOpenURL) SetIconURL(u string) *ActionOpenURL    { a.IconURL = u; return a }

// ActionSubmit gathers input fields and sends an event to the client.
// Schema: https://adaptivecards.io/explorer/Action.Submit.html
type ActionSubmit struct {
	ActionBase
	Type             string           `json:"type"`
	Data             any              `json:"data,omitempty"`
	AssociatedInputs AssociatedInputs `json:"associatedInputs,omitempty"`
}

func (*ActionSubmit) actionType() string { return "Action.Submit" }

func NewActionSubmit(title string, data any) *ActionSubmit {
	return &ActionSubmit{
		ActionBase: ActionBase{Title: title},
		Type:       "Action.Submit",
		Data:       data,
	}
}

func (a *ActionSubmit) SetID(id string) *ActionSubmit        { a.ID = id; return a }
func (a *ActionSubmit) SetStyle(s ActionStyle) *ActionSubmit { a.Style = s; return a }
func (a *ActionSubmit) SetTooltip(t string) *ActionSubmit    { a.Tooltip = t; return a }
func (a *ActionSubmit) SetEnabled(v bool) *ActionSubmit      { a.IsEnabled = &v; return a }
func (a *ActionSubmit) SetAssociatedInputs(v AssociatedInputs) *ActionSubmit {
	a.AssociatedInputs = v
	return a
}
func (a *ActionSubmit) SetIconURL(u string) *ActionSubmit { a.IconURL = u; return a }

// ActionShowCard defines an inline sub-card that is shown when the action is invoked.
// Schema: https://adaptivecards.io/explorer/Action.ShowCard.html
type ActionShowCard struct {
	ActionBase
	Type string `json:"type"`
	Card *Card  `json:"card,omitempty"`
}

func (*ActionShowCard) actionType() string { return "Action.ShowCard" }

func NewActionShowCard(title string, card *Card) *ActionShowCard {
	return &ActionShowCard{
		ActionBase: ActionBase{Title: title},
		Type:       "Action.ShowCard",
		Card:       card,
	}
}

func (a *ActionShowCard) SetID(id string) *ActionShowCard        { a.ID = id; return a }
func (a *ActionShowCard) SetStyle(s ActionStyle) *ActionShowCard { a.Style = s; return a }

// ActionToggleVisibility toggles the visibility of associated elements.
// Schema: https://adaptivecards.io/explorer/Action.ToggleVisibility.html
type ActionToggleVisibility struct {
	ActionBase
	Type           string          `json:"type"`
	TargetElements []TargetElement `json:"targetElements,omitempty"`
}

func (*ActionToggleVisibility) actionType() string { return "Action.ToggleVisibility" }

func NewActionToggleVisibility(title string, targets ...TargetElement) *ActionToggleVisibility {
	return &ActionToggleVisibility{
		ActionBase:     ActionBase{Title: title},
		Type:           "Action.ToggleVisibility",
		TargetElements: targets,
	}
}

func (a *ActionToggleVisibility) AddTarget(t TargetElement) *ActionToggleVisibility {
	a.TargetElements = append(a.TargetElements, t)
	return a
}

func (a *ActionToggleVisibility) SetID(id string) *ActionToggleVisibility { a.ID = id; return a }

// TargetElement identifies an element whose visibility should be toggled.
type TargetElement struct {
	ElementID string `json:"elementId"`
	IsVisible *bool  `json:"isVisible,omitempty"`
}

// ActionExecute is a universal action that sends an Invoke activity (version 1.4+).
// Schema: https://adaptivecards.io/explorer/Action.Execute.html
type ActionExecute struct {
	ActionBase
	Type             string           `json:"type"`
	Verb             string           `json:"verb,omitempty"`
	Data             any              `json:"data,omitempty"`
	AssociatedInputs AssociatedInputs `json:"associatedInputs,omitempty"`
}

func (*ActionExecute) actionType() string { return "Action.Execute" }

func NewActionExecute(title string) *ActionExecute {
	return &ActionExecute{
		ActionBase: ActionBase{Title: title},
		Type:       "Action.Execute",
	}
}

func (a *ActionExecute) SetVerb(v string) *ActionExecute { a.Verb = v; return a }
func (a *ActionExecute) SetData(d any) *ActionExecute    { a.Data = d; return a }
func (a *ActionExecute) SetAssociatedInputs(v AssociatedInputs) *ActionExecute {
	a.AssociatedInputs = v
	return a
}
func (a *ActionExecute) SetID(id string) *ActionExecute        { a.ID = id; return a }
func (a *ActionExecute) SetStyle(s ActionStyle) *ActionExecute { a.Style = s; return a }
func (a *ActionExecute) SetTooltip(t string) *ActionExecute    { a.Tooltip = t; return a }
func (a *ActionExecute) SetEnabled(v bool) *ActionExecute      { a.IsEnabled = &v; return a }
