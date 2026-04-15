// Package adaptivecard provides typed Go structs for the Adaptive Card 1.5 schema.
// All types live in a single flat package for ergonomic imports.
//
// See https://adaptivecards.io/explorer/ for the full schema reference.
package adaptivecard

// FontSize controls the size of text.
type FontSize string

const (
	SizeDefault    FontSize = "default"
	SizeSmall      FontSize = "small"
	SizeMedium     FontSize = "medium"
	SizeLarge      FontSize = "large"
	SizeExtraLarge FontSize = "extraLarge"
)

// FontWeight controls the weight of text.
type FontWeight string

const (
	WeightDefault FontWeight = "default"
	WeightLighter FontWeight = "lighter"
	WeightBolder  FontWeight = "bolder"
)

// FontType selects the font family.
type FontType string

const (
	FontDefault   FontType = "default"
	FontMonospace FontType = "monospace"
)

// TextColor controls the color of text elements.
type TextColor string

const (
	ColorDefault   TextColor = "default"
	ColorDark      TextColor = "dark"
	ColorLight     TextColor = "light"
	ColorAccent    TextColor = "accent"
	ColorGood      TextColor = "good"
	ColorWarning   TextColor = "warning"
	ColorAttention TextColor = "attention"
)

// TextBlockStyle is the accessibility style of a TextBlock.
type TextBlockStyle string

const (
	TextStyleDefault TextBlockStyle = "default"
	TextStyleHeading TextBlockStyle = "heading"
)

// Spacing controls the gap between this element and the preceding one.
type Spacing string

const (
	SpacingDefault    Spacing = "default"
	SpacingNone       Spacing = "none"
	SpacingSmall      Spacing = "small"
	SpacingMedium     Spacing = "medium"
	SpacingLarge      Spacing = "large"
	SpacingExtraLarge Spacing = "extraLarge"
	SpacingPadding    Spacing = "padding"
)

// BlockElementHeight specifies the height of a block element.
type BlockElementHeight string

const (
	HeightAuto    BlockElementHeight = "auto"
	HeightStretch BlockElementHeight = "stretch"
)

// HorizontalAlignment controls horizontal alignment.
type HorizontalAlignment string

const (
	HAlignLeft   HorizontalAlignment = "left"
	HAlignCenter HorizontalAlignment = "center"
	HAlignRight  HorizontalAlignment = "right"
)

// VerticalAlignment controls vertical alignment.
type VerticalAlignment string

const (
	VAlignTop    VerticalAlignment = "top"
	VAlignCenter VerticalAlignment = "center"
	VAlignBottom VerticalAlignment = "bottom"
)

// VerticalContentAlignment defines how content is aligned vertically within a container.
type VerticalContentAlignment string

const (
	VContentTop    VerticalContentAlignment = "top"
	VContentCenter VerticalContentAlignment = "center"
	VContentBottom VerticalContentAlignment = "bottom"
)

// ImageSize controls the approximate size of an image.
type ImageSize string

const (
	ImageAuto    ImageSize = "auto"
	ImageStretch ImageSize = "stretch"
	ImageSmall   ImageSize = "small"
	ImageMedium  ImageSize = "medium"
	ImageLarge   ImageSize = "large"
)

// ImageStyle controls how an image is displayed.
type ImageStyle string

const (
	ImageDefault ImageStyle = "default"
	ImagePerson  ImageStyle = "person"
)

// ImageFillMode describes how a background image fills its area.
type ImageFillMode string

const (
	FillCover              ImageFillMode = "cover"
	FillRepeatHorizontally ImageFillMode = "repeatHorizontally"
	FillRepeatVertically   ImageFillMode = "repeatVertically"
	FillRepeat             ImageFillMode = "repeat"
)

// ContainerStyle defines the visual style of a container.
type ContainerStyle string

const (
	StyleDefault   ContainerStyle = "default"
	StyleEmphasis  ContainerStyle = "emphasis"
	StyleGood      ContainerStyle = "good"
	StyleAttention ContainerStyle = "attention"
	StyleWarning   ContainerStyle = "warning"
	StyleAccent    ContainerStyle = "accent"
)

// ActionStyle controls the visual style of an action.
type ActionStyle string

const (
	ActionDefault     ActionStyle = "default"
	ActionPositive    ActionStyle = "positive"
	ActionDestructive ActionStyle = "destructive"
)

// ActionMode determines whether an action is displayed as a button or in the overflow menu.
type ActionMode string

const (
	ModePrimary   ActionMode = "primary"
	ModeSecondary ActionMode = "secondary"
)

// TextInputStyle is the style hint for Input.Text.
type TextInputStyle string

const (
	InputStyleText     TextInputStyle = "text"
	InputStyleTel      TextInputStyle = "tel"
	InputStyleURL      TextInputStyle = "url"
	InputStyleEmail    TextInputStyle = "email"
	InputStylePassword TextInputStyle = "password"
)

// ChoiceInputStyle is the style hint for Input.ChoiceSet.
type ChoiceInputStyle string

const (
	ChoiceCompact  ChoiceInputStyle = "compact"
	ChoiceExpanded ChoiceInputStyle = "expanded"
	ChoiceFiltered ChoiceInputStyle = "filtered"
)

// ColumnWidth defines how a column's width is determined.
type ColumnWidth string

const (
	ColumnAuto    ColumnWidth = "auto"
	ColumnStretch ColumnWidth = "stretch"
	// Weighted widths like "1", "2" are also valid as plain strings.
)

// AssociatedInputs controls which inputs are gathered on Action.Submit / Action.Execute.
type AssociatedInputs string

const (
	InputsAuto AssociatedInputs = "auto"
	InputsNone AssociatedInputs = "none"
)
