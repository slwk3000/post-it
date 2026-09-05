package ui

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"

	"github.com/slwk3000/post-it/internal/model"
)

//go:embed assets templates
var FS embed.FS

var (
	once            sync.Once
	caveatFontFace  template.CSS
	drawablyCSS     template.CSS
	paperCSS        template.CSS
	drawablyJS      template.JS
	noteJS          template.JS
	menuJS          template.JS
	noteTemplate    *template.Template
	menuTemplate    *template.Template
	initErr         error
)

func initAssets() error {
	fontData, err := FS.ReadFile("assets/fonts/Caveat.woff2")
	if err != nil {
		return fmt.Errorf("read font: %w", err)
	}
	fontB64 := base64.StdEncoding.EncodeToString(fontData)
	caveatFontFace = template.CSS(fmt.Sprintf(`
@font-face {
  font-family: 'Caveat';
  src: url('data:font/woff2;charset=utf-8;base64,%s') format('woff2');
  font-weight: 400 700;
  font-style: normal;
  font-display: swap;
}`, fontB64))

	dCSS, err := FS.ReadFile("assets/css/drawably.css")
	if err != nil {
		return fmt.Errorf("read drawably css: %w", err)
	}
	drawablyCSS = template.CSS(dCSS)

	pCSS, err := FS.ReadFile("assets/css/paper.css")
	if err != nil {
		return fmt.Errorf("read paper css: %w", err)
	}
	paperCSS = template.CSS(pCSS)

	dJS, err := FS.ReadFile("assets/js/drawably.bundle.js")
	if err != nil {
		return fmt.Errorf("read drawably js: %w", err)
	}
	drawablyJS = template.JS(dJS)

	nJS, err := FS.ReadFile("assets/js/note.js")
	if err != nil {
		return fmt.Errorf("read note js: %w", err)
	}
	noteJS = template.JS(nJS)

	mJS, err := FS.ReadFile("assets/js/menu.js")
	if err != nil {
		return fmt.Errorf("read menu js: %w", err)
	}
	menuJS = template.JS(mJS)

	noteTmplData, err := FS.ReadFile("templates/note.html")
	if err != nil {
		return fmt.Errorf("read note template: %w", err)
	}
	noteTemplate, err = template.New("note").Parse(string(noteTmplData))
	if err != nil {
		return fmt.Errorf("parse note template: %w", err)
	}

	menuTmplData, err := FS.ReadFile("templates/menu.html")
	if err != nil {
		return fmt.Errorf("read menu template: %w", err)
	}
	menuTemplate, err = template.New("menu").Parse(string(menuTmplData))
	if err != nil {
		return fmt.Errorf("parse menu template: %w", err)
	}

	return nil
}

type NoteViewData struct {
	Note           *model.Note
	NoteJSON       template.JS
	CaveatFontFace template.CSS
	DrawablyCSS    template.CSS
	PaperCSS       template.CSS
	DrawablyJS     template.JS
	NoteJS         template.JS
}

type MenuViewData struct {
	Settings       *model.Settings
	SettingsJSON   template.JS
	CaveatFontFace template.CSS
	DrawablyCSS    template.CSS
	PaperCSS       template.CSS
	DrawablyJS     template.JS
	MenuJS         template.JS
}

func RenderNoteHTML(note *model.Note) (string, error) {
	once.Do(func() {
		initErr = initAssets()
	})
	if initErr != nil {
		return "", initErr
	}

	rawJSON, err := json.Marshal(note)
	if err != nil {
		return "", fmt.Errorf("marshal note json: %w", err)
	}

	data := NoteViewData{
		Note:           note,
		NoteJSON:       template.JS(rawJSON),
		CaveatFontFace: caveatFontFace,
		DrawablyCSS:    drawablyCSS,
		PaperCSS:       paperCSS,
		DrawablyJS:     drawablyJS,
		NoteJS:         noteJS,
	}

	var buf bytes.Buffer
	if err := noteTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute note template: %w", err)
	}

	return buf.String(), nil
}

func RenderMenuHTML(settings *model.Settings) (string, error) {
	once.Do(func() {
		initErr = initAssets()
	})
	if initErr != nil {
		return "", initErr
	}

	rawJSON, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("marshal settings json: %w", err)
	}

	data := MenuViewData{
		Settings:       settings,
		SettingsJSON:   template.JS(rawJSON),
		CaveatFontFace: caveatFontFace,
		DrawablyCSS:    drawablyCSS,
		PaperCSS:       paperCSS,
		DrawablyJS:     drawablyJS,
		MenuJS:         menuJS,
	}

	var buf bytes.Buffer
	if err := menuTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute menu template: %w", err)
	}

	return buf.String(), nil
}
