package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/slwk3000/post-it/internal/model"
)

func TestRenderNoteHTML(t *testing.T) {
	note := &model.Note{
		ID:           "test-note-1",
		Content:      "Test note content for Post-it",
		X:            100,
		Y:            150,
		Width:        280,
		Height:       260,
		PaperType:    model.PaperPolen,
		PaperPattern: model.PatternDotted,
		PenColor:     model.PenBlue,
		Alignment:    model.AlignCorner,
		Color:        "#fcf5e5",
		Saturation:   80,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	html, err := RenderNoteHTML(note)
	if err != nil {
		t.Fatalf("RenderNoteHTML failed: %v", err)
	}

	_ = os.WriteFile("/tmp/rendered_note_app.html", []byte(html), 0644)

	if !strings.Contains(html, "Test note content for Post-it") {
		t.Errorf("expected HTML to contain note content")
	}
	if !strings.Contains(html, "Caveat") {
		t.Errorf("expected HTML to define Caveat font")
	}
	if !strings.Contains(html, "drawably") {
		t.Errorf("expected HTML to include drawably")
	}
	if !strings.Contains(html, "paper-polen") {
		t.Errorf("expected HTML to contain paper-polen class")
	}
}

func TestRenderMenuHTML(t *testing.T) {
	settings := model.DefaultSettings()

	html, err := RenderMenuHTML(settings)
	if err != nil {
		t.Fatalf("RenderMenuHTML failed: %v", err)
	}

	_ = os.WriteFile("/tmp/menu_rendered.html", []byte(html), 0644)

	if !strings.Contains(html, "post-it") {
		t.Errorf("expected HTML to contain post-it wordmark")
	}
	if !strings.Contains(html, "pólen") {
		t.Errorf("expected HTML to contain pólen paper option")
	}
	if !strings.Contains(html, "pontilhado") {
		t.Errorf("expected HTML to contain pontilhado pattern option")
	}
}
