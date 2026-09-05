package macos

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework WebKit -framework Carbon
#include "bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"sync"
	"unsafe"
)

type WebActionHandler func(panelID string, action string, rawPayload json.RawMessage)
type SimpleHandler func()
type PanelMovedHandler func(panelID string, x, y float64)
type PanelResizedHandler func(panelID string, w, h float64)

var (
	handlerMu     sync.RWMutex
	actionHandler WebActionHandler
	newNoteH      SimpleHandler
	toggleNotesH  SimpleHandler
	openMenuH     SimpleHandler
	appReopenH    SimpleHandler
	deleteNoteH   SimpleHandler
	panelMovedH   PanelMovedHandler
	panelResizedH PanelResizedHandler
)

func SetDeleteHandler(h SimpleHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	deleteNoteH = h
}

func SetWebActionHandler(h WebActionHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	actionHandler = h
}

func SetPanelMovedHandler(h PanelMovedHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	panelMovedH = h
}

func SetPanelResizedHandler(h PanelResizedHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	panelResizedH = h
}

func SetTrayHandlers(onNew, onToggle, onMenu, onReopen SimpleHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	newNoteH = onNew
	toggleNotesH = onToggle
	openMenuH = onMenu
	appReopenH = onReopen
}

type webMsg struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

//export onWebMessage
func onWebMessage(panelID *C.char, messageJSON *C.char) {
	pID := C.GoString(panelID)
	mJSON := C.GoString(messageJSON)

	var msg webMsg
	if err := json.Unmarshal([]byte(mJSON), &msg); err != nil {
		return
	}

	handlerMu.RLock()
	h := actionHandler
	handlerMu.RUnlock()

	if h != nil {
		h(pID, msg.Action, msg.Payload)
	}
}

//export onTrayNewNote
func onTrayNewNote() {
	handlerMu.RLock()
	h := newNoteH
	handlerMu.RUnlock()
	if h != nil {
		h()
	}
}

//export onTrayToggleNotes
func onTrayToggleNotes() {
	handlerMu.RLock()
	h := toggleNotesH
	handlerMu.RUnlock()
	if h != nil {
		h()
	}
}

//export onTrayOpenMenu
func onTrayOpenMenu() {
	handlerMu.RLock()
	h := openMenuH
	handlerMu.RUnlock()
	if h != nil {
		h()
	}
}

//export onAppReopen
func onAppReopen() {
	handlerMu.RLock()
	h := appReopenH
	handlerMu.RUnlock()
	if h != nil {
		h()
	}
}

//export onPanelMoved
func onPanelMoved(panelID *C.char, x C.double, y C.double) {
	pID := C.GoString(panelID)
	handlerMu.RLock()
	h := panelMovedH
	handlerMu.RUnlock()
	if h != nil {
		h(pID, float64(x), float64(y))
	}
}

//export onPanelResized
func onPanelResized(panelID *C.char, w C.double, h C.double) {
	pID := C.GoString(panelID)
	handlerMu.RLock()
	cb := panelResizedH
	handlerMu.RUnlock()
	if cb != nil {
		cb(pID, float64(w), float64(h))
	}
}

//export onHotKeyTriggered
func onHotKeyTriggered(id C.int) {
	handlerMu.RLock()
	newH := newNoteH
	toggleH := toggleNotesH
	openH := openMenuH
	delH := deleteNoteH
	handlerMu.RUnlock()

	switch int(id) {
	case 1: // Toggle notes (Cmd+Shift+P)
		if toggleH != nil {
			toggleH()
		}
	case 2: // New note (Cmd+Shift+N)
		if newH != nil {
			newH()
		}
	case 3: // Open Ajustes (Cmd+Shift+A)
		if openH != nil {
			openH()
		}
	case 4: // Delete note (Cmd+Shift+D)
		if delH != nil {
			delH()
		}
	}
}

func RegisterHotkeys() {
	C.macosRegisterHotkeys()
}

// Public API wrappers for Go
func InitApp() {
	C.macosInitApp()
}

func RunApp() {
	C.macosRunApp()
}

func TerminateApp() {
	C.macosTerminateApp()
}

func CreatePanel(id string, x, y, w, h float64, html string, isFloating bool) unsafe.Pointer {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))

	flt := 0
	if isFloating {
		flt = 1
	}

	return C.macosCreatePanel(cID, C.double(x), C.double(y), C.double(w), C.double(h), cHTML, C.int(flt))
}

func UpdatePanelHTML(panelPtr unsafe.Pointer, html string) {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	C.macosUpdatePanelHTML(panelPtr, cHTML)
}

func EvaluateJS(panelPtr unsafe.Pointer, script string) {
	cScript := C.CString(script)
	defer C.free(unsafe.Pointer(cScript))
	C.macosEvaluateJS(panelPtr, cScript)
}

func SetPanelVisible(panelPtr unsafe.Pointer, visible bool) {
	v := 0
	if visible {
		v = 1
	}
	C.macosSetPanelVisible(panelPtr, C.int(v))
}

func SetPanelFrame(panelPtr unsafe.Pointer, x, y, w, h float64) {
	C.macosSetPanelFrame(panelPtr, C.double(x), C.double(y), C.double(w), C.double(h))
}

func GetPanelFrame(panelPtr unsafe.Pointer) (float64, float64, float64, float64) {
	var x, y, w, h C.double
	C.macosGetPanelFrame(panelPtr, &x, &y, &w, &h)
	return float64(x), float64(y), float64(w), float64(h)
}

func ClosePanel(panelPtr unsafe.Pointer) {
	C.macosClosePanel(panelPtr)
}

func StartDrag(panelPtr unsafe.Pointer) {
	C.macosStartDrag(panelPtr)
}

func MovePanel(panelPtr unsafe.Pointer, dx, dy float64) {
	C.macosMovePanel(panelPtr, C.double(dx), C.double(dy))
}

func ResizePanel(panelPtr unsafe.Pointer, dw, dh float64) {
	C.macosResizePanel(panelPtr, C.double(dw), C.double(dh))
}

func SetupTray(title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.macosSetupTray(cTitle)
}

func GetMousePos() (float64, float64) {
	var x, y C.double
	C.macosGetMousePos(&x, &y)
	return float64(x), float64(y)
}
