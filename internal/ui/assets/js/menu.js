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

  // Header Dragging
  const header = document.getElementById("drag-header");
  if (header) {
    header.addEventListener("mousedown", (e) => {
      if (e.target.closest("button")) return;
      sendAction("drag_start", { id: "menu" });
    });
  }

  // Radio Groups
  bindRadioGroup("paper_type", (val) => {
    currentSettings.default_paper_type = val;
    currentSettings.menu_paper_type = val;
    updateConfigPaperVisual(val);
    onSettingsChanged();
  });

  bindRadioGroup("paper_pattern", (val) => {
    currentSettings.default_paper_pattern = val;
    onSettingsChanged();
  });

  bindRadioGroup("pen_color", (val) => {
    currentSettings.default_pen_color = val;
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

    // Titles and text decorations
    const wordConfigs = document.getElementById("word-configs");
    if (wordConfigs) {
      try { d.drawablyCircle(wordConfigs); } catch (err) {}
    }

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

function updateConfigPaperVisual(paperType) {
  const card = document.getElementById("config-card");
  if (!card) return;
  card.className = "config-postit-card";
  card.classList.add(`paper-${paperType}`);
  card.classList.add("pattern-plain");
  card.classList.add("pen-blue");
}

function onSettingsChanged() {
  clearTimeout(changeTimeout);
  changeTimeout = setTimeout(() => {
    sendAction("save_settings", currentSettings);
  }, 100);
}

window.initMenu = initMenu;
