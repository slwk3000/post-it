package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/slwk3000/post-it/internal/config"
	"github.com/slwk3000/post-it/internal/model"
	"github.com/slwk3000/post-it/internal/platform/macos"
	"github.com/slwk3000/post-it/internal/store"
)

type App struct {
	mu            sync.RWMutex
	cfg           *config.Config
	store         *store.Store
	wm            *macos.WindowManager
	shakeDetector *macos.ShakeDetector
	settings      *model.Settings
	notes         map[string]*model.Note
	activeNoteID  string
}

func New() (*App, error) {
	cfg, err := config.Default()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	st := store.New(cfg)
	settings, err := st.LoadSettings()
	if err != nil {
		settings = model.DefaultSettings()
	}

	wm := macos.NewWindowManager()

	a := &App{
		cfg:      cfg,
		store:    st,
		wm:       wm,
		settings: settings,
		notes:    make(map[string]*model.Note),
	}

	a.shakeDetector = macos.NewShakeDetector(func() {
		a.ToggleAllNotes()
	})
	a.shakeDetector.SetEnabled(settings.ShakeEnabled)
	a.shakeDetector.SetSensitivity(settings.ShakeSensitivity)

	return a, nil
}

func (a *App) Start() error {
	// Initialize macOS Application
	macos.InitApp()

	// Check if this is the very first time launching the app
	isFirstRun := !a.store.NotesFileExists()
	existingNotes, err := a.store.LoadNotes()
	if err != nil {
		log.Printf("Error loading notes: %v", err)
	}

	// Only show welcome note on the very first run
	if isFirstRun && len(existingNotes) == 0 {
		welcomeNote := model.NewNote(generateID(), a.settings)
		welcomeNote.Content = "Bem-vindo ao Post-it\n\n- Arraste a nota pelo topo\n- Balance o mouse para ocultar ou exibir\n- Cmd+Shift+N para criar nota\n- Botao * para abrir Ajustes\n- Cmd+Shift+U / Cmd+Shift+R para alternar notas\n- Cmd+C, Cmd+V e Cmd+A suportados"
		welcomeNote.X = 140
		welcomeNote.Y = 140
		existingNotes = []*model.Note{welcomeNote}
		if err := a.store.SaveNotes(existingNotes); err != nil {
			log.Printf("Error saving welcome note: %v", err)
		}
	} else if len(existingNotes) == 0 {
		// Subsequent runs with 0 notes: open a fresh blank note ready for typing
		blankNote := model.NewNote(generateID(), a.settings)
		blankNote.Content = ""
		blankNote.X = 140
		blankNote.Y = 140
		existingNotes = []*model.Note{blankNote}
		_ = a.store.SaveNotes(existingNotes)
	}

	a.mu.Lock()
	for _, n := range existingNotes {
		a.notes[n.ID] = n
	}
	a.mu.Unlock()

	// Setup handlers
	macos.SetWebActionHandler(a.handleWebAction)
	macos.SetPanelMovedHandler(a.onNoteMoved)
	macos.SetPanelResizedHandler(a.onNoteResized)
	macos.SetTrayHandlers(
		func() { a.CreateNewNote() },
		func() { a.ToggleAllNotes() },
		func() { a.ToggleMenu() },
		func() { a.OnAppReopen() },
	)
	macos.SetDeleteHandler(func() {
		a.DeleteActiveNote()
	})
	macos.SetNoteNavigationHandlers(
		func() { a.FocusNextNote() },
		func() { a.FocusPrevNote() },
	)
	macos.SetAppTerminateHandler(func() {
		a.mu.Lock()
		allNotes := a.getAllNotesSliceLocked()
		a.mu.Unlock()
		_ = a.store.SaveNotes(allNotes)
	})

	// Setup macOS tray item (no emojis)
	macos.SetupTray("Post-it")

	// Register global hotkeys (Cmd+Shift+P, Cmd+Shift+N, Cmd+Shift+A, Cmd+Shift+D, Cmd+Shift+U, Cmd+Shift+R)
	macos.RegisterHotkeys()

	// Start mouse shake detector
	a.shakeDetector.Start()

	// Create windows for all notes
	var lastNoteID string
	for _, n := range existingNotes {
		if err := a.wm.CreateNoteWindow(n); err != nil {
			log.Printf("Error creating note window %s: %v", n.ID, err)
		}
		lastNoteID = n.ID
	}

	// Focus the active note immediately for instant typing
	if lastNoteID != "" {
		a.activeNoteID = lastNoteID
		_ = a.wm.FocusNoteWindow(lastNoteID)
	}

	// Run macOS runloop
	macos.RunApp()
	return nil
}

