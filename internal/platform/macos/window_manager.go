package macos

import (
	"encoding/json"
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

	noteJSON, err := json.Marshal(note)
	if err != nil {
		return err
	}

	EvaluateJS(ptr, fmt.Sprintf("if (window.updateNoteConfig) { window.updateNoteConfig(%s); }", string(noteJSON)))
	return nil
}

func (wm *WindowManager) ToggleAllNotes() bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.notesVisible = !wm.notesVisible
	for _, ptr := range wm.notePanels {
		SetPanelVisible(ptr, wm.notesVisible)
	}
	if wm.menuPanel != nil && wm.menuVisible {
		SetPanelVisible(wm.menuPanel, wm.notesVisible)
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
	if wm.menuPanel != nil && wm.menuVisible {
		SetPanelVisible(wm.menuPanel, visible)
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
		// Center on screen, rectangular post-it size
		wm.menuPanel = CreatePanel("menu", 200, 150, 540, 440, html, true)
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

func (wm *WindowManager) IsMenuVisible() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.menuVisible
}

func (wm *WindowManager) FocusNoteWindow(noteID string) error {
	wm.mu.Lock()
	ptr, exists := wm.notePanels[noteID]
	if !exists {
		wm.mu.Unlock()
		return fmt.Errorf("note window %s not found", noteID)
	}
	if !wm.notesVisible {
		wm.notesVisible = true
		for _, p := range wm.notePanels {
			SetPanelVisible(p, true)
		}
	}
	wm.mu.Unlock()

	FocusPanel(ptr)
	return nil
}

func (wm *WindowManager) StartDrag(panelID string) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if panelID == "menu" && wm.menuPanel != nil {
		StartDrag(wm.menuPanel)
		return
	}
	if ptr, exists := wm.notePanels[panelID]; exists {
		StartDrag(ptr)
	}
}

func (wm *WindowManager) MovePanel(panelID string, dx, dy float64) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if panelID == "menu" && wm.menuPanel != nil {
		MovePanel(wm.menuPanel, dx, dy)
		return
	}
	if ptr, exists := wm.notePanels[panelID]; exists {
		MovePanel(ptr, dx, dy)
	}
}

func (wm *WindowManager) ResizePanel(panelID string, dw, dh float64) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if panelID == "menu" && wm.menuPanel != nil {
		ResizePanel(wm.menuPanel, dw, dh)
		return
	}
	if ptr, exists := wm.notePanels[panelID]; exists {
		ResizePanel(ptr, dw, dh)
	}
}

func (wm *WindowManager) GetPanelFrame(panelID string) (float64, float64, float64, float64, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if panelID == "menu" && wm.menuPanel != nil {
		x, y, w, h := GetPanelFrame(wm.menuPanel)
		return x, y, w, h, true
	}
	if ptr, exists := wm.notePanels[panelID]; exists {
		x, y, w, h := GetPanelFrame(ptr)
		return x, y, w, h, true
	}

	return 0, 0, 0, 0, false
}
