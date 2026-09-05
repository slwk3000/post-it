package model

import "time"

type PaperType string

const (
	PaperPolen   PaperType = "polen"
	PaperSulfite PaperType = "sulfite"
	PaperCouche  PaperType = "couche"
	PaperKraft   PaperType = "kraft"
)

type PaperPattern string

const (
	PatternDotted PaperPattern = "dotted"
	PatternGrid   PaperPattern = "grid"
	PatternLines  PaperPattern = "lines"
	PatternPlain  PaperPattern = "plain"
)

type PenColor string

const (
	PenBlue  PenColor = "blue"
	PenBlack PenColor = "black"
	PenRed   PenColor = "red"
	PenWhite PenColor = "white"
)

type Alignment string

const (
	AlignCorner  Alignment = "corner"
	AlignCenter  Alignment = "center"
	AlignJustify Alignment = "justify"
)

type Note struct {
	ID           string       `json:"id"`
	Content      string       `json:"content"`
	X            float64      `json:"x"`
	Y            float64      `json:"y"`
	Width        float64      `json:"width"`
	Height       float64      `json:"height"`
	PaperType    PaperType    `json:"paper_type"`
	PaperPattern PaperPattern `json:"paper_pattern"`
	PenColor     PenColor     `json:"pen_color"`
	Alignment    Alignment    `json:"alignment"`
	Color        string       `json:"color"`
	Saturation   int          `json:"saturation"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Settings struct {
	DefaultPaperType    PaperType    `json:"default_paper_type"`
	DefaultPaperPattern PaperPattern `json:"default_paper_pattern"`
	DefaultPenColor     PenColor     `json:"default_pen_color"`
	DefaultAlignment    Alignment    `json:"default_alignment"`
	DefaultColor        string       `json:"default_color"`
	DefaultSaturation   int          `json:"default_saturation"`
	MenuPaperType       PaperType    `json:"menu_paper_type"`
	ShakeEnabled        bool         `json:"shake_enabled"`
	ShakeSensitivity    string       `json:"shake_sensitivity"`
}

func DefaultSettings() *Settings {
	return &Settings{
		DefaultPaperType:    PaperPolen,
		DefaultPaperPattern: PatternDotted,
		DefaultPenColor:     PenBlue,
		DefaultAlignment:    AlignCorner,
		DefaultColor:        "#fcf5e5",
		DefaultSaturation:   80,
		MenuPaperType:       PaperPolen,
		ShakeEnabled:        true,
		ShakeSensitivity:    "normal",
	}
}

func NewNote(id string, s *Settings) *Note {
	now := time.Now()
	if s == nil {
		s = DefaultSettings()
	}

	return &Note{
		ID:           id,
		Content:      "",
		X:            100,
		Y:            100,
		Width:        280,
		Height:       260,
		PaperType:    s.DefaultPaperType,
		PaperPattern: s.DefaultPaperPattern,
		PenColor:     s.DefaultPenColor,
		Alignment:    s.DefaultAlignment,
		Color:        s.DefaultColor,
		Saturation:   s.DefaultSaturation,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
