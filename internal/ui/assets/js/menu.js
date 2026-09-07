let currentSettings = null;
let changeTimeout = null;

function sendAction(action, payload = {}) {
  const msg = JSON.stringify({ action, payload });
  if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.postit) {
    window.webkit.messageHandlers.postit.postMessage(msg);
  } else {
    console.log("Menu Action:", action, payload);
  }
}

function initMenu(settings) {
  currentSettings = settings;
  updateCardStyles();

  // Header Dragging
  const header = document.getElementById("drag-header");
  if (header) {
    header.addEventListener("mousedown", (e) => {
      if (e.target.closest("button")) return;
      sendAction("drag_start", { id: "menu" });
    });
  }

  // Hover activation
  window.addEventListener("mouseenter", () => {
    sendAction("panel_hover", { id: "menu" });
  });

  // Keyboard Shortcuts inside menu
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
        sendAction("close_menu", {});
      } else if (key === "u") {
        e.preventDefault();
        sendAction("next_note", {});
      } else if (key === "r") {
        e.preventDefault();
        sendAction("prev_note", {});
      } else if (key === "n") {
        e.preventDefault();
        sendAction("new_note", {});
      } else if (key === "p") {
        e.preventDefault();
        sendAction("toggle_all", {});
      }
    } else if (e.key === "Escape") {
      e.preventDefault();
      sendAction("close_menu", {});
    }
  });

  // Radio Groups
  bindRadioGroup("paper_type", (val) => {
    currentSettings.default_paper_type = val;
    currentSettings.menu_paper_type = val;
    updateCardStyles();
    onSettingsChanged();
  });

  bindRadioGroup("paper_pattern", (val) => {
    currentSettings.default_paper_pattern = val;
    updateCardStyles();
    onSettingsChanged();
  });

  bindRadioGroup("pen_color", (val) => {
    currentSettings.default_pen_color = val;
    updateCardStyles();
    onSettingsChanged();
  });

  bindRadioGroup("alignment", (val) => {
    currentSettings.default_alignment = val;
    onSettingsChanged();
  });

  // Shake Checkbox
  const shakeToggle = document.getElementById("shake-toggle");
  if (shakeToggle) {
    shakeToggle.addEventListener("change", (e) => {
      currentSettings.shake_enabled = e.target.checked;
      onSettingsChanged();
    });
  }

  // Action Buttons
  const btnClose = document.getElementById("btn-close-menu");
  if (btnClose) {
    btnClose.addEventListener("click", () => {
      sendAction("close_menu", {});
    });
  }

  const btnNew = document.getElementById("btn-new-note");
  if (btnNew) {
    btnNew.addEventListener("click", () => {
      sendAction("new_note", {});
    });
  }

  const btnToggle = document.getElementById("btn-toggle-all");
  if (btnToggle) {
    btnToggle.addEventListener("click", () => {
      sendAction("toggle_all", {});
    });
  }

  // Initialize all Drawably hand-drawn components
  if (window.drawably) {
    const d = window.drawably;

    // Radios
    document.querySelectorAll(".radio-wrap").forEach((el) => {
      try { d.drawablyRadio(el); } catch (err) {}
    });

    // Checkboxes
    const wrapShake = document.getElementById("wrap-shake");
    if (wrapShake) {
      try { d.drawablyCheckbox(wrapShake); } catch (err) {}
    }

    // Dividers
    document.querySelectorAll(".drawably-divider").forEach((el) => {
      try { d.drawablyDivider(el); } catch (err) {}
    });

    // Section labels and text decorations
    const lblPapel = document.getElementById("lbl-papel");
    if (lblPapel) {
      try { d.drawablyHighlight(lblPapel); } catch (err) {}
    }

    const lblPauta = document.getElementById("lbl-pauta");
    if (lblPauta) {
      try { d.drawablyUnderline(lblPauta); } catch (err) {}
    }

    const lblCaneta = document.getElementById("lbl-caneta");
    if (lblCaneta) {
      try { d.drawablyHighlight(lblCaneta); } catch (err) {}
    }

    const lblAlign = document.getElementById("lbl-align");
    if (lblAlign) {
      try { d.drawablyUnderline(lblAlign); } catch (err) {}
    }

    const lblMouse = document.getElementById("lbl-mouse");
    if (lblMouse) {
      try { d.drawablyUnderline(lblMouse); } catch (err) {}
    }

    const lblKeymaps = document.getElementById("lbl-keymaps");
    if (lblKeymaps) {
      try { d.drawablyHighlight(lblKeymaps); } catch (err) {}
    }

    // Shortcuts List
    const shortcutsList = document.getElementById("shortcuts-list");
    if (shortcutsList) {
      try { d.drawablyList(shortcutsList, { marker: "dash" }); } catch (err) {}
    }

    // Buttons
    document.querySelectorAll(".drawably-btn").forEach((btn) => {
      try { d.drawablyButton(btn, { variant: "outline" }); } catch (err) {}
    });
  }
}

function bindRadioGroup(name, onChange) {
  const radios = document.querySelectorAll(`input[type="radio"][name="${name}"]`);
  radios.forEach((radio) => {
    radio.addEventListener("change", (e) => {
      if (e.target.checked) {
        onChange(e.target.value);
      }
    });
  });
}

function updateCardStyles() {
  const card = document.getElementById("config-card");
  if (!card || !currentSettings) return;
  const paper = currentSettings.default_paper_type || currentSettings.menu_paper_type || "polen";
  const pattern = currentSettings.default_paper_pattern || "plain";
  const pen = currentSettings.default_pen_color || "blue";
  card.className = "config-postit-card";
  card.classList.add(`paper-${paper}`);
  card.classList.add(`pattern-${pattern}`);
  card.classList.add(`pen-${pen}`);
}

function onSettingsChanged() {
  sendAction("save_settings", currentSettings);
}

window.initMenu = initMenu;
