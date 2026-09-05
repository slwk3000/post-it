var drawably = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // dist/index.js
  var index_exports = {};
  __export(index_exports, {
    drawablyArrow: () => drawablyArrow,
    drawablyBadge: () => drawablyBadge,
    drawablyButton: () => drawablyButton,
    drawablyCard: () => drawablyCard,
    drawablyCheckbox: () => drawablyCheckbox,
    drawablyCircle: () => drawablyCircle,
    drawablyDivider: () => drawablyDivider,
    drawablyHighlight: () => drawablyHighlight,
    drawablyInput: () => drawablyInput,
    drawablyList: () => drawablyList,
    drawablyRadio: () => drawablyRadio,
    drawablySelect: () => drawablySelect,
    drawablyTextarea: () => drawablyTextarea,
    drawablyToggle: () => drawablyToggle,
    drawablyUnderline: () => drawablyUnderline,
    mulberry32: () => mulberry32,
    randomSeed: () => randomSeed,
    roughArrow: () => roughArrow,
    roughCheckmark: () => roughCheckmark,
    roughCircle: () => roughCircle,
    roughEllipse: () => roughEllipse,
    roughLine: () => roughLine,
    roughRoundedRect: () => roughRoundedRect,
    scribbleFill: () => scribbleFill,
    variants: () => variants
  });

  // dist/prng.js
  function mulberry32(seed) {
    let a = seed >>> 0;
    return () => {
      a = a + 1831565813 >>> 0;
      let t = a;
      t = Math.imul(t ^ t >>> 15, t | 1);
      t ^= t + Math.imul(t ^ t >>> 7, t | 61);
      return ((t ^ t >>> 14) >>> 0) / 4294967296;
    };
  }
  function randomSeed() {
    return Math.floor(Math.random() * 4294967296);
  }

  // dist/rough.js
  function sampleLine(x1, y1, x2, y2, step = 8) {
    const n = Math.max(2, Math.ceil(Math.hypot(x2 - x1, y2 - y1) / step));
    return Array.from({ length: n + 1 }, (_, i) => [
      x1 + (x2 - x1) * i / n,
      y1 + (y2 - y1) * i / n
    ]);
  }
  function arcPoints(cx, cy, r, a0, a1, n = 4) {
    return ellipsePoints(cx, cy, r, r, a0, a1, n);
  }
  function ellipsePoints(cx, cy, rx, ry, a0, a1, n) {
    return Array.from({ length: n + 1 }, (_, i) => {
      const a = a0 + (a1 - a0) * i / n;
      return [cx + rx * Math.cos(a), cy + ry * Math.sin(a)];
    });
  }
  function roundedRectPoints(x, y, w, h, r) {
    r = Math.min(r, w / 2, h / 2);
    return [
      ...sampleLine(x + r, y, x + w - r, y),
      ...arcPoints(x + w - r, y + r, r, -Math.PI / 2, 0),
      ...sampleLine(x + w, y + r, x + w, y + h - r),
      ...arcPoints(x + w - r, y + h - r, r, 0, Math.PI / 2),
      ...sampleLine(x + w - r, y + h, x + r, y + h),
      ...arcPoints(x + r, y + h - r, r, Math.PI / 2, Math.PI),
      ...sampleLine(x, y + h - r, x, y + r),
      ...arcPoints(x + r, y + r, r, Math.PI, Math.PI * 1.5)
    ];
  }
  function jitter(points, rand, amp) {
    return points.map(([x, y]) => [x + (rand() * 2 - 1) * amp, y + (rand() * 2 - 1) * amp]);
  }
  function toPath(points, close) {
    let d = `M${points[0][0].toFixed(2)} ${points[0][1].toFixed(2)}`;
    for (let i = 1; i < points.length - 1; i++) {
      const [cx, cy] = points[i];
      const mx = (cx + points[i + 1][0]) / 2;
      const my = (cy + points[i + 1][1]) / 2;
      d += `Q${cx.toFixed(2)} ${cy.toFixed(2)} ${mx.toFixed(2)} ${my.toFixed(2)}`;
    }
    const [lx, ly] = points[points.length - 1];
    d += `L${lx.toFixed(2)} ${ly.toFixed(2)}`;
    return close ? d + "Z" : d;
  }
  function boilPass(points, o) {
    if (!o.boil || o.boilSeed === void 0)
      return points;
    return jitter(points, mulberry32(o.boilSeed), o.boil);
  }
  function doubleStroke(points, o, close) {
    const rand = mulberry32(o.seed);
    const amp = 1.5 * o.roughness;
    return toPath(boilPass(jitter(points, rand, amp), o), close) + toPath(boilPass(jitter(points, rand, amp * 1.4), o), close);
  }
  function roughLine(x1, y1, x2, y2, o) {
    return doubleStroke(sampleLine(x1, y1, x2, y2), o, false);
  }
  function roughCircle(cx, cy, r, o) {
    return roughEllipse(cx, cy, r, r, o);
  }
  function roughEllipse(cx, cy, rx, ry, o) {
    const h = ((rx - ry) / (rx + ry)) ** 2;
    const perimeter = Math.PI * (rx + ry) * (1 + 3 * h / (10 + Math.sqrt(4 - 3 * h)));
    const n = Math.max(8, Math.ceil(perimeter / 8));
    return doubleStroke(ellipsePoints(cx, cy, rx, ry, 0, Math.PI * 2, n).slice(0, -1), o, true);
  }
  var ARROW_HEAD = 12;
  var ARROW_HEAD_ANGLE = Math.PI / 6;
  function roughArrow(x1, y1, x2, y2, o) {
    const a = Math.atan2(y2 - y1, x2 - x1);
    const wing = (da) => [
      x2 - ARROW_HEAD * Math.cos(a + da),
      y2 - ARROW_HEAD * Math.sin(a + da)
    ];
    const [lx, ly] = wing(ARROW_HEAD_ANGLE);
    const [rx, ry] = wing(-ARROW_HEAD_ANGLE);
    const rand = mulberry32(o.seed);
    const amp = 1.2 * o.roughness;
    const head = (px, py) => toPath(boilPass(jitter(sampleLine(x2, y2, px, py, 4), rand, amp), o), false);
    return roughLine(x1, y1, x2, y2, o) + head(lx, ly) + head(rx, ry);
  }
  function roughRoundedRect(x, y, w, h, r, o) {
    return doubleStroke(roundedRectPoints(x, y, w, h, r), o, true);
  }
  function roughCheckmark(x, y, w, h, o) {
    const rand = mulberry32(o.seed);
    const pts = [
      ...sampleLine(x, y + h * 0.6, x + w * 0.35, y + h, 4),
      // ponytail: duplicate vertex keeps the corner sharp under midpoint smoothing
      ...sampleLine(x + w * 0.35, y + h, x + w, y, 4)
    ];
    return toPath(boilPass(jitter(pts, rand, 1.2 * o.roughness), o), false);
  }
  function scribbleFill(x, y, w, h, o) {
    const rand = mulberry32(o.seed);
    const gap = 6;
    const pts = [];
    let flip = false;
    for (let t = gap; t < w + h; t += gap) {
      const a = [x + Math.max(0, t - h), y + Math.min(t, h)];
      const b = [x + Math.min(t, w), y + Math.max(0, t - w)];
      pts.push(...flip ? [b, a] : [a, b]);
      flip = !flip;
    }
    if (pts.length < 2)
      return "";
    return toPath(boilPass(jitter(pts, rand, 1.2 * o.roughness), o), false);
  }
  function variants(gen, o, n = 3) {
    return Array.from({ length: n }, (_, i) => gen({ ...o, boilSeed: o.seed + (i + 1) * 7919 }));
  }

  // dist/controls.js
  var SVG_NS = "http://www.w3.org/2000/svg";
  var INSET = 3;
  function applyTheme(el, opts) {
    if (opts.stroke)
      el.style.setProperty("--drawably-stroke", opts.stroke);
    if (opts.fill)
      el.style.setProperty("--drawably-fill", opts.fill);
    if (opts.paper)
      el.style.setProperty("--drawably-paper", opts.paper);
    if (opts.width !== void 0)
      el.style.setProperty("--drawably-width", String(opts.width));
  }
  function createSvg() {
    const svg = document.createElementNS(SVG_NS, "svg");
    svg.setAttribute("class", "drawably-svg");
    svg.setAttribute("aria-hidden", "true");
    return svg;
  }
  function paint(svg, layers, boxes, o) {
    const w = Math.max(...boxes.map((b) => b.x + b.w));
    const h = Math.max(...boxes.map((b) => b.y + b.h));
    svg.setAttribute("viewBox", `0 0 ${w} ${h}`);
    svg.textContent = "";
    for (const layer of layers) {
      boxes.forEach((box, k) => {
        const ds = variants((lo) => layer.gen(box.w, box.h, lo), { ...o, seed: o.seed + k }, o.boil ? 3 : 1);
        ds.forEach((d, i) => {
          const p = document.createElementNS(SVG_NS, "path");
          p.setAttribute("d", d);
          p.setAttribute("class", ds.length > 1 ? `drawably-boil ${layer.className}` : layer.className);
          p.dataset.i = String(i);
          if (layer.pathLength)
            p.setAttribute("pathLength", "1");
          if (box.x || box.y)
            p.setAttribute("transform", `translate(${box.x} ${box.y})`);
          svg.append(p);
        });
      });
    }
  }
  function elementBox(el) {
    return [{ x: 0, y: 0, w: el.offsetWidth || 120, h: el.offsetHeight || 36 }];
  }
  function lineBoxes(el, svg) {
    const rects = [...el.getClientRects()];
    if (rects.length < 2) {
      svg.style.cssText = "";
      return elementBox(el);
    }
    const [first] = rects;
    const left = Math.min(...rects.map((r) => r.left));
    const top = Math.min(...rects.map((r) => r.top));
    const right = Math.max(...rects.map((r) => r.right));
    const bottom = Math.max(...rects.map((r) => r.bottom));
    svg.style.cssText = `left:${left - first.left}px;top:${top - first.top}px;width:${right - left}px;height:${bottom - top}px`;
    return rects.map((r) => ({ x: r.left - left, y: r.top - top, w: r.width, h: r.height }));
  }
  function blockAncestor(el) {
    let p = el.parentElement;
    while (p && getComputedStyle(p).display === "inline")
      p = p.parentElement;
    return p;
  }
  function reducedMotion() {
    return typeof matchMedia !== "undefined" && matchMedia("(prefers-reduced-motion: reduce)").matches;
  }
  function attachChrome(el, layers, opts, interactive, inline = false) {
    if (!(el instanceof HTMLElement))
      throw new Error("drawably: expected an HTMLElement");
    el.classList.add("drawably-host");
    applyTheme(el, opts);
    const svg = createSvg();
    el.prepend(svg);
    const roughness = opts.roughness ?? 1;
    const boil = opts.boil ?? 0.3;
    let seed = opts.seed ?? randomSeed();
    const draw = () => paint(svg, layers, inline ? lineBoxes(el, svg) : elementBox(el), { seed, roughness, boil });
    draw();
    const ro = typeof ResizeObserver !== "undefined" ? new ResizeObserver(() => draw()) : null;
    ro?.observe(el);
    if (inline) {
      const block = blockAncestor(el);
      if (block)
        ro?.observe(block);
    }
    const resketch = (s) => {
      seed = s ?? randomSeed();
      draw();
    };
    const onPointer = () => resketch();
    if (interactive && !reducedMotion()) {
      el.addEventListener("pointerenter", onPointer);
      el.addEventListener("pointerdown", onPointer);
    }
    return {
      resketch,
      destroy() {
        ro?.disconnect();
        el.removeEventListener("pointerenter", onPointer);
        el.removeEventListener("pointerdown", onPointer);
        svg.remove();
        el.classList.remove("drawably-host");
      }
    };
  }
  var outlineRect = (r) => (w, h, o) => roughRoundedRect(INSET, INSET, w - 2 * INSET, h - 2 * INSET, r, o);
  var focusRect = (r) => (w, h, o) => roughRoundedRect(-1, -1, w + 2, h + 2, r, o);
  function drawablyButton(el, opts = {}) {
    const variant = opts.variant ?? "outline";
    const layers = [];
    if (variant === "solid")
      layers.push({ className: "drawably-blob", gen: outlineRect(8) });
    if (variant === "scribble")
      layers.push({
        className: "drawably-scribble",
        gen: (w, h, o) => scribbleFill(INSET + 2, INSET + 2, w - 2 * INSET - 4, h - 2 * INSET - 4, o)
      });
    layers.push({ className: "drawably-outline", gen: outlineRect(8) });
    layers.push({ className: "drawably-focus", gen: focusRect(10) });
    const sketch = attachChrome(el, layers, opts, true);
    el.classList.add("drawably-button", `drawably-button--${variant}`);
    if (opts.tone)
      el.classList.add(`drawably-button--${opts.tone}`);
    const setState = (state) => {
      if (state === "idle")
        delete el.dataset.state;
      else
        el.dataset.state = state;
    };
    if (opts.state)
      setState(opts.state);
    return {
      resketch: sketch.resketch,
      setState,
      destroy() {
        sketch.destroy();
        delete el.dataset.state;
      }
    };
  }
  function drawablyCard(el, opts = {}) {
    const sketch = attachChrome(el, [{ className: "drawably-outline", gen: outlineRect(10) }], opts, false);
    el.classList.add("drawably-card");
    return sketch;
  }
  function syncedControl(el, type, layers, opts, cls) {
    const input = el?.querySelector?.(`input[type="${type}"]`);
    if (!input)
      throw new Error(`drawably: ${cls} wrapper needs an <input type="${type}">`);
    const sync = () => {
      if (input.checked)
        el.dataset.checked = "";
      else
        delete el.dataset.checked;
    };
    sync();
    const target = type === "radio" ? document : input;
    target.addEventListener("change", sync);
    const sketch = attachChrome(el, layers, opts, true);
    el.classList.add(cls);
    return {
      resketch: sketch.resketch,
      destroy() {
        target.removeEventListener("change", sync);
        sketch.destroy();
        el.classList.remove(cls);
        delete el.dataset.checked;
      }
    };
  }
  function drawablyCheckbox(el, opts = {}) {
    return syncedControl(el, "checkbox", [
      { className: "drawably-outline", gen: outlineRect(5) },
      {
        className: "drawably-check",
        pathLength: true,
        gen: (w, h, o) => roughCheckmark(w * 0.24, h * 0.2, w * 0.52, h * 0.5, o)
      },
      { className: "drawably-focus", gen: focusRect(7) }
    ], opts, "drawably-checkbox");
  }
  function drawablyRadio(el, opts = {}) {
    return syncedControl(el, "radio", [
      {
        className: "drawably-outline",
        gen: (w, h, o) => roughCircle(w / 2, h / 2, Math.min(w, h) / 2 - INSET, o)
      },
      {
        className: "drawably-dot",
        gen: (w, h, o) => roughCircle(w / 2, h / 2, Math.min(w, h) * 0.18, o)
      },
      {
        className: "drawably-focus",
        gen: (w, h, o) => roughCircle(w / 2, h / 2, Math.min(w, h) / 2 + 1, o)
      }
    ], opts, "drawably-radio");
  }
  function drawablyToggle(el, opts = {}) {
    return syncedControl(el, "checkbox", [
      { className: "drawably-outline", gen: (w, h, o) => outlineRect((h - 2 * INSET) / 2)(w, h, o) },
      {
        className: "drawably-blob drawably-knob",
        gen: (_w, h, o) => roughCircle(h / 2, h / 2, h / 2 - INSET - 3, o)
      },
      { className: "drawably-focus", gen: focusRect(12) }
    ], opts, "drawably-toggle");
  }
  function drawablyDivider(el, opts = {}) {
    const sketch = attachChrome(el, [{ className: "drawably-outline", gen: (w, h, o) => roughLine(INSET, h / 2, w - INSET, h / 2, o) }], opts, false);
    el.classList.add("drawably-divider");
    return sketch;
  }
  function fieldBox(el, field, cls, extra, opts) {
    if (!el?.querySelector?.(field))
      throw new Error(`drawably: ${cls} wrapper needs a <${field}>`);
    return decoration(el, cls, [{ className: "drawably-outline", gen: outlineRect(6) }, ...extra, { className: "drawably-focus", gen: focusRect(8) }], opts, false);
  }
  function drawablyInput(el, opts = {}) {
    return fieldBox(el, "input", "drawably-inputbox", [], opts);
  }
  function drawablyTextarea(el, opts = {}) {
    return fieldBox(el, "textarea", "drawably-textarea", [], opts);
  }
  var CHEVRON_W = 12;
  var CHEVRON_H = 6;
  var CHEVRON_RIGHT = 12;
  var CHEVRON_ROUGHNESS = 0.4;
  function reserveWidestOption(select) {
    const ctx = document.createElement("canvas").getContext("2d");
    if (!ctx)
      return;
    const cs = getComputedStyle(select);
    ctx.font = `${cs.fontWeight} ${cs.fontSize} ${cs.fontFamily}`;
    const widest = Math.max(0, ...[...select.options].map((o) => ctx.measureText(o.text).width));
    const pad = (parseFloat(cs.paddingLeft) || 0) + (parseFloat(cs.paddingRight) || 0);
    select.style.minWidth = `${Math.ceil(widest + pad)}px`;
  }
  var PICKER_RADIUS = 6;
  function pickerFrame(select, o) {
    const frame = createSvg();
    frame.classList.add("drawably-picker");
    select.append(frame);
    const draw = () => {
      const w = frame.clientWidth;
      const h = frame.clientHeight;
      if (w && h)
        paint(frame, [{ className: "drawably-outline", gen: outlineRect(PICKER_RADIUS) }], [{ x: 0, y: 0, w, h }], o());
    };
    const ro = typeof ResizeObserver !== "undefined" ? new ResizeObserver(draw) : null;
    ro?.observe(frame);
    return {
      draw,
      destroy() {
        ro?.disconnect();
        frame.remove();
      }
    };
  }
  var CHECK_BOX = 14;
  var CHECK_INSET = 2;
  function checkMask(o) {
    const side = CHECK_BOX - CHECK_INSET * 2;
    const d = roughCheckmark(CHECK_INSET, CHECK_INSET, side, side, o);
    const svg = `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 ${CHECK_BOX} ${CHECK_BOX}'><path d='${d}' fill='none' stroke='#000' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'/></svg>`;
    return `url("data:image/svg+xml,${encodeURIComponent(svg)}")`;
  }
  function drawablySelect(el, opts = {}) {
    const chevron = {
      className: "drawably-chevron",
      gen: (w, h, o) => {
        const x = w - CHEVRON_RIGHT - CHEVRON_W;
        const y = h / 2 - CHEVRON_H / 2;
        const co = { ...o, roughness: o.roughness * CHEVRON_ROUGHNESS };
        return roughLine(x, y, x + CHEVRON_W / 2, y + CHEVRON_H, co) + roughLine(x + CHEVRON_W / 2, y + CHEVRON_H, x + CHEVRON_W, y, { ...co, seed: o.seed + 1 });
      }
    };
    let seed = opts.seed ?? randomSeed();
    const sketch = fieldBox(el, "select", "drawably-select", [chevron], { ...opts, seed });
    const select = el.querySelector("select");
    const roughness = opts.roughness ?? 1;
    const boil = opts.boil ?? 0.3;
    reserveWidestOption(select);
    const frame = pickerFrame(select, () => ({ seed, roughness, boil }));
    const mark = () => select.style.setProperty("--drawably-check", checkMask({ seed, roughness }));
    mark();
    return {
      resketch(s) {
        seed = s ?? randomSeed();
        sketch.resketch(seed);
        frame.draw();
        mark();
      },
      destroy() {
        sketch.destroy();
        frame.destroy();
        select.style.removeProperty("min-width");
        select.style.removeProperty("--drawably-check");
      }
    };
  }
  function drawablyBadge(el, opts = {}) {
    const variant = opts.variant ?? "outline";
    const layers = [];
    if (variant === "scribble")
      layers.push({
        className: "drawably-scribble",
        gen: (w, h, o) => scribbleFill(INSET + 1, INSET + 1, w - 2 * INSET - 2, h - 2 * INSET - 2, o)
      });
    layers.push({ className: "drawably-outline", gen: outlineRect(2) });
    const sketch = decoration(el, "drawably-badge", layers, opts, false);
    el.classList.add(`drawably-badge--${variant}`);
    return {
      resketch: sketch.resketch,
      destroy() {
        sketch.destroy();
        el.classList.remove(`drawably-badge--${variant}`);
      }
    };
  }
  var MARKER_LEFT = -18;
  var MARKER_W = 10;
  var MARKER_LINE = 22;
  function drawablyList(el, opts = {}) {
    if (!(el instanceof HTMLElement))
      throw new Error("drawably: expected an HTMLElement");
    const marker = opts.marker ?? "dash";
    const seed = opts.seed ?? randomSeed();
    const items = [...el.querySelectorAll(":scope > li")];
    const sketches = items.map((li, i) => {
      const line = () => parseFloat(getComputedStyle(li).lineHeight) || MARKER_LINE;
      const layer = marker === "check" ? {
        className: "drawably-marker",
        gen: (_w, _h, o) => roughCheckmark(MARKER_LEFT, line() / 2 - MARKER_W / 2, MARKER_W, MARKER_W, o)
      } : {
        className: "drawably-marker",
        gen: (_w, _h, o) => roughLine(MARKER_LEFT, line() / 2, MARKER_LEFT + MARKER_W, line() / 2, o)
      };
      return attachChrome(li, [layer], { ...opts, seed: seed + i }, false);
    });
    el.classList.add("drawably-list");
    return {
      resketch(s) {
        const base = s ?? randomSeed();
        sketches.forEach((sk, i) => sk.resketch(base + i));
      },
      destroy() {
        for (const sk of sketches)
          sk.destroy();
        el.classList.remove("drawably-list");
      }
    };
  }
  function decoration(el, cls, layers, opts, interactive) {
    const sketch = attachChrome(el, layers, opts, interactive, true);
    el.classList.add(cls);
    return {
      resketch: sketch.resketch,
      destroy() {
        sketch.destroy();
        el.classList.remove(cls);
      }
    };
  }
  var UNDERLINE_GAP = 2;
  var CIRCLE_PAD_X = 1.15;
  var CIRCLE_PAD_Y = 1.4;
  var CIRCLE_PAD = 4;
  function drawablyUnderline(el, opts = {}) {
    return decoration(el, "drawably-underline", [{ className: "drawably-outline", gen: (w, h, o) => roughLine(0, h + UNDERLINE_GAP, w, h + UNDERLINE_GAP, o) }], opts, true);
  }
  function drawablyHighlight(el, opts = {}) {
    return decoration(el, "drawably-highlight", [{ className: "drawably-wash", gen: (w, h, o) => scribbleFill(0, 0, w, h, o) }], opts, false);
  }
  function drawablyCircle(el, opts = {}) {
    return decoration(el, "drawably-circle", [
      {
        className: "drawably-outline",
        gen: (w, h, o) => roughEllipse(w / 2, h / 2, w / 2 * CIRCLE_PAD_X + CIRCLE_PAD, h / 2 * CIRCLE_PAD_Y + CIRCLE_PAD, o)
      }
    ], opts, true);
  }
  var ARROW_GAP = 6;
  function drawablyArrow(from, to, opts = {}) {
    if (!(from instanceof HTMLElement) || !(to instanceof HTMLElement))
      throw new Error("drawably: arrow needs two anchor elements");
    const svg = createSvg();
    svg.classList.add("drawably-arrow");
    applyTheme(svg, opts);
    document.body.append(svg);
    const roughness = opts.roughness ?? 1;
    const boil = opts.boil ?? 0.3;
    let seed = opts.seed ?? randomSeed();
    function draw() {
      const a = from.getBoundingClientRect();
      const b = to.getBoundingClientRect();
      const left = Math.min(a.left, b.left);
      const top = Math.min(a.top, b.top);
      const w = Math.max(a.right, b.right) - left;
      const h = Math.max(a.bottom, b.bottom) - top;
      svg.style.left = `${left + scrollX}px`;
      svg.style.top = `${top + scrollY}px`;
      svg.style.width = `${w}px`;
      svg.style.height = `${h}px`;
      const ax = a.left - left + a.width / 2;
      const ay = a.top - top + a.height / 2;
      const bx = b.left - left + b.width / 2;
      const by = b.top - top + b.height / 2;
      const dx = bx - ax;
      const dy = by - ay;
      const len = Math.hypot(dx, dy) || 1;
      const ux = dx / len;
      const uy = dy / len;
      const exit = (r) => Math.min(r.width / 2 / Math.abs(ux) || 0, r.height / 2 / Math.abs(uy) || 0);
      const t0 = Math.min(exit(a) + ARROW_GAP, len / 2);
      const t1 = Math.min(exit(b) + ARROW_GAP, len / 2);
      const x1 = ax + ux * t0;
      const y1 = ay + uy * t0;
      const x2 = bx - ux * t1;
      const y2 = by - uy * t1;
      paint(svg, [{ className: "drawably-outline", gen: (_w, _h, o) => roughArrow(x1, y1, x2, y2, o) }], [{ x: 0, y: 0, w, h }], { seed, roughness, boil });
    }
    draw();
    const ro = typeof ResizeObserver !== "undefined" ? new ResizeObserver(() => draw()) : null;
    ro?.observe(from);
    ro?.observe(to);
    addEventListener("resize", draw);
    return {
      resketch(s) {
        seed = s ?? randomSeed();
        draw();
      },
      destroy() {
        ro?.disconnect();
        removeEventListener("resize", draw);
        svg.remove();
      }
    };
  }
  return __toCommonJS(index_exports);
})();
