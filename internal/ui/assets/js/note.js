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

function createCheckboxElement(checked, onChange) {
  const wrap = document.createElement("span");
  wrap.className = "note-checkbox-wrap";
  wrap.contentEditable = "false";

  const checkWrap = document.createElement("span");
  checkWrap.className = "check-wrap";

  const input = document.createElement("input");
  input.type = "checkbox";
  input.checked = !!checked;

  checkWrap.appendChild(input);
  wrap.appendChild(checkWrap);

  if (window.drawably && window.drawably.drawablyCheckbox) {
    try {
      window.drawably.drawablyCheckbox(checkWrap);
    } catch (err) {
      console.error(err);
    }
  }

  input.addEventListener("change", () => {
    const line = wrap.closest(".note-line");
    if (line) {
      line.dataset.checked = input.checked ? "true" : "false";
    }
    if (typeof onChange === "function") {
      onChange();
    }
  });

  return wrap;
}

function createLineElement(type, text, checked, onChange) {
  const line = document.createElement("div");
  line.className = "note-line";
  line.dataset.type = type;

  if (type === "checkbox") {
    line.dataset.checked = checked ? "true" : "false";
    const cb = createCheckboxElement(checked, onChange);
    line.appendChild(cb);
  }

  const textSpan = document.createElement("span");
  textSpan.className = "note-line-text";
  textSpan.contentEditable = "true";
  textSpan.spellcheck = false;
  textSpan.textContent = text || "";
  line.appendChild(textSpan);

  return line;
}

function setCaretToEnd(el) {
  if (!el) return;
  el.focus();
  const range = document.createRange();
  const sel = window.getSelection();
  range.selectNodeContents(el);
  range.collapse(false);
  sel.removeAllRanges();
  sel.addRange(range);
}

function setCaretToStart(el) {
  if (!el) return;
  el.focus();
  const range = document.createRange();
  const sel = window.getSelection();
  range.selectNodeContents(el);
  range.collapse(true);
  sel.removeAllRanges();
  sel.addRange(range);
}

function getCaretOffset(element) {
  let caretOffset = 0;
  const sel = window.getSelection();
  if (sel && sel.rangeCount > 0) {
    const range = sel.getRangeAt(0);
    const preCaretRange = range.cloneRange();
    preCaretRange.selectNodeContents(element);
    preCaretRange.setEnd(range.endContainer, range.endOffset);
    caretOffset = preCaretRange.toString().length;
  }
  return caretOffset;
}

function serializeEditor(editor) {
  const lines = [];
  editor.querySelectorAll(".note-line").forEach(line => {
    const isCheckbox = line.dataset.type === "checkbox";
    const textSpan = line.querySelector(".note-line-text");
    const text = textSpan ? textSpan.textContent : "";
    if (isCheckbox) {
      const checked = line.dataset.checked === "true";
      lines.push("- [" + (checked ? "x" : " ") + "] " + text);
    } else {
      lines.push(text);
    }
  });
  return lines.join("\n");
}

function renderEditor(editor, rawText, onChange) {
  editor.innerHTML = "";
  const lines = (rawText || "").split("\n");
  if (lines.length === 0 || (lines.length === 1 && lines[0] === "")) {
    editor.appendChild(createLineElement("text", "", false, onChange));
    return;
  }
  for (const rawLine of lines) {
    const cbUnchecked = rawLine.match(/^-\s*\[\s*\]\s*(.*)$/);
    const cbChecked = rawLine.match(/^-\s*\[[xX]\]\s*(.*)$/);
    if (cbChecked) {
      editor.appendChild(createLineElement("checkbox", cbChecked[1], true, onChange));
    } else if (cbUnchecked) {
      editor.appendChild(createLineElement("checkbox", cbUnchecked[1], false, onChange));
    } else {
      editor.appendChild(createLineElement("text", rawLine, false, onChange));
    }
  }
}

