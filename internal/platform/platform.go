package platform

import "github.com/slwk3000/post-it/internal/model"

// WindowManager defines the interface for managing floating sticky note windows
// across desktop environments (macOS Cocoa, Linux X11/Wayland/WebKitGTK).
type WindowManager interface {
	CreateNoteWindow(note *model.Note) error
	CloseNoteWindow(noteID string)
	UpdateNoteWindow(note *model.Note) error
	ToggleAllNotes() bool
	SetNotesVisible(visible bool)
	AreNotesVisible() bool
	OpenMenu(settings *model.Settings) error
	CloseMenu()
	ToggleMenu(settings *model.Settings) error
	StartDrag(noteID string)
	MovePanel(noteID string, dx, dy float64)
	ResizePanel(noteID string, dw, dh float64)
}