func (a *App) CreateNewNote(sourceID ...string) {
	a.mu.Lock()
	noteID := generateID()
	note := model.NewNote(noteID, a.settings)

	// Determine reference note (explicit source ID, or activeNoteID, or most recently updated)
	refID := ""
	if len(sourceID) > 0 && sourceID[0] != "" {
		refID = sourceID[0]
	}
	if refID == "" {
		refID = a.activeNoteID
	}
	var refNote *model.Note
	if refID != "" {
		refNote = a.notes[refID]
	}
	if refNote == nil {
		var latestTime time.Time
		for _, n := range a.notes {
			if refNote == nil || n.UpdatedAt.After(latestTime) {
				refNote = n
				latestTime = n.UpdatedAt
			}
		}
	}

	if refNote != nil {
		note.X = refNote.X + 35
		note.Y = refNote.Y + 35
		note.PaperType = refNote.PaperType
		note.PaperPattern = refNote.PaperPattern
		note.PenColor = refNote.PenColor
		note.Alignment = refNote.Alignment
	} else {
		note.X = 180
		note.Y = 180
	}

	a.activeNoteID = noteID
	a.notes[noteID] = note
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()

	if err := a.store.SaveNotes(allNotes); err != nil {
		log.Printf("Save notes error: %v", err)
	}

	if err := a.wm.CreateNoteWindow(note); err != nil {
		log.Printf("Create note window error: %v", err)
	}

	// Immediately focus the newly created note so user can type right away
	_ = a.wm.FocusNoteWindow(note.ID)
}

func (a *App) getOrderedNoteIDsLocked() []string {
	type noteItem struct {
		id        string
		createdAt time.Time
	}
	var items []noteItem
	for id, n := range a.notes {
		items = append(items, noteItem{id: id, createdAt: n.CreatedAt})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].createdAt.Before(items[j].createdAt)
	})
	var ids []string
	for _, it := range items {
		ids = append(ids, it.id)
	}
	return ids
}

