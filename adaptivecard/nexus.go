package adaptivecard

// Nexus IM extensions for Adaptive Card.
// These action types are specific to the Nexus IM platform and are not part
// of the standard Adaptive Card schema.

// ActionOpenMiniApp opens a Mini App in the IM client's WebView.
// This is a Nexus-specific extension action.
type ActionOpenMiniApp struct {
	ActionBase
	Type       string `json:"type"`
	MiniAppURL string `json:"miniAppUrl"`
}

func (*ActionOpenMiniApp) actionType() string { return "Action.OpenMiniApp" }

func NewActionOpenMiniApp(title, url string) *ActionOpenMiniApp {
	return &ActionOpenMiniApp{
		ActionBase: ActionBase{Title: title},
		Type:       "Action.OpenMiniApp",
		MiniAppURL: url,
	}
}

func (a *ActionOpenMiniApp) SetID(id string) *ActionOpenMiniApp        { a.ID = id; return a }
func (a *ActionOpenMiniApp) SetStyle(s ActionStyle) *ActionOpenMiniApp { a.Style = s; return a }
func (a *ActionOpenMiniApp) SetTooltip(t string) *ActionOpenMiniApp    { a.Tooltip = t; return a }
