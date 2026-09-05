let currentSettings = null;

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

  // Render selections
  setupChipGroup("paper-types", currentSettings.default_paper_type, (val) => {
    currentSettings.default_paper_type = val;
    onSettingsChanged();
  });

  setupChipGroup("paper-patterns", currentSettings.default_paper_pattern, (val) => {
    currentSettings.default_paper_pattern = val;
    onSettingsChanged();
  });

  setupPenColors(currentSettings.default_pen_color, (val) => {
    currentSettings.default_pen_color = val;
    onSettingsChanged();
  });

  setupChipGroup("text-alignments", currentSettings.default_alignment, (val) => {
    currentSettings.default_alignment = val;
    onSettingsChanged();
  });

  // Color picker & Saturation
  const colorPicker = document.getElementById("color-picker");
  if (colorPicker) {
    colorPicker.value = currentSettings.default_color || "#fcf5e5";
    colorPicker.addEventListener("input", (e) => {
      currentSettings.default_color = e.target.value;
      onSettingsChanged();
    });
  }

  const satSlider = document.getElementById("saturation-slider");
  const satValue = document.getElementById("sat-value");
  if (satSlider) {
    satSlider.value = currentSettings.default_saturation || 80;
    if (satValue) satValue.textContent = `${satSlider.value}%`;
    satSlider.addEventListener("input", (e) => {
      currentSettings.default_saturation = parseInt(e.target.value, 10);
      if (satValue) satValue.textContent = `${currentSettings.default_saturation}%`;
      onSettingsChanged();
    });
  }

  // Shake Toggle
  const shakeToggle = document.getElementById("shake-toggle");
  if (shakeToggle) {
    shakeToggle.checked = currentSettings.shake_enabled !== false;
    shakeToggle.addEventListener("change", (e) => {
      currentSettings.shake_enabled = e.target.checked;
      onSettingsChanged();
    });
  }

  // Actions
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

  const btnClose = document.getElementById("btn-close-menu");
  if (btnClose) {
    btnClose.addEventListener("click", () => {
      sendAction("close_menu", {});
    });
  }

  // Attach Drawably
  if (window.drawably) {
    if (window.drawably.drawablyButton) {
      document.querySelectorAll(".drawably-button").forEach((btn) => {
        try {
          window.drawably.drawablyButton(btn, { variant: "outline" });
        } catch (err) {}
      });
    }
    if (window.drawably.drawablyCard) {
      const card = document.getElementById("menu-paper");
      if (card) {
        try {
          window.drawably.drawablyCard(card);
        } catch (err) {}
      }
    }
  }
}

function setupChipGroup(containerId, activeVal, onSelect) {
  const container = document.getElementById(containerId);
  if (!container) return;

  const chips = container.querySelectorAll(".paper-chip");
  chips.forEach((chip) => {
    if (chip.dataset.val === activeVal) {
      chip.classList.add("active");
    } else {
      chip.classList.remove("active");
    }

    chip.addEventListener("click", () => {
      chips.forEach((c) => c.classList.remove("active"));
      chip.classList.add("active");
      onSelect(chip.dataset.val);
    });
  });
}

function setupPenColors(activeVal, onSelect) {
  const container = document.getElementById("pen-colors");
  if (!container) return;

  const pens = container.querySelectorAll(".pen-chip");
  pens.forEach((pen) => {
    if (pen.dataset.val === activeVal) {
      pen.classList.add("active");
    } else {
      pen.classList.remove("active");
    }

    pen.addEventListener("click", () => {
      pens.forEach((p) => p.classList.remove("active"));
      pen.classList.add("active");
      onSelect(pen.dataset.val);
    });
  });
}

let changeTimeout = null;
function onSettingsChanged() {
  clearTimeout(changeTimeout);
  changeTimeout = setTimeout(() => {
    sendAction("save_settings", currentSettings);
  }, 150);
}

window.initMenu = initMenu;
