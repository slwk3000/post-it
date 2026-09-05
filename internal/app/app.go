package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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

	// Load existing notes
	existingNotes, err := a.store.LoadNotes()
	if err != nil {
		log.Printf("Error loading notes: %v", err)
	}

	// If no notes exist yet, create initial welcome note
	if len(existingNotes) == 0 {
		welcomeNote := model.NewNote(generateID(), a.settings)
		welcomeNote.Content = "Olá! 👋 Bem-vindo ao Post-it!\n\n• Arraste a nota pelo topo\n• Balance o mouse para esconder/mostrar\n• Cmd+Shift+N para nova nota\n• ⚙️ para abrir o menu de papéis"
		welcomeNote.X = 140
		welcomeNote.Y = 140
		existingNotes = []*model.Note{welcomeNote}
		if err := a.store.SaveNotes(existingNotes); err != nil {
			log.Printf("Error saving welcome note: %v", err)
		}
	}

	a.mu.Lock()
	for _, n := range existingNotes {
		a.notes[n.ID] = n
	}
	a.mu.Unlock()

	// Setup handlers
	macos.SetWebActionHandler(a.handleWebAction)
	macos.SetPanelMovedHandler(a.onNoteMoved)
	macos.SetTrayHandlers(
		func() { a.CreateNewNote() },
		func() { a.ToggleAllNotes() },
		func() { a.OpenMenu() },
		func() { a.OnAppReopen() },
	)

	// Setup macOS tray item
	macos.SetupTray("📝")

	// Register global hotkeys (Cmd+Shift+P, Cmd+Shift+N)
	macos.RegisterHotkeys()

	// Start mouse shake detector
	a.shakeDetector.Start()

	// Create windows for all notes
	for _, n := range existingNotes {
		if err := a.wm.CreateNoteWindow(n); err != nil {
			log.Printf("Error creating note window %s: %v", n.ID, err)
		}
	}

	// Run macOS runloop
	macos.RunApp()
	return nil
}

func (a *App) CreateNewNote() {
	a.mu.Lock()
	noteID := generateID()
	// Offset new note slightly based on existing count
	offset := float64((len(a.notes) % 10) * 30)
	note := model.NewNote(noteID, a.settings)
	note.X = 150 + offset
	note.Y = 150 + offset

	a.notes[noteID] = note
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()

	if err := a.store.SaveNotes(allNotes); err != nil {
		log.Printf("Save notes error: %v", err)
	}

	if err := a.wm.CreateNoteWindow(note); err != nil {
		log.Printf("Create note window error: %v", err)
	}
}

func (a *App) DeleteNote(id string) {
	a.mu.Lock()
	delete(a.notes, id)
	allNotes := a.getAllNotesSliceLocked()
	a.mu.Unlock()

	a.wm.CloseNoteWindow(id)
	if err := a.store.SaveNotes(allNotes); err != nil {
		log.Printf("Save notes error: %v", err)
	}
}

func (a *App) ToggleAllNotes() {
	a.wm.ToggleAllNotes()
}

func (a *App) OpenMenu() {
	a.mu.RLock()
	st := a.settings
	a.mu.RUnlock()
	a.wm.OpenMenu(st)
}

func (a *App) OnAppReopen() {
	// If notes are hidden, show them; otherwise toggle menu
	if !a.wm.AreNotesVisible() {
		a.wm.SetNotesVisible(true)
	} else {
		a.OpenMenu()
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

func (a *App) handleWebAction(panelID string, action string, payload json.RawMessage) {
	switch action {
	case "save_content":
		var p struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
			a.mu.Lock()
			if note, ok := a.notes[p.ID]; ok {
				note.Content = p.Content
				note.UpdatedAt = time.Now()
			}
			allNotes := a.getAllNotesSliceLocked()
			a.mu.Unlock()
			a.store.SaveNotes(allNotes)
		}

	case "drag_start":
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(payload, &p); err == nil {
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
		a.OpenMenu()

	case "close_menu":
		a.wm.CloseMenu()

	case "new_note":
		a.CreateNewNote()

	case "toggle_all":
		a.ToggleAllNotes()

	case "save_settings":
		var newSettings model.Settings
		if err := json.Unmarshal(payload, &newSettings); err == nil {
			a.mu.Lock()
			a.settings = &newSettings
			a.shakeDetector.SetEnabled(newSettings.ShakeEnabled)
			a.shakeDetector.SetSensitivity(newSettings.ShakeSensitivity)
			a.mu.Unlock()

			a.store.SaveSettings(&newSettings)

			// Update all currently open notes to reflect default styles if desired
			a.mu.RLock()
			for _, note := range a.notes {
				note.PaperType = newSettings.DefaultPaperType
				note.PaperPattern = newSettings.DefaultPaperPattern
				note.PenColor = newSettings.DefaultPenColor
				note.Alignment = newSettings.DefaultAlignment
				note.Color = newSettings.DefaultColor
				note.Saturation = newSettings.DefaultSaturation
				a.wm.UpdateNoteWindow(note)
			}
			allNotes := a.getAllNotesSliceLocked()
			a.mu.RUnlock()
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
