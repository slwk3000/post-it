let currentNote = null;
let saveTimeout = null;

function sendAction(action, payload = {}) {
  const msg = JSON.stringify({ action, payload });
  if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.postit) {
    window.webkit.messageHandlers.postit.postMessage(msg);
  } else {
    console.log("Action:", action, payload);
  }
}

function initNote(noteData) {
  currentNote = noteData;
  applyNoteConfig(noteData);

  const textarea = document.getElementById("note-text");
  if (textarea) {
    textarea.value = noteData.content || "";
    textarea.addEventListener("input", () => {
      clearTimeout(saveTimeout);
      saveTimeout = setTimeout(() => {
        sendAction("save_content", {
          id: currentNote.id,
          content: textarea.value
        });
      }, 250);
    });
  }

  // Header Dragging (Native macOS Cocoa Window Drag)
  const header = document.getElementById("drag-header");
  if (header) {
    header.addEventListener("mousedown", (e) => {
      if (e.target.closest("button")) return;
      sendAction("drag_start", { id: currentNote.id });
    });
  }

  // Action Buttons
  const btnDelete = document.getElementById("btn-delete");
  if (btnDelete) {
    btnDelete.addEventListener("click", (e) => {
      e.stopPropagation();
      sendAction("delete_note", { id: currentNote.id });
    });
  }

  const btnMenu = document.getElementById("btn-menu");
  if (btnMenu) {
    btnMenu.addEventListener("click", (e) => {
      e.stopPropagation();
      sendAction("open_menu", { id: currentNote.id });
    });
  }

  const btnNew = document.getElementById("btn-new");
  if (btnNew) {
    btnNew.addEventListener("click", (e) => {
      e.stopPropagation();
      sendAction("new_note", { id: currentNote.id });
    });
  }

  // Resizing
  const resizer = document.querySelector(".resize-handle");
  if (resizer) {
    let isResizing = false;
    let startX = 0;
    let startY = 0;

    resizer.addEventListener("mousedown", (e) => {
      e.stopPropagation();
      e.preventDefault();
      isResizing = true;
      startX = e.screenX;
      startY = e.screenY;
    });

    window.addEventListener("mousemove", (e) => {
      if (!isResizing) return;
      const dw = e.screenX - startX;
      const dh = e.screenY - startY;
      if (dw !== 0 || dh !== 0) {
        startX = e.screenX;
        startY = e.screenY;
        sendAction("resize_move", { id: currentNote.id, dw, dh });
      }
    });

    window.addEventListener("mouseup", () => {
      if (isResizing) {
        isResizing = false;
      }
    });
  }

  // Context Menu
  window.addEventListener("contextmenu", (e) => {
    e.preventDefault();
    sendAction("open_menu", { id: currentNote.id });
  });

  // Track active note and hover-to-focus without clicking
  const onHover = () => {
    if (currentNote && currentNote.id) {
      sendAction("panel_hover", { id: currentNote.id });
      const textarea = document.getElementById("note-text");
      if (textarea && document.activeElement !== textarea) {
        textarea.focus();
      }
    }
  };
  window.addEventListener("mouseenter", onHover);
  window.addEventListener("focus", onHover);
  document.addEventListener("mousedown", onHover);

  // Keyboard Shortcuts inside note window
  window.addEventListener("keydown", (e) => {
    if (e.metaKey && e.key.toLowerCase() === "q") {
      e.preventDefault();
      sendAction("quit_app", {});
      return;
    }
    if (e.metaKey && e.shiftKey) {
      const key = e.key.toLowerCase();
      if (key === "a") {
        e.preventDefault();
        sendAction("open_menu", { id: currentNote.id });
      } else if (key === "d" || key === "x") {
        e.preventDefault();
        sendAction("delete_note", { id: currentNote.id });
      } else if (key === "n") {
        e.preventDefault();
        sendAction("new_note", { id: currentNote.id });
      } else if (key === "p") {
        e.preventDefault();
        sendAction("toggle_all", {});
      }
    }
  });

  // Attach Drawably effects with outline buttons
  if (window.drawably && window.drawably.drawablyButton) {
    const btnNew = document.getElementById("btn-new");
    const btnMenu = document.getElementById("btn-menu");
    const btnDelete = document.getElementById("btn-delete");
    try {
      if (btnNew) window.drawably.drawablyButton(btnNew, { variant: "outline" });
      if (btnMenu) window.drawably.drawablyButton(btnMenu, { variant: "outline" });
      if (btnDelete) window.drawably.drawablyButton(btnDelete, { variant: "outline" });
    } catch (err) {}
  }
}

function applyNoteConfig(note) {
  const card = document.getElementById("postit-card");
  const textarea = document.getElementById("note-text");
  if (!card || !textarea) return;

  // Paper Type, Pattern & Pen Color on card so header buttons inherit pen color
  card.className = "postit-card";
  card.classList.add(`paper-${note.paper_type || 'polen'}`);
  card.classList.add(`pattern-${note.paper_pattern || 'dotted'}`);
  card.classList.add(`pen-${note.pen_color || 'blue'}`);

  // Pen Color & Alignment on textarea
  const alignClass = note.alignment === "corner" ? "left" : (note.alignment || "left");
  textarea.className = "postit-textarea";
  textarea.classList.add(`pen-${note.pen_color || 'blue'}`);
  textarea.classList.add(`align-${alignClass}`);
}

window.updateNoteConfig = function(updatedNote) {
  currentNote = Object.assign({}, currentNote, updatedNote);
  applyNoteConfig(currentNote);
};

window.initNote = initNote;
