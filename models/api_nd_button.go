package models

type NDButton struct {
	Command string        `json:"command,omitempty"`
	Label   string        `json:"label"`
	Data    any           `json:"data,omitempty"`
	Opts    *NDButtonOpts `json:"opts,omitempty"`
}

type NDButtonOpts struct {
	Silent          bool          `json:"silent,omitempty"`
	HSize           int           `json:"h_size,omitempty"`
	ShowAlert       bool          `json:"show_alert,omitempty"`
	AlertText       string        `json:"alert_text,omitempty"`
	FontColor       string        `json:"font_color,omitempty"`
	BackgroundColor string        `json:"background_color,omitempty"`
	Align           NDButtonAlign `json:"align,omitempty"`
	Link            string        `json:"link,omitempty"`
	Handler         string        `json:"handler,omitempty"`
}

type NDButtonAlign string

const (
	AlignLeft   NDButtonAlign = NDButtonAlign("left")
	AlignCenter NDButtonAlign = NDButtonAlign("center")
	AlignRight  NDButtonAlign = NDButtonAlign("right")
)

type NDButtonOption func(b *NDButton)

func NewLinkButton(label string, url string, options ...NDButtonOption) NDButton {
	b := NDButton{
		Label: label,
		Opts: &NDButtonOpts{
			Link:    url,
			Handler: "client",
		},
	}
	for _, opt := range options {
		opt(&b)
	}
	return b
}

func WithButtonFontColor(color string) NDButtonOption {
	return func(b *NDButton) {
		b.Opts.FontColor = color
	}
}

func WithButtonBackgroundColor(color string) NDButtonOption {
	return func(b *NDButton) {
		b.Opts.BackgroundColor = color
	}
}

func WithButtonContentAlign(align NDButtonAlign) NDButtonOption {
	return func(b *NDButton) {
		b.Opts.Align = align
	}
}

func WithButtonHorizontalSize(hSize int) NDButtonOption {
	return func(b *NDButton) {
		b.Opts.HSize = hSize
	}
}

func WithButtonDisableAlert() NDButtonOption {
	return func(b *NDButton) {
		b.Opts.AlertText = ""
		b.Opts.ShowAlert = false
	}
}

func WithButtonAlert(text string) NDButtonOption {
	return func(b *NDButton) {
		b.Opts.AlertText = text
		b.Opts.ShowAlert = true
	}
}