func (a *App) FocusNextNote() {
	a.mu.Lock()
	ids := a.getOrderedNoteIDsLocked()
	if len(ids) == 0 {
		a.mu.Unlock()
		return
	}
	currentIdx := -1
	for i, id := range ids {
		if id == a.activeNoteID {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(ids)
	targetID := ids[nextIdx]
	a.activeNoteID = targetID
	a.mu.Unlock()

	_ = a.wm.FocusNoteWindow(targetID)
}

func (a *App) FocusPrevNote() {
	a.mu.Lock()
	ids := a.getOrderedNoteIDsLocked()
	if len(ids) == 0 {
		a.mu.Unlock()
		return
	}
	currentIdx := -1
	for i, id := range ids {
		if id == a.activeNoteID {
			currentIdx = i
			break
		}
	}
	prevIdx := (currentIdx - 1 + len(ids)) % len(ids)
	targetID := ids[prevIdx]
	a.activeNoteID = targetID
	a.mu.Unlock()

	_ = a.wm.FocusNoteWindow(targetID)
}

func (a *App) DeleteNote(id string) {
	a.mu.Lock()
	delete(a.notes, id)
	if a.activeNoteID == id {
		a.activeNoteID = ""
	}
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()

	a.wm.CloseNoteWindow(id)
	if err := a.store.SaveNotes(allNotes); err != nil {
		log.Printf("Save notes error: %v", err)
	}
}

func (a *App) DeleteActiveNote() {
	a.mu.Lock()
	targetID := a.activeNoteID
	if targetID == "" || a.notes[targetID] == nil {
		var latestTime time.Time
		for id, n := range a.notes {
			if targetID == "" || n.UpdatedAt.After(latestTime) {
				targetID = id
				latestTime = n.UpdatedAt
			}
		}
	}
	a.mu.Unlock()

	if targetID != "" {
		a.DeleteNote(targetID)
	}
}

func (a *App) ToggleAllNotes() {
	a.wm.ToggleAllNotes()
}

func (a *App) ToggleMenu(targetID ...string) {
	if a.wm.IsMenuVisible() {
		a.wm.CloseMenu()
		return
	}
	a.OpenMenu(targetID...)
}

func (a *App) OpenMenu(targetID ...string) {
	a.mu.Lock()
	if len(targetID) > 0 && targetID[0] != "" {
		a.activeNoteID = targetID[0]
	}
	st := a.settings
	if a.activeNoteID != "" {
		if note, ok := a.notes[a.activeNoteID]; ok {
			st = &model.Settings{
				DefaultPaperType:    note.PaperType,
				DefaultPaperPattern: note.PaperPattern,
				DefaultPenColor:     note.PenColor,
				DefaultAlignment:    note.Alignment,
				DefaultColor:        note.Color,
				DefaultSaturation:   note.Saturation,
				MenuPaperType:       note.PaperType,
				ShakeEnabled:        a.settings.ShakeEnabled,
				ShakeSensitivity:    a.settings.ShakeSensitivity,
			}
		}
	}
	a.mu.Unlock()
	a.wm.OpenMenu(st)
}

func (a *App) OnAppReopen() {
	// If notes are hidden, show them; otherwise toggle menu
	if !a.wm.AreNotesVisible() {
		a.wm.SetNotesVisible(true)
	} else {
		a.ToggleMenu()
	}
}

func (a *App) onNoteMoved(id string, x, y float64) {
	a.mu.Lock()
	if note, ok := a.notes[id]; ok {
		note.X = x
		note.Y = y
		note.UpdatedAt = time.Now()
	}
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()
	a.store.SaveNotes(allNotes)
}

func (a *App) onNoteResized(id string, w, h float64) {
	a.mu.Lock()
	if note, ok := a.notes[id]; ok {
		note.Width = w
		note.Height = h
		note.UpdatedAt = time.Now()
	}
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()
	a.store.SaveNotes(allNotes)
}

func (a *App) handleWebAction(panelID string, action string, payload json.RawMessage) {
	switch action {
	case "save_content":
		var p struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			a.mu.Lock()
			a.activeNoteID = p.ID
			if note, ok := a.notes[p.ID]; ok {
				note.Content = p.Content
				note.UpdatedAt = time.Now()
			}
			allNotes := a.getAllNotesSliceLocked()
			a.mu.Unlock()
			a.store.SaveNotes(allNotes)
		}

	case "note_focus":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil && p.ID != "" {
			a.mu.Lock()
			a.activeNoteID = p.ID
			a.mu.Unlock()
		}

	case "drag_start":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			if p.ID != "menu" && p.ID != "" {
				a.mu.Lock()
				a.activeNoteID = p.ID
				a.mu.Unlock()
			}
			a.wm.StartDrag(p.ID)
		}

	case "drag_move":
		var p struct {
			ID string  `json:"id"`
			Dx float64 `json:"dx"`
			Dy float64 `json:"dy"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			a.wm.MovePanel(p.ID, p.Dx, p.Dy)
		}

	case "resize_move":
		var p struct {
			ID string  `json:"id"`
			Dw float64 `json:"dw"`
			Dh float64 `json:"dh"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			a.wm.ResizePanel(p.ID, p.Dw, p.Dh)
		}

	case "drag_end":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			if x, y, w, h, ok := a.wm.GetPanelFrame(p.ID); ok {
				a.mu.Lock()
				if note, exists := a.notes[p.ID]; exists {
					note.X = x
					note.Y = y
					note.Width = w
					note.Height = h
					note.UpdatedAt = time.Now()
				}
				allNotes := a.getAllNotesSliceLocked()
				a.mu.Unlock()
				a.store.SaveNotes(allNotes)
			}
		}

	case "delete_note":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			a.DeleteNote(p.ID)
		}

	case "open_menu":
		var p struct {
			ID string `json:"id"`
		}
		json.Unmarshal(payload, &p)
		a.OpenMenu(p.ID)

	case "toggle_menu":
		var p struct {
			ID string `json:"id"`
		}
		json.Unmarshal(payload, &p)
		a.ToggleMenu(p.ID)

	case "close_menu":
		a.wm.CloseMenu()

	case "next_note":
		a.FocusNextNote()

	case "prev_note":
		a.FocusPrevNote()

	case "new_note":
		var p struct {
			ID string `json:"id"`
		}
		json.Unmarshal(payload, &p)
		a.CreateNewNote(p.ID)

	case "panel_hover":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil && p.ID != "" && p.ID != "menu" {
			a.mu.Lock()
			a.activeNoteID = p.ID
			a.mu.Unlock()
		}

	case "toggle_all":
		a.ToggleAllNotes()

	case "quit_app":
		a.mu.Lock()
		allNotes := a.getAllNotesSliceLocked()
		a.mu.Unlock()
		_ = a.store.SaveNotes(allNotes)
		macos.TerminateApp()

	case "save_settings":
		var newSettings model.Settings
		if err := json.Unmarshal(payload, &newSettings); err == nil {
			a.mu.Lock()
			// Update default settings for future notes
			a.settings.DefaultPaperType = newSettings.DefaultPaperType
			a.settings.DefaultPaperPattern = newSettings.DefaultPaperPattern
			a.settings.DefaultPenColor = newSettings.DefaultPenColor
			a.settings.DefaultAlignment = newSettings.DefaultAlignment
			a.settings.MenuPaperType = newSettings.DefaultPaperType
			a.settings.ShakeEnabled = newSettings.ShakeEnabled
			a.settings.ShakeSensitivity = newSettings.ShakeSensitivity
			a.shakeDetector.SetEnabled(newSettings.ShakeEnabled)
			a.shakeDetector.SetSensitivity(newSettings.ShakeSensitivity)

			// Update ONLY the active note (if specified), NOT all notes!
			if a.activeNoteID != "" {
				if note, ok := a.notes[a.activeNoteID]; ok {
					note.PaperType = newSettings.DefaultPaperType
					note.PaperPattern = newSettings.DefaultPaperPattern
					note.PenColor = newSettings.DefaultPenColor
					note.Alignment = newSettings.DefaultAlignment
					note.UpdatedAt = time.Now()
					a.wm.UpdateNoteWindow(note)
				}
			}
			allNotes := a.getAllNotesSliceLocked()
			a.mu.Unlock()

			a.store.SaveSettings(a.settings)
			a.store.SaveNotes(allNotes)
		}
	}
}

func (a *App) getAllNotesSliceLocked() []*model.Note {
	list := make([]*model.Note, 0, len(a.notes))
	for _, n := range a.notes {
		list = append(list, n)
	}
	return list
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
