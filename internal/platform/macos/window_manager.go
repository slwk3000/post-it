package macos

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/slwk3000/post-it/internal/model"
	"github.com/slwk3000/post-it/internal/ui"
)

type WindowManager struct {
	mu           sync.RWMutex
	notePanels   map[string]unsafe.Pointer
	menuPanel    unsafe.Pointer
	notesVisible bool
	menuVisible  bool
}

func NewWindowManager() *WindowManager {
	return &WindowManager{
		notePanels:   make(map[string]unsafe.Pointer),
		notesVisible: true,
		menuVisible:  false,
	}
}

func (wm *WindowManager) CreateNoteWindow(note *model.Note) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// If panel already exists, return
	if _, exists := wm.notePanels[note.ID]; exists {
		return nil
	}

	html, err := ui.RenderNoteHTML(note)
	if err != nil {
		return fmt.Errorf("render note html: %w", err)
	}

	ptr := CreatePanel(note.ID, note.X, note.Y, note.Width, note.Height, html, true)
	if ptr == nil {
		return fmt.Errorf("create note panel failed")
	}

	wm.notePanels[note.ID] = ptr
	SetPanelVisible(ptr, wm.notesVisible)
	return nil
}

func (wm *WindowManager) CloseNoteWindow(noteID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	ptr, exists := wm.notePanels[noteID]
	if !exists {
		return
	}

	ClosePanel(ptr)
	delete(wm.notePanels, noteID)
}

func (wm *WindowManager) UpdateNoteWindow(note *model.Note) error {
	wm.mu.RLock()
	ptr, exists := wm.notePanels[note.ID]
	wm.mu.RUnlock()

	if !exists {
		return nil
	}

	html, err := ui.RenderNoteHTML(note)
	if err != nil {
		return err
	}

	UpdatePanelHTML(ptr, html)
	return nil
}

func (wm *WindowManager) ToggleAllNotes() bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.notesVisible = !wm.notesVisible
	for _, ptr := range wm.notePanels {
		SetPanelVisible(ptr, wm.notesVisible)
	}
	return wm.notesVisible
}

func (wm *WindowManager) SetNotesVisible(visible bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.notesVisible = visible
	for _, ptr := range wm.notePanels {
		SetPanelVisible(ptr, visible)
	}
}

func (wm *WindowManager) AreNotesVisible() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.notesVisible
}

func (wm *WindowManager) OpenMenu(settings *model.Settings) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.menuPanel == nil {
		html, err := ui.RenderMenuHTML(settings)
		if err != nil {
			return fmt.Errorf("render menu html: %w", err)
		}
		// Center on screen
		wm.menuPanel = CreatePanel("menu", 200, 150, 420, 600, html, false)
	} else {
		html, err := ui.RenderMenuHTML(settings)
		if err == nil {
			UpdatePanelHTML(wm.menuPanel, html)
		}
		SetPanelVisible(wm.menuPanel, true)
	}

	wm.menuVisible = true
	return nil
}

func (wm *WindowManager) CloseMenu() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.menuPanel != nil {
		SetPanelVisible(wm.menuPanel, false)
		wm.menuVisible = false
	}
}

func (wm *WindowManager) ToggleMenu(settings *model.Settings) error {
	wm.mu.RLock()
	visible := wm.menuVisible
	wm.mu.RUnlock()

	if visible {
		wm.CloseMenu()
		return nil
	}
	return wm.OpenMenu(settings)
}

func (wm *WindowManager) GetPanelFrame(noteID string) (float64, float64, float64, float64, bool) {
	wm.mu.RLock()
	ptr, exists := wm.notePanels[noteID]
	wm.mu.RUnlock()

	if !exists {
		return 0, 0, 0, 0, false
	}

	x, y, w, h := GetPanelFrame(ptr)
	return x, y, w, h, true
}