function initNote(noteData) {
  currentNote = noteData;
  applyNoteConfig(noteData);

  const editor = document.getElementById("note-text");
  const contentArea = document.getElementById("postit-content");

  const flushSave = () => {
    clearTimeout(saveTimeout);
    if (currentNote && editor) {
      sendAction("save_content", {
        id: currentNote.id,
        content: serializeEditor(editor)
      });
    }
  };

  const scheduleSave = () => {
    clearTimeout(saveTimeout);
    saveTimeout = setTimeout(flushSave, 80);
  };

  if (editor) {
    renderEditor(editor, noteData.content || "", scheduleSave);

    // Cocoa bridge compatibility properties
    Object.defineProperty(editor, "value", {
      get() { return serializeEditor(editor); },
      set(val) { renderEditor(editor, val, scheduleSave); },
      configurable: true
    });

    editor.setSelectionRange = function() {
      const last = editor.lastElementChild;
      if (last) {
        const span = last.querySelector(".note-line-text");
        if (span) setCaretToEnd(span);
      }
    };

    editor.addEventListener("input", scheduleSave);
    editor.addEventListener("blur", flushSave);
    window.addEventListener("beforeunload", flushSave);
    window.addEventListener("pagehide", flushSave);

    // Keyboard navigation and shortcuts inside note
    editor.addEventListener("keydown", (e) => {
      // Cmd + Q -> Quit
      if (e.metaKey && e.key.toLowerCase() === "q") {
        e.preventDefault();
        flushSave();
        sendAction("quit_app", {});
        return;
      }

      // Cmd + Shift shortcuts
      if (e.metaKey && e.shiftKey) {
        const key = e.key.toLowerCase();
        if (key === "a") {
          e.preventDefault();
          sendAction("toggle_menu", { id: currentNote.id });
          return;
        } else if (key === "u") {
          e.preventDefault();
          sendAction("next_note", { id: currentNote.id });
          return;
        } else if (key === "r") {
          e.preventDefault();
          sendAction("prev_note", { id: currentNote.id });
          return;
        } else if (key === "d" || key === "x") {
          e.preventDefault();
          sendAction("delete_note", { id: currentNote.id });
          return;
        } else if (key === "n") {
          e.preventDefault();
          sendAction("new_note", { id: currentNote.id });
          return;
        } else if (key === "p") {
          e.preventDefault();
          sendAction("toggle_all", {});
          return;
        }
      }

      const textSpan = e.target.closest(".note-line-text");
      if (!textSpan) return;
      const line = textSpan.closest(".note-line");
      if (!line) return;

      // 1. Shift + Enter: Toggle checkbox on current line
      if (e.key === "Enter" && e.shiftKey) {
        e.preventDefault();
        if (line.dataset.type === "checkbox") {
          const input = line.querySelector("input[type='checkbox']");
          if (input) {
            input.checked = !input.checked;
            input.dispatchEvent(new Event("change", { bubbles: true }));
            scheduleSave();
          }
        }
        return;
      }

      // 2. Tab: if line text starts with '-' (or is just '-'), turn into checkbox!
      if (e.key === "Tab" && !e.shiftKey && !e.metaKey && !e.ctrlKey && !e.altKey) {
        const text = textSpan.textContent;
        const matchDash = text.match(/^\s*-\s*(.*)$/);
        if (matchDash) {
          e.preventDefault();
          const remainder = matchDash[1];
          if (line.dataset.type !== "checkbox") {
            line.dataset.type = "checkbox";
            line.dataset.checked = "false";
            const cb = createCheckboxElement(false, scheduleSave);
            line.insertBefore(cb, textSpan);
            textSpan.textContent = remainder;
            setCaretToStart(textSpan);
            scheduleSave();
          }
          return;
        }
      }

      // 3. Enter (without Shift):
      if (e.key === "Enter" && !e.shiftKey && !e.metaKey) {
        e.preventDefault();
        const isCheckbox = line.dataset.type === "checkbox";
        const text = textSpan.textContent;

        if (isCheckbox && text.trim() === "") {
          // Empty checkbox line: turn back into normal text line
          line.dataset.type = "text";
          delete line.dataset.checked;
          const cb = line.querySelector(".note-checkbox-wrap");
          if (cb) cb.remove();
          textSpan.textContent = "";
          setCaretToStart(textSpan);
          scheduleSave();
          return;
        }

        // Split text at caret position if cursor is in middle
        const caretOffset = getCaretOffset(textSpan);
        const firstPart = text.slice(0, caretOffset);
        const secondPart = text.slice(caretOffset);
        textSpan.textContent = firstPart;

        // Create new line
        const newLineType = isCheckbox ? "checkbox" : "text";
        const newLine = createLineElement(newLineType, secondPart, false, scheduleSave);
        line.after(newLine);
        const newTextSpan = newLine.querySelector(".note-line-text");
        if (newTextSpan) {
          setCaretToStart(newTextSpan);
        }
        scheduleSave();
        return;
      }

      // 4. Backspace at start of line:
      if (e.key === "Backspace") {
        const caretOffset = getCaretOffset(textSpan);
        if (caretOffset === 0) {
          if (line.dataset.type === "checkbox") {
            e.preventDefault();
            // Remove checkbox, convert back to text
            line.dataset.type = "text";
            delete line.dataset.checked;
            const cb = line.querySelector(".note-checkbox-wrap");
            if (cb) cb.remove();
            setCaretToStart(textSpan);
            scheduleSave();
            return;
          } else if (line.previousElementSibling) {
            e.preventDefault();
            const prevLine = line.previousElementSibling;
            const prevTextSpan = prevLine.querySelector(".note-line-text");
            if (prevTextSpan) {
              const prevLen = prevTextSpan.textContent.length;
              prevTextSpan.textContent += textSpan.textContent;
              line.remove();
              prevTextSpan.focus();
              const range = document.createRange();
              const sel = window.getSelection();
              if (prevTextSpan.firstChild) {
                range.setStart(prevTextSpan.firstChild, prevLen);
                range.collapse(true);
              } else {
                range.selectNodeContents(prevTextSpan);
                range.collapse(false);
              }
              sel.removeAllRanges();
              sel.addRange(range);
              scheduleSave();
            }
            return;
          }
        }
      }

      // 5. Arrow Up / Down navigation
      if (e.key === "ArrowUp") {
        const prevLine = line.previousElementSibling;
        if (prevLine) {
          const prevText = prevLine.querySelector(".note-line-text");
          if (prevText) {
            e.preventDefault();
            setCaretToEnd(prevText);
          }
        }
      } else if (e.key === "ArrowDown") {
        const nextLine = line.nextElementSibling;
        if (nextLine) {
          const nextText = nextLine.querySelector(".note-line-text");
          if (nextText) {
            e.preventDefault();
            setCaretToEnd(nextText);
          }
        }
      }
    });

    // Paste handler for multiline text / markdown checklists
    editor.addEventListener("paste", (e) => {
      const text = (e.clipboardData || window.clipboardData).getData("text/plain");
      if (!text || !text.includes("\n")) return; // Let single-line paste be default
      e.preventDefault();

      const textSpan = e.target.closest(".note-line-text");
      const currentLine = textSpan ? textSpan.closest(".note-line") : editor.lastElementChild;
      const lines = text.split("\n");

      let lastInsertedLine = currentLine;
      for (let i = 0; i < lines.length; i++) {
        const lineStr = lines[i];
        const cbUnchecked = lineStr.match(/^-\s*\[\s*\]\s*(.*)$/);
        const cbChecked = lineStr.match(/^-\s*\[[xX]\]\s*(.*)$/);
        let newLine;
        if (cbChecked) {
          newLine = createLineElement("checkbox", cbChecked[1], true, scheduleSave);
        } else if (cbUnchecked) {
          newLine = createLineElement("checkbox", cbUnchecked[1], false, scheduleSave);
        } else {
          newLine = createLineElement("text", lineStr, false, scheduleSave);
        }

        if (i === 0 && textSpan && textSpan.textContent === "") {
          currentLine.replaceWith(newLine);
          lastInsertedLine = newLine;
        } else {
          lastInsertedLine.after(newLine);
          lastInsertedLine = newLine;
        }
      }

      const lastText = lastInsertedLine.querySelector(".note-line-text");
      if (lastText) setCaretToEnd(lastText);
      scheduleSave();
    });

    // Click on empty space in container focuses last line
    if (contentArea) {
      contentArea.addEventListener("click", (e) => {
        if (e.target === contentArea || e.target === editor) {
          const lastLine = editor.lastElementChild;
          if (lastLine) {
            const textSpan = lastLine.querySelector(".note-line-text");
            if (textSpan) setCaretToEnd(textSpan);
          }
        }
      });
    }

    // Auto-focus immediately so user can start writing without any click
    setTimeout(() => {
      const lastLine = editor.lastElementChild || editor.firstElementChild;
      if (lastLine) {
        const textSpan = lastLine.querySelector(".note-line-text");
        if (textSpan) setCaretToEnd(textSpan);
      }
    }, 30);
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
      if (editor) {
        const focused = document.activeElement && document.activeElement.closest(".note-line-text");
        if (!focused) {
          const lastLine = editor.lastElementChild || editor.firstElementChild;
          if (lastLine) {
            const textSpan = lastLine.querySelector(".note-line-text");
            if (textSpan) setCaretToEnd(textSpan);
          }
        }
      }
    }
  };
  window.addEventListener("mouseenter", onHover);
  window.addEventListener("focus", onHover);
  document.addEventListener("mousedown", onHover);

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
