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

  // Header Dragging
  const header = document.getElementById("drag-header");
  if (header) {
    let isDragging = false;
    let startX = 0;
    let startY = 0;

    header.addEventListener("mousedown", (e) => {
      // Don't drag if clicking buttons
      if (e.target.closest("button")) return;
      isDragging = true;
      startX = e.screenX;
      startY = e.screenY;
      sendAction("drag_start", { id: currentNote.id });
    });

    window.addEventListener("mousemove", (e) => {
      if (!isDragging) return;
      const dx = e.screenX - startX;
      const dy = e.screenY - startY;
      if (dx !== 0 || dy !== 0) {
        startX = e.screenX;
        startY = e.screenY;
        sendAction("drag_move", { id: currentNote.id, dx, dy });
      }
    });

    window.addEventListener("mouseup", () => {
      if (isDragging) {
        isDragging = false;
        sendAction("drag_end", { id: currentNote.id });
      }
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
      sendAction("new_note", {});
    });
  }

  // Context Menu
  window.addEventListener("contextmenu", (e) => {
    e.preventDefault();
    sendAction("open_menu", { id: currentNote.id });
  });

  // Attach Drawably effects if available
  if (window.drawably && window.drawably.drawablyButton) {
    document.querySelectorAll(".drawably-btn").forEach((btn) => {
      try {
        window.drawably.drawablyButton(btn, { variant: "outline" });
      } catch (err) {}
    });
  }
}

function applyNoteConfig(note) {
  const card = document.getElementById("postit-card");
  const textarea = document.getElementById("note-text");
  if (!card || !textarea) return;

  // Paper Type
  card.className = "postit-card";
  card.classList.add(`paper-${note.paper_type || 'polen'}`);
  card.classList.add(`pattern-${note.paper_pattern || 'dotted'}`);

  // Custom Colors for Sulfite / Couche
  if (note.color) {
    card.style.setProperty("--custom-color", note.color);
  }
  if (note.saturation !== undefined) {
    card.style.setProperty("--custom-sat", `${note.saturation}%`);
  }

  // Pen Color
  textarea.className = "postit-textarea";
  textarea.classList.add(`pen-${note.pen_color || 'blue'}`);
  textarea.classList.add(`align-${note.alignment || 'corner'}`);
}

window.updateNoteConfig = function(updatedNote) {
  currentNote = Object.assign({}, currentNote, updatedNote);
  applyNoteConfig(currentNote);
};

window.initNote = initNote;
