package adaptivecard

import "encoding/json"

const (
	schemaURI = "https://adaptivecards.io/schemas/adaptive-card.json"
	version15 = "1.5"
)

// Card is the root Adaptive Card object.
// Schema: https://adaptivecards.io/explorer/AdaptiveCard.html
type Card struct {
	Type                     string                   `json:"type"`
	Version                  string                   `json:"version"`
	Schema                   string                   `json:"$schema,omitempty"`
	Body                     []Element                `json:"body,omitempty"`
	Actions                  []Action                 `json:"actions,omitempty"`
	SelectAction             Action                   `json:"selectAction,omitempty"`
	FallbackText             string                   `json:"fallbackText,omitempty"`
	BackgroundImage          *BackgroundImage         `json:"backgroundImage,omitempty"`
	MinHeight                string                   `json:"minHeight,omitempty"`
	Rtl                      *bool                    `json:"rtl,omitempty"`
	Speak                    string                   `json:"speak,omitempty"`
	Lang                     string                   `json:"lang,omitempty"`
	VerticalContentAlignment VerticalContentAlignment `json:"verticalContentAlignment,omitempty"`
	Refresh                  *Refresh                 `json:"refresh,omitempty"`
	Authentication           *Authentication          `json:"authentication,omitempty"`
	Metadata                 *Metadata                `json:"metadata,omitempty"`
}

// NewCard creates a new Adaptive Card with version 1.5 defaults.
func NewCard() *Card {
	return &Card{
		Type:    "AdaptiveCard",
		Version: version15,
		Schema:  schemaURI,
	}
}

// --- Card builder methods (all return *Card for chaining) ---

func (c *Card) AddBody(el Element) *Card                  { c.Body = append(c.Body, el); return c }
func (c *Card) AddAction(act Action) *Card                { c.Actions = append(c.Actions, act); return c }
func (c *Card) SetFallbackText(t string) *Card            { c.FallbackText = t; return c }
func (c *Card) SetMinHeight(h string) *Card               { c.MinHeight = h; return c }
func (c *Card) SetLang(lang string) *Card                 { c.Lang = lang; return c }
func (c *Card) SetSpeak(s string) *Card                   { c.Speak = s; return c }
func (c *Card) SetRtl(v bool) *Card                       { c.Rtl = &v; return c }
func (c *Card) SetSelectAction(a Action) *Card            { c.SelectAction = a; return c }
func (c *Card) SetVersion(v string) *Card                 { c.Version = v; return c }
func (c *Card) SetRefresh(r *Refresh) *Card               { c.Refresh = r; return c }
func (c *Card) SetAuthentication(a *Authentication) *Card { c.Authentication = a; return c }

func (c *Card) SetBackgroundImage(url string) *Card {
	c.BackgroundImage = &BackgroundImage{URL: url}
	return c
}

func (c *Card) SetVerticalContentAlignment(v VerticalContentAlignment) *Card {
	c.VerticalContentAlignment = v
	return c
}

// JSON serializes the card to a JSON string.
func (c *Card) JSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustJSON serializes the card to JSON, panicking on error.
func (c *Card) MustJSON() string {
	s, err := c.JSON()
	if err != nil {
		panic("adaptivecard: " + err.Error())
	}
	return s
}

// MarshalJSON implements json.Marshaler with defaults.
func (c *Card) MarshalJSON() ([]byte, error) {
	type alias Card
	cc := (*alias)(c)
	if cc.Type == "" {
		cc.Type = "AdaptiveCard"
	}
	if cc.Version == "" {
		cc.Version = version15
	}
	return json.Marshal(cc)
}

// UnmarshalJSON implements json.Unmarshaler with polymorphic body/actions.
func (c *Card) UnmarshalJSON(b []byte) error {
	fields, rest, err := extractFields(b, "body", "actions", "selectAction")
	if err != nil {
		return err
	}

	type alias Card
	var base alias
	if err := json.Unmarshal(rest, &base); err != nil {
		return err
	}

	if base.Type == "" {
		base.Type = "AdaptiveCard"
	}
	if base.Version == "" {
		base.Version = version15
	}

	if raw, ok := fields["body"]; ok {
		base.Body, err = unmarshalElements(raw)
		if err != nil {
			return err
		}
	}
	if raw, ok := fields["actions"]; ok {
		base.Actions, err = unmarshalActions(raw)
		if err != nil {
			return err
		}
	}
	if raw, ok := fields["selectAction"]; ok {
		base.SelectAction, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}

	*c = Card(base)
	return nil
}

// Refresh defines how the card can be refreshed (version 1.4+).
type Refresh struct {
	Action  Action   `json:"action,omitempty"`
	Expires string   `json:"expires,omitempty"`
	UserIDs []string `json:"userIds,omitempty"`
}

func (r *Refresh) UnmarshalJSON(data []byte) error {
	fields, rest, err := extractFields(data, "action")
	if err != nil {
		return err
	}
	type alias Refresh
	var a alias
	if err := json.Unmarshal(rest, &a); err != nil {
		return err
	}
	if raw, ok := fields["action"]; ok {
		a.Action, err = unmarshalOptionalAction(raw)
		if err != nil {
			return err
		}
	}
	*r = Refresh(a)
	return nil
}

// Authentication defines authentication requirements (version 1.4+).
type Authentication struct {
	Text                  string                 `json:"text,omitempty"`
	ConnectionName        string                 `json:"connectionName,omitempty"`
	TokenExchangeResource *TokenExchangeResource `json:"tokenExchangeResource,omitempty"`
	Buttons               []AuthCardButton       `json:"buttons,omitempty"`
}

// TokenExchangeResource for Authentication.
type TokenExchangeResource struct {
	ID         string `json:"id"`
	URI        string `json:"uri"`
	ProviderID string `json:"providerId"`
}

// AuthCardButton for Authentication.
type AuthCardButton struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Title string `json:"title,omitempty"`
	Image string `json:"image,omitempty"`
}

// Metadata holds card metadata (version 1.6+, reserved).
type Metadata struct {
	WebURL string `json:"webUrl,omitempty"`
}
