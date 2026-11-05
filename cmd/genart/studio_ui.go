package main

const studioWorkbenchHTML = `<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>GenerativeArt Studio — 可视化调参工作台</title>
  <style>
    :root {
      --bg: #070a12;
      --bg-panel: #0d121f;
      --bg-elev: #131b2e;
      --bg-input: #090d18;
      --border: rgba(255, 255, 255, 0.09);
      --border-strong: rgba(56, 189, 248, 0.42);
      --text: #f1f5f9;
      --muted: #94a3b8;
      --accent: #38bdf8;
      --accent-2: #a855f7;
      --accent-3: #f43f5e;
      --ok: #10b981;
      --warn: #f59e0b;
      --shadow: 0 16px 40px -12px rgba(0, 0, 0, 0.7);
    }
    [data-theme="light"] {
      --bg: #f3efe6;
      --bg-panel: #fbf9f5;
      --bg-elev: #efeae0;
      --bg-input: #ffffff;
      --border: rgba(15, 23, 42, 0.12);
      --border-strong: rgba(2, 132, 199, 0.45);
      --text: #0f172a;
      --muted: #64748b;
      --accent: #0284c7;
      --accent-2: #7c3aed;
      --accent-3: #e11d48;
      --ok: #059669;
      --warn: #d97706;
      --shadow: 0 12px 32px -10px rgba(15, 23, 42, 0.15);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    html, body {
      width: 100%;
      height: 100vh;
      overflow: hidden;
      font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "PingFang SC", "Segoe UI", Roboto, sans-serif;
      background: var(--bg);
      color: var(--text);
      font-size: 13px;
      line-height: 1.45;
    }
    a { color: inherit; text-decoration: none; }

    /* Top Header */
    .topbar {
      height: 54px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 1.15rem;
      background: var(--bg-panel);
      border-bottom: 1px solid var(--border);
      gap: 0.75rem;
      z-index: 20;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 0.65rem;
      font-weight: 700;
      font-size: 1.02rem;
      letter-spacing: -0.01em;
      white-space: nowrap;
    }
    .brand-dot {
      width: 11px;
      height: 11px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--accent), var(--accent-3));
      box-shadow: 0 0 12px var(--accent);
    }
    .brand-badge {
      font-size: 0.72rem;
      font-weight: 600;
      padding: 0.16rem 0.55rem;
      border-radius: 999px;
      background: rgba(56, 189, 248, 0.14);
      border: 1px solid rgba(56, 189, 248, 0.32);
      color: var(--accent);
    }
    .topbar-right {
      display: flex;
      align-items: center;
      gap: 0.55rem;
      flex-wrap: wrap;
    }

    /* Buttons & Inputs */
    .btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 0.38rem;
      padding: 0.42rem 0.82rem;
      border-radius: 8px;
      border: 1px solid var(--border);
      background: var(--bg-elev);
      color: var(--text);
      font-size: 0.8rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.15s ease;
      white-space: nowrap;
      user-select: none;
    }
    .btn:hover:not(:disabled) {
      border-color: var(--border-strong);
      transform: translateY(-1px);
    }
    .btn:active:not(:disabled) { transform: translateY(0); }
    .btn:disabled { opacity: 0.45; cursor: not-allowed; }
    .btn-primary {
      background: linear-gradient(135deg, var(--accent), var(--accent-2));
      color: #ffffff;
      border: none;
      box-shadow: 0 4px 14px rgba(56, 189, 248, 0.28);
    }
    .btn-dice {
      background: linear-gradient(135deg, #f43f5e, #fb923c);
      color: #ffffff;
      border: none;
      box-shadow: 0 4px 14px rgba(244, 63, 94, 0.28);
    }
    .btn-success {
      background: linear-gradient(135deg, #10b981, #0284c7);
      color: #ffffff;
      border: none;
      box-shadow: 0 4px 14px rgba(16, 185, 129, 0.25);
    }
    .btn-sm {
      padding: 0.28rem 0.58rem;
      font-size: 0.74rem;
      border-radius: 6px;
    }

    .select-input, .text-input, .num-input {
      background: var(--bg-input);
      color: var(--text);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 0.42rem 0.65rem;
      font-size: 0.82rem;
      outline: none;
      transition: border-color 0.15s ease;
    }
    .select-input:focus, .text-input:focus, .num-input:focus {
      border-color: var(--accent);
    }

    /* Main 3-Column Studio Grid */
    .studio-layout {
      display: grid;
      grid-template-columns: 360px 1fr 330px;
      height: calc(100vh - 54px);
      width: 100%;
      overflow: hidden;
    }
    @media (max-width: 1180px) {
      .studio-layout {
        grid-template-columns: 330px 1fr 290px;
      }
    }
    @media (max-width: 920px) {
      html, body { height: auto; overflow: auto; }
      .studio-layout {
        grid-template-columns: 1fr;
        height: auto;
      }
    }

    /* Sidebars */
    .sidebar-left, .sidebar-right {
      background: var(--bg-panel);
      overflow-y: auto;
      padding: 0.9rem;
      display: flex;
      flex-direction: column;
      gap: 0.85rem;
    }
    .sidebar-left { border-right: 1px solid var(--border); }
    .sidebar-right { border-left: 1px solid var(--border); }

    .panel-card {
      background: var(--bg-elev);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 0.8rem;
      display: flex;
      flex-direction: column;
      gap: 0.65rem;
    }
    .panel-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.5rem;
    }
    .panel-title {
      font-size: 0.78rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      color: var(--accent);
      display: flex;
      align-items: center;
      gap: 0.38rem;
    }

    /* Category Filter Pills */
    .cat-pills {
      display: flex;
      flex-wrap: wrap;
      gap: 0.32rem;
    }
    .cat-pill {
      padding: 0.22rem 0.56rem;
      border-radius: 999px;
      border: 1px solid var(--border);
      background: var(--bg-input);
      color: var(--muted);
      font-size: 0.72rem;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.14s ease;
    }
    .cat-pill:hover { color: var(--text); border-color: var(--border-strong); }
    .cat-pill.active {
      background: linear-gradient(135deg, var(--accent), var(--accent-2));
      color: #fff;
      border-color: transparent;
    }

    /* Engine Selector Row */
    .art-select-row {
      display: flex;
      gap: 0.35rem;
      align-items: center;
    }
    .art-select-row select {
      flex: 1;
      min-width: 0;
      font-weight: 600;
    }

    /* Palette Swatches */
    .swatch-bar {
      display: flex;
      align-items: center;
      gap: 5px;
      flex-wrap: wrap;
      padding-top: 0.15rem;
    }
    .swatch-dot {
      width: 18px;
      height: 18px;
      border-radius: 50%;
      border: 1px solid rgba(255, 255, 255, 0.28);
      box-shadow: 0 2px 5px rgba(0, 0, 0, 0.35);
    }
    .palette-quick-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 0.32rem;
      margin-top: 0.2rem;
    }
    .pal-chip {
      display: flex;
      align-items: center;
      gap: 0.3rem;
      padding: 0.24rem 0.42rem;
      border-radius: 6px;
      border: 1px solid var(--border);
      background: var(--bg-input);
      color: var(--muted);
      font-size: 0.7rem;
      font-weight: 600;
      cursor: pointer;
      overflow: hidden;
    }
    .pal-chip:hover { color: var(--text); border-color: var(--border-strong); }
    .pal-chip.active {
      border-color: var(--accent);
      color: var(--text);
      background: rgba(56, 189, 248, 0.13);
    }
    .pal-mini-bar {
      display: flex;
      width: 24px;
      height: 8px;
      border-radius: 3px;
      overflow: hidden;
      flex-shrink: 0;
    }
    .pal-mini-bar span { flex: 1; height: 100%; }

    /* Parameter Sliders */
    .param-list {
      display: flex;
      flex-direction: column;
      gap: 0.65rem;
    }
    .param-item {
      background: var(--bg-input);
      border: 1px solid var(--border);
      border-radius: 9px;
      padding: 0.52rem 0.65rem;
      display: flex;
      flex-direction: column;
      gap: 0.36rem;
    }
    .param-top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.4rem;
    }
    .param-name {
      font-weight: 600;
      font-size: 0.79rem;
      display: flex;
      align-items: center;
      gap: 0.35rem;
    }
    .param-type-badge {
      font-family: "SF Mono", Menlo, monospace;
      font-size: 0.65rem;
      padding: 0.08rem 0.36rem;
      border-radius: 4px;
      background: rgba(168, 85, 247, 0.15);
      color: #c084fc;
      font-weight: 600;
    }
    .param-def-btn {
      font-size: 0.68rem;
      color: var(--muted);
      cursor: pointer;
      background: none;
      border: none;
      text-decoration: underline dotted;
    }
    .param-def-btn:hover { color: var(--accent); }
    .param-control-row {
      display: flex;
      align-items: center;
      gap: 0.55rem;
    }
    .range-slider {
      flex: 1;
      -webkit-appearance: none;
      appearance: none;
      height: 6px;
      border-radius: 999px;
      background: rgba(148, 163, 184, 0.25);
      outline: none;
      cursor: pointer;
    }
    .range-slider::-webkit-slider-thumb {
      -webkit-appearance: none;
      appearance: none;
      width: 15px;
      height: 15px;
      border-radius: 50%;
      background: var(--accent);
      border: 2px solid #ffffff;
      box-shadow: 0 2px 6px rgba(0, 0, 0, 0.45);
      cursor: grab;
    }
    .range-slider::-webkit-slider-thumb:active {
      cursor: grabbing;
      transform: scale(1.12);
    }
    .param-num-box {
      width: 74px;
      text-align: right;
      font-family: "SF Mono", Menlo, monospace;
      font-size: 0.76rem;
      padding: 0.24rem 0.42rem;
    }
    .preset-chips {
      display: flex;
      flex-wrap: wrap;
      gap: 0.28rem;
      margin-top: 0.2rem;
    }
    .preset-chip {
      font-family: "SF Mono", Menlo, monospace;
      font-size: 0.68rem;
      padding: 0.14rem 0.42rem;
      border-radius: 5px;
      border: 1px solid var(--border);
      background: var(--bg-elev);
      color: var(--text);
      cursor: pointer;
    }
    .preset-chip:hover { border-color: var(--accent); color: var(--accent); }

    /* Center Stage */
    .center-stage {
      display: flex;
      flex-direction: column;
      height: 100%;
      overflow: hidden;
      background: radial-gradient(circle at 50% 42%, rgba(56, 189, 248, 0.06), transparent 70%);
    }
    .stage-toolbar {
      padding: 0.65rem 1rem;
      border-bottom: 1px solid var(--border);
      background: var(--bg-panel);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.6rem;
      flex-wrap: wrap;
    }
    .stage-toolbar-group {
      display: flex;
      align-items: center;
      gap: 0.45rem;
      flex-wrap: wrap;
    }
    .canvas-viewport {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 1.1rem;
      overflow: auto;
      position: relative;
    }
    .canvas-frame {
      position: relative;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      border-radius: 14px;
      padding: 12px;
      background: rgba(13, 18, 31, 0.85);
      border: 1px solid var(--border-strong);
      box-shadow: var(--shadow);
      transition: all 0.2s ease;
    }
    .canvas-frame.mat-light {
      background: #f5f2eb;
      padding: 22px;
      border-color: rgba(0, 0, 0, 0.2);
    }
    #renderCanvas {
      display: block;
      max-width: min(68vh, 100%);
      max-height: 68vh;
      width: auto;
      height: auto;
      border-radius: 8px;
      background: #05070c;
      box-shadow: 0 8px 28px rgba(0, 0, 0, 0.55);
    }
    .canvas-frame.zoom-actual #renderCanvas {
      max-width: none;
      max-height: none;
    }
    .loading-overlay {
      position: absolute;
      inset: 12px;
      border-radius: 8px;
      background: rgba(7, 10, 18, 0.62);
      backdrop-filter: blur(3px);
      display: none;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 0.55rem;
      color: #ffffff;
      font-weight: 600;
      z-index: 5;
    }
    .loading-overlay.active { display: flex; }
    .spinner {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      border: 3px solid rgba(56, 189, 248, 0.25);
      border-top-color: var(--accent);
      animation: spin 0.75s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* History Filmstrip */
    .history-strip-wrap {
      height: 98px;
      border-top: 1px solid var(--border);
      background: var(--bg-panel);
      padding: 0.45rem 0.9rem;
      display: flex;
      flex-direction: column;
      gap: 0.3rem;
    }
    .history-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 0.72rem;
      color: var(--muted);
      font-weight: 600;
    }
    .history-strip {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      overflow-x: auto;
      padding-bottom: 2px;
    }
    .hist-thumb {
      width: 56px;
      height: 56px;
      border-radius: 7px;
      border: 2px solid var(--border);
      overflow: hidden;
      cursor: pointer;
      flex-shrink: 0;
      position: relative;
      background: #05070c;
      transition: transform 0.14s ease, border-color 0.14s ease;
    }
    .hist-thumb:hover {
      transform: translateY(-2px);
      border-color: var(--accent);
    }
    .hist-thumb.active {
      border-color: var(--accent);
      box-shadow: 0 0 10px rgba(56, 189, 248, 0.4);
    }
    .hist-thumb img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    /* Right Sidebar: Aesthetic Score & Export */
    .score-hero {
      display: flex;
      align-items: center;
      justify-content: space-between;
      background: var(--bg-input);
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 0.65rem 0.8rem;
    }
    .score-num {
      font-size: 1.75rem;
      font-weight: 800;
      letter-spacing: -0.03em;
      background: linear-gradient(135deg, var(--accent), #10b981);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    .score-grade {
      font-size: 0.78rem;
      font-weight: 700;
      padding: 0.22rem 0.6rem;
      border-radius: 999px;
      background: rgba(16, 185, 129, 0.16);
      color: var(--ok);
      border: 1px solid rgba(16, 185, 129, 0.35);
    }
    .metric-bars {
      display: flex;
      flex-direction: column;
      gap: 0.45rem;
    }
    .metric-row {
      display: flex;
      flex-direction: column;
      gap: 0.18rem;
    }
    .metric-label-row {
      display: flex;
      justify-content: space-between;
      font-size: 0.73rem;
      color: var(--muted);
    }
    .metric-bar-bg {
      height: 6px;
      border-radius: 999px;
      background: rgba(148, 163, 184, 0.18);
      overflow: hidden;
    }
    .metric-bar-fill {
      height: 100%;
      border-radius: 999px;
      background: linear-gradient(90deg, var(--accent), var(--accent-2));
      transition: width 0.25s ease;
    }

    .code-box {
      background: var(--bg-input);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 0.55rem 0.65rem;
      font-family: "SF Mono", Menlo, Consolas, monospace;
      font-size: 0.72rem;
      color: var(--text);
      white-space: pre-wrap;
      word-break: break-word;
      line-height: 1.45;
      max-height: 120px;
      overflow-y: auto;
    }

    .dropzone {
      border: 1.5px dashed var(--border-strong);
      border-radius: 10px;
      padding: 0.75rem;
      text-align: center;
      color: var(--muted);
      font-size: 0.75rem;
      cursor: pointer;
      background: rgba(56, 189, 248, 0.04);
      transition: all 0.15s ease;
    }
    .dropzone:hover, .dropzone.dragover {
      background: rgba(56, 189, 248, 0.12);
      color: var(--text);
      border-color: var(--accent);
    }

    .toggle-label {
      display: inline-flex;
      align-items: center;
      gap: 0.4rem;
      font-size: 0.78rem;
      font-weight: 600;
      cursor: pointer;
      user-select: none;
    }
    .toggle-label input {
      accent-color: var(--accent);
      width: 15px;
      height: 15px;
      cursor: pointer;
    }
    .meta-tag {
      font-family: "SF Mono", Menlo, monospace;
      font-size: 0.72rem;
      padding: 0.2rem 0.5rem;
      border-radius: 6px;
      background: var(--bg-input);
      border: 1px solid var(--border);
      color: var(--muted);
    }
  </style>
</head>
<body>

  <!-- Top Navigation Bar -->
  <header class="topbar">
    <div class="brand">
      <span class="brand-dot"></span>
      <span>GenerativeArt Studio</span>
      <span class="brand-badge" id="engineCountBadge">148 Arts · 可视化调参工作台</span>
    </div>

    <div class="topbar-right">
      <label class="toggle-label" title="拖动滑块或切换调色板时自动实时渲染">
        <input type="checkbox" id="autoRenderToggle" checked />
        <span>⚡ 实时自动渲染</span>
      </label>

      <select id="sizeSelect" class="select-input" onchange="onResolutionChange()" title="渲染画布分辨率">
        <option value="360">360 × 360 (极速预览)</option>
        <option value="480" selected>480 × 480 (标准实时)</option>
        <option value="600">600 × 600 (高清原生)</option>
        <option value="800">800 × 800 (精修大图)</option>
        <option value="1200">1200 × 1200 (2K 典藏)</option>
      </select>

      <button type="button" class="btn" onclick="toggleTheme()" id="themeBtn">☀️ 浅色</button>
      <a href="/" class="btn">🖼️ 返回展览画廊</a>
    </div>
  </header>

  <!-- Main 3-Column Studio Workbench -->
  <div class="studio-layout">

    <!-- LEFT SIDEBAR: Art Engine, Palette, Seed, Dynamic Parameter Sliders & Post-FX -->
    <aside class="sidebar-left">

      <!-- 1. Art Selector Card -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">🎨 1. 选择艺术算法 (Art Engine)</span>
          <span class="meta-tag" id="activeCatTag">geom</span>
        </div>

        <div class="cat-pills" id="catPills">
          <button type="button" class="cat-pill active" data-cat="all" onclick="setCategory('all')">全部</button>
          <button type="button" class="cat-pill" data-cat="geom" onclick="setCategory('geom')">几何/分形</button>
          <button type="button" class="cat-pill" data-cat="chaos" onclick="setCategory('chaos')">混沌/吸引子</button>
          <button type="button" class="cat-pill" data-cat="flow" onclick="setCategory('flow')">流场/物理</button>
          <button type="button" class="cat-pill" data-cat="media" onclick="setCategory('media')">水墨/材质</button>
          <button type="button" class="cat-pill" data-cat="bio" onclick="setCategory('bio')">生物/形态</button>
        </div>

        <input type="text" id="artSearchInput" class="text-input" placeholder="🔍 搜索算法名称或 ID (例如 escher, moebius, nebula...)" oninput="filterArtSelect()" />

        <div class="art-select-row">
          <button type="button" class="btn btn-sm" onclick="stepArt(-1)" title="上一个算法">‹</button>
          <select id="artSelect" class="select-input" onchange="onArtSelectChange()"></select>
          <button type="button" class="btn btn-sm" onclick="stepArt(1)" title="下一个算法">›</button>
        </div>

        <div style="display:grid; grid-template-columns: 1fr 1fr; gap: 0.45rem; margin-top: 0.15rem;">
          <button type="button" class="btn btn-dice" onclick="randomizeCurrentArt()" title="随机生成当前算法的参数、调色板与种子">
            🎲 随机参数灵感
          </button>
          <button type="button" class="btn" onclick="surpriseRandomArt()" title="随机抽取一个艺术算法并随机生成">
            🎰 随机切换算法
          </button>
        </div>
      </section>

      <!-- 2. Palette & Seed Card -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">🌈 2. 调色板与随机种子 (Palette & Seed)</span>
          <button type="button" class="param-def-btn" onclick="resetPaletteToDefault()">恢复默认色板</button>
        </div>

        <div style="display:flex; flex-direction:column; gap:0.35rem;">
          <label style="font-size:0.74rem; color:var(--muted); font-weight:600;">色彩方案 (Color Schema Dropdown):</label>
          <select id="paletteSelect" class="select-input" onchange="onPaletteChange()"></select>
          <div class="swatch-bar" id="activePaletteSwatches"></div>
          <div class="palette-quick-grid" id="paletteQuickGrid"></div>
        </div>

        <div style="display:flex; align-items:center; justify-content:space-between; gap:0.45rem; margin-top:0.25rem; padding-top:0.5rem; border-top:1px solid var(--border);">
          <label style="font-size:0.74rem; color:var(--muted); font-weight:600; white-space:nowrap;">🌱 种子 (Seed):</label>
          <div style="display:flex; align-items:center; gap:0.3rem; flex:1; justify-content:flex-end;">
            <button type="button" class="btn btn-sm" onclick="stepSeed(-1)">-1</button>
            <input type="number" id="seedInput" class="num-input" style="width:108px; text-align:center; font-family:monospace;" value="20261105" onchange="scheduleRender()" />
            <button type="button" class="btn btn-sm" onclick="stepSeed(1)">+1</button>
            <button type="button" class="btn btn-sm" onclick="rollNewSeed()" title="随机新种子">🎲</button>
          </div>
        </div>
      </section>

      <!-- 3. Dynamic Engine Parameter Sliders -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">🎛️ 3. 算法参数滑块 (Engine Sliders)</span>
          <button type="button" class="param-def-btn" onclick="resetEngineParams()">重置全部默认</button>
        </div>
        <div id="paramSlidersContainer" class="param-list"></div>
      </section>

      <!-- 4. HDR Optical & Material Post-Processing Sliders -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">✨ 4. HDR 光学与材质后处理 (Post-FX)</span>
          <button type="button" class="param-def-btn" onclick="resetPostFX()">清空特效</button>
        </div>
        <div class="param-list" id="postFxContainer"></div>
      </section>

    </aside>

    <!-- CENTER STAGE: Live Canvas + Action & Download Toolbar + Session History Filmstrip -->
    <main class="center-stage">

      <!-- Top Action & Download Bar -->
      <div class="stage-toolbar">
        <div class="stage-toolbar-group">
          <button type="button" class="btn btn-primary" onclick="triggerRender(false)">
            ✨ 在线生成 (Render Now)
          </button>
          <button type="button" class="btn btn-dice" onclick="randomizeCurrentArt()">
            🎲 随机生成 (Randomize)
          </button>
          <button type="button" class="btn" onclick="resetAllAndRender()">
            🔄 恢复默认
          </button>
        </div>

        <div class="stage-toolbar-group">
          <span class="meta-tag" id="renderStatusBadge">就绪 · Ready</span>
          <button type="button" class="btn" onclick="toggleCanvasMat()" id="matBtn" title="切换画框衬底">🖼️ 画框</button>
          <button type="button" class="btn" onclick="toggleCanvasZoom()" id="zoomBtn" title="切换适应窗口 / 1:1 原始像素">🔍 适应窗口</button>
          <button type="button" class="btn btn-success" onclick="downloadCurrentPNG()">
            💾 下载 PNG
          </button>
          <button type="button" class="btn" onclick="downloadHighResPNG(1200)" title="使用当前参数即时渲染并下载 1200×1200 高清 PNG">
            ⬇️ 下载 1200px 高清图
          </button>
        </div>
      </div>

      <!-- Interactive Canvas Stage -->
      <div class="canvas-viewport" id="canvasViewport">
        <div class="canvas-frame" id="canvasFrame">
          <canvas id="renderCanvas" width="480" height="480"></canvas>
          <div class="loading-overlay" id="loadingOverlay">
            <div class="spinner"></div>
            <span id="loadingText">正在实时渲染 64-Bit HDR 画布...</span>
          </div>
        </div>
      </div>

      <!-- Bottom Session History Filmstrip -->
      <div class="history-strip-wrap">
        <div class="history-header">
          <span>🎞️ 本次调参生成历史 (点击缩略图可随时回溯当时的滑块参数与画布)</span>
          <span id="historyCountLabel">0 幅记录</span>
        </div>
        <div class="history-strip" id="historyStrip"></div>
      </div>

    </main>

    <!-- RIGHT SIDEBAR: Aesthetic Metrics, CLI/Go Code & PNG Recipe Inspector -->
    <aside class="sidebar-right">

      <!-- 1. Oklab Perceptual Aesthetic Score -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">🏆 Oklab 感知美学评分</span>
          <span class="meta-tag" id="renderTimeTag">-- ms</span>
        </div>

        <div class="score-hero">
          <div>
            <div style="font-size:0.72rem; color:var(--muted); font-weight:600;">综合美学得分 (Total Score)</div>
            <div class="score-num" id="scoreTotalVal">--</div>
          </div>
          <span class="score-grade" id="scoreGradeBadge">待评估</span>
        </div>

        <div class="metric-bars">
          <div class="metric-row">
            <div class="metric-label-row">
              <span>动态范围 (Dynamic Range)</span>
              <span id="mDynVal">0.00</span>
            </div>
            <div class="metric-bar-bg"><div class="metric-bar-fill" id="mDynBar" style="width:0%"></div></div>
          </div>
          <div class="metric-row">
            <div class="metric-label-row">
              <span>亮部对比度 (Contrast Std)</span>
              <span id="mConVal">0.00</span>
            </div>
            <div class="metric-bar-bg"><div class="metric-bar-fill" id="mConBar" style="width:0%"></div></div>
          </div>
          <div class="metric-row">
            <div class="metric-label-row">
              <span>色彩和谐度 (Color Harmony)</span>
              <span id="mHarVal">0.00</span>
            </div>
            <div class="metric-bar-bg"><div class="metric-bar-fill" id="mHarBar" style="width:0%"></div></div>
          </div>
          <div class="metric-row">
            <div class="metric-label-row">
              <span>多尺度结构熵 (Structural Entropy)</span>
              <span id="mStrVal">0.00</span>
            </div>
            <div class="metric-bar-bg"><div class="metric-bar-fill" id="mStrBar" style="width:0%"></div></div>
          </div>
          <div class="metric-row">
            <div class="metric-label-row">
              <span>视觉焦点显著度 (Focal Saliency)</span>
              <span id="mFocVal">0.00</span>
            </div>
            <div class="metric-bar-bg"><div class="metric-bar-fill" id="mFocBar" style="width:0%"></div></div>
          </div>
        </div>
      </section>

      <!-- 2. Export & Download Options -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">💾 导出与下载作品 (Download)</span>
        </div>
        <div style="display:grid; grid-template-columns: 1fr 1fr; gap: 0.45rem;">
          <button type="button" class="btn btn-success" onclick="downloadCurrentPNG()">
            💾 保存当前 PNG
          </button>
          <button type="button" class="btn btn-primary" onclick="downloadHighResPNG(1200)">
            🖼️ 导出 1200×1200
          </button>
        </div>
        <div style="font-size:0.71rem; color:var(--muted);">
          导出的 PNG 自动内嵌无损 <code style="color:var(--accent);">GenerativeArt-Recipe</code> 元数据块，随时可拖入下方解析器完整还原所有滑块参数。
        </div>
      </section>

      <!-- 3. Reproducible CLI Command & Go Code -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">💻 CLI 复现命令</span>
          <button type="button" class="btn btn-sm" id="copyCliBtn" onclick="copyTextBox('cliCommandBox', 'copyCliBtn')">复制命令</button>
        </div>
        <div class="code-box" id="cliCommandBox">go run ./cmd/genart render ...</div>

        <div class="panel-header" style="margin-top:0.25rem;">
          <span class="panel-title">🧩 Go 调用代码片段</span>
          <button type="button" class="btn btn-sm" id="copyGoBtn" onclick="copyTextBox('goSnippetBox', 'copyGoBtn')">复制 Go 代码</button>
        </div>
        <div class="code-box" id="goSnippetBox"></div>
      </section>

      <!-- 4. PNG Recipe Drag-and-Drop Inspector -->
      <section class="panel-card">
        <div class="panel-header">
          <span class="panel-title">🔍 PNG 配方逆向解析器</span>
        </div>
        <div class="dropzone" id="recipeDropzone" onclick="document.getElementById('recipeFileInput').click()">
          📂 点击选择或拖入任意由 GenerativeArt 生成的 PNG 图片，自动提取内嵌参数并还原到工作台滑块
          <input type="file" id="recipeFileInput" accept="image/png" style="display:none" onchange="onRecipeFilePicked(event)" />
        </div>
        <div id="inspectResultMsg" style="font-size:0.72rem; color:var(--muted); display:none;"></div>
      </section>

    </aside>

  </div>

  <script>
    // Global Studio State
    let ALL_ENGINES = [];
    let FILTERED_ENGINES = [];
    let PALETTE_NAMES = [];
    let PALETTE_COLORS = {};
    let currentCategory = "all";
    let currentEngine = null;
    let currentParams = {};
    let currentImageBase64 = "";
    let renderHistory = [];
    let renderTimer = null;
    let activeAbortCtrl = null;
    let isRendering = false;
    let pendingRenderAfterCurrent = false;

    const POST_FX_SPECS = [
      { key: "bloom",      label: "Bloom 辉光",         min: 0,   max: 0.85, step: 0.01, def: 0 },
      { key: "relief",     label: "Relief 2.5D 浮雕",   min: 0,   max: 1.00, step: 0.02, def: 0 },
      { key: "chromatic",  label: "Chromatic 色散",     min: 0,   max: 2.00, step: 0.05, def: 0 },
      { key: "dof",        label: "DoF 薄透镜景深",     min: 0,   max: 10.0, step: 0.20, def: 0 },
      { key: "kuwahara",   label: "Kuwahara 油画滤镜",  min: 0,   max: 6,    step: 1,    def: 0 },
      { key: "paper",      label: "Paper 棉纸纹理",     min: 0,   max: 0.60, step: 0.02, def: 0 },
      { key: "vignette",   label: "Vignette 暗角",      min: 0,   max: 0.60, step: 0.02, def: 0 },
      { key: "deckle",     label: "Deckle 手工毛边框",  min: 0,   max: 0.07, step: 0.005, def: 0 },
      { key: "contrast",   label: "Contrast 对比度",    min: 0.6, max: 1.50, step: 0.02, def: 1.0 },
      { key: "saturation", label: "Saturation 饱和度",  min: 0.0, max: 1.80, step: 0.02, def: 1.0 }
    ];
    let currentPostFX = {};
    POST_FX_SPECS.forEach(sp => { currentPostFX[sp.key] = sp.def; });

    const CATEGORY_LABELS = {
      geom: "几何/分形 · Geom",
      chaos: "混沌/吸引子 · Chaos",
      flow: "流场/物理 · Flow",
      media: "水墨/材质 · Media",
      bio: "生物/形态 · Bio"
    };

    async function initStudio() {
      buildPostFXSliders();
      setupDropzone();

      try {
        const resp = await fetch("/api/engines");
        const data = await resp.json();
        ALL_ENGINES = data.engines || [];
        PALETTE_NAMES = data.palettes || [];
        PALETTE_COLORS = data.palette_colors || {};

        document.getElementById("engineCountBadge").textContent =
          ALL_ENGINES.length + " Arts · 可视化调参工作台";

        buildPaletteSelectors();
        filterArtSelect();

        // Check URL query param ?art=...
        const urlParams = new URLSearchParams(window.location.search);
        const requestedArt = urlParams.get("art");
        const requestedSeed = urlParams.get("seed");
        if (requestedSeed) {
          document.getElementById("seedInput").value = requestedSeed;
        }
        if (requestedArt && ALL_ENGINES.some(e => e.id === requestedArt)) {
          selectEngineById(requestedArt, false);
        } else if (ALL_ENGINES.length > 0) {
          selectEngineById(ALL_ENGINES[0].id, false);
        }
        await triggerRender(false);
      } catch (err) {
        document.getElementById("renderStatusBadge").textContent = "加载引擎失败: " + err.message;
      }
    }

    function buildPaletteSelectors() {
      const sel = document.getElementById("paletteSelect");
      sel.innerHTML = PALETTE_NAMES.map(name => {
        return '<option value="' + name + '">' + name.toUpperCase() + '</option>';
      }).join("");

      const qGrid = document.getElementById("paletteQuickGrid");
      qGrid.innerHTML = PALETTE_NAMES.map(name => {
        const cols = PALETTE_COLORS[name] || ["#38bdf8", "#a855f7", "#f43f5e"];
        const spans = cols.slice(0, 5).map(c => '<span style="background:' + c + '"></span>').join("");
        return '<button type="button" class="pal-chip" data-pal="' + name + '" onclick="pickPalette(\'' + name + '\')">' +
          '<span class="pal-mini-bar">' + spans + '</span>' +
          '<span style="overflow:hidden;text-overflow:ellipsis;">' + name + '</span>' +
          '</button>';
      }).join("");
    }

    function updatePaletteUI(palName) {
      const sel = document.getElementById("paletteSelect");
      if (sel.value !== palName && PALETTE_NAMES.includes(palName)) {
        sel.value = palName;
      }
      const cols = PALETTE_COLORS[palName] || [];
      const swatchBar = document.getElementById("activePaletteSwatches");
      swatchBar.innerHTML = cols.map(c =>
        '<span class="swatch-dot" style="background:' + c + '" title="' + c + '"></span>'
      ).join("");

      document.querySelectorAll(".pal-chip").forEach(btn => {
        btn.classList.toggle("active", btn.getAttribute("data-pal") === palName);
      });
    }

    function pickPalette(name) {
      document.getElementById("paletteSelect").value = name;
      onPaletteChange();
    }

    function onPaletteChange() {
      const val = document.getElementById("paletteSelect").value;
      updatePaletteUI(val);
      scheduleRender();
    }

    function resetPaletteToDefault() {
      if (!currentEngine) return;
      pickPalette(currentEngine.default_palette);
    }

    function setCategory(cat) {
      currentCategory = cat;
      document.querySelectorAll(".cat-pill").forEach(btn => {
        btn.classList.toggle("active", btn.getAttribute("data-cat") === cat);
      });
      filterArtSelect();
      if (FILTERED_ENGINES.length > 0 && (!currentEngine || !FILTERED_ENGINES.some(e => e.id === currentEngine.id))) {
        selectEngineById(FILTERED_ENGINES[0].id, true);
      }
    }

    function filterArtSelect() {
      const q = (document.getElementById("artSearchInput").value || "").trim().toLowerCase();
      FILTERED_ENGINES = ALL_ENGINES.filter(eng => {
        if (currentCategory !== "all" && eng.category !== currentCategory) return false;
        if (!q) return true;
        return eng.id.toLowerCase().includes(q) ||
               (eng.struct_name && eng.struct_name.toLowerCase().includes(q)) ||
               (eng.category && eng.category.toLowerCase().includes(q));
      });

      const sel = document.getElementById("artSelect");
      const prevVal = currentEngine ? currentEngine.id : sel.value;
      sel.innerHTML = FILTERED_ENGINES.map((eng, idx) => {
        const title = eng.struct_name || eng.id;
        return '<option value="' + eng.id + '">[' + (idx + 1) + '/' + FILTERED_ENGINES.length + '] ' + title + ' (' + eng.id + ')</option>';
      }).join("");

      if (FILTERED_ENGINES.some(e => e.id === prevVal)) {
        sel.value = prevVal;
      } else if (FILTERED_ENGINES.length > 0) {
        sel.value = FILTERED_ENGINES[0].id;
      }
    }

    function selectEngineById(artID, shouldRender) {
      const eng = ALL_ENGINES.find(e => e.id === artID);
      if (!eng) return;
      currentEngine = eng;

      const sel = document.getElementById("artSelect");
      if (sel.value !== artID) {
        // Ensure option is visible
        if (!FILTERED_ENGINES.some(e => e.id === artID)) {
          document.getElementById("artSearchInput").value = "";
          setCategory("all");
        }
        sel.value = artID;
      }

      document.getElementById("activeCatTag").textContent = CATEGORY_LABELS[eng.category] || eng.category;
      updatePaletteUI(eng.default_palette);

      // Initialize default parameters & build sliders
      currentParams = {};
      (eng.parameters || []).forEach(sp => {
        currentParams[sp.name] = sp.default;
      });
      buildEngineParamSliders(eng);

      // Update browser URL cleanly without reloading
      const url = new URL(window.location.href);
      url.searchParams.set("art", eng.id);
      window.history.replaceState({}, "", url.toString());

      if (shouldRender) {
        scheduleRender();
      }
    }

    function onArtSelectChange() {
      const artID = document.getElementById("artSelect").value;
      selectEngineById(artID, true);
    }

    function stepArt(delta) {
      if (!FILTERED_ENGINES.length) return;
      const curIdx = FILTERED_ENGINES.findIndex(e => currentEngine && e.id === currentEngine.id);
      const nextIdx = (curIdx + delta + FILTERED_ENGINES.length) % FILTERED_ENGINES.length;
      selectEngineById(FILTERED_ENGINES[nextIdx].id, true);
    }

    function computeSliderRange(sp) {
      const def = Number(sp.default);
      const nameLower = sp.name.toLowerCase();
      if (sp.type === "int" || sp.type === "uint") {
        if (def <= 0) return { min: 0, max: 20, step: 1 };
        if (def === 1) return { min: 1, max: 8, step: 1 };
        if (def <= 5) return { min: 1, max: Math.max(12, def * 3), step: 1 };
        if (def >= 10000) {
          return {
            min: Math.max(500, Math.round(def * 0.25)),
            max: Math.round(def * 2.2),
            step: Math.max(100, Math.round(def * 0.02))
          };
        }
        return {
          min: Math.max(1, Math.floor(def * 0.25)),
          max: Math.max(def + 6, Math.ceil(def * 2.4)),
          step: 1
        };
      }
      // float64
      if (Math.abs(def) < 1e-6) {
        return { min: 0, max: 2.0, step: 0.02 };
      }
      if (nameLower.includes("angle") && def > 100) {
        return { min: 130.0, max: 145.0, step: 0.005 };
      }
      if (def < 0) {
        const minV = Number((def * 2.2).toFixed(3));
        const maxV = Number((-def * 1.2).toFixed(3));
        const stepV = Number(Math.max(0.001, (maxV - minV) / 200).toFixed(4));
        return { min: minV, max: maxV, step: stepV };
      }
      const minV = Number((def * 0.2).toFixed(4));
      const maxV = Number((def * 2.4).toFixed(4));
      const stepV = Number(Math.max(0.0005, (maxV - minV) / 200).toFixed(4));
      return { min: minV, max: maxV, step: stepV };
    }

    function buildEngineParamSliders(eng) {
      const container = document.getElementById("paramSlidersContainer");
      const specs = eng.parameters || [];
      if (specs.length === 0) {
        container.innerHTML = '<div style="font-size:0.75rem; color:var(--muted); padding:0.35rem 0;">该算法使用纯确定性场或固定拓扑结构，可通过切换上方【调色板】、【随机种子】或下方【HDR 后处理滑块】实时探索不同视觉变化。</div>';
        return;
      }

      container.innerHTML = specs.map(sp => {
        const val = currentParams[sp.name] !== undefined ? currentParams[sp.name] : sp.default;
        if (sp.type === "string") {
          let presetsHtml = "";
          if (sp.name.toLowerCase().includes("seq")) {
            const seqs = ["AABAB", "AB", "AABB", "ABBB", "AABBB"];
            presetsHtml = '<div class="preset-chips">' + seqs.map(s =>
              '<button type="button" class="preset-chip" onclick="setStringParam(\'' + sp.name + '\', \'' + s + '\')">' + s + '</button>'
            ).join("") + '</div>';
          }
          return '<div class="param-item">' +
            '<div class="param-top">' +
              '<span class="param-name">' + sp.name + ' <span class="param-type-badge">string</span></span>' +
              '<button type="button" class="param-def-btn" onclick="resetSingleParam(\'' + sp.name + '\')">默认: ' + sp.default + '</button>' +
            '</div>' +
            '<input type="text" id="pinput_' + sp.name + '" class="text-input" value="' + val + '" oninput="onStringParamInput(\'' + sp.name + '\', this.value)" />' +
            presetsHtml +
          '</div>';
        }
        if (sp.type === "bool") {
          const checked = val ? "checked" : "";
          return '<div class="param-item">' +
            '<div class="param-top">' +
              '<span class="param-name">' + sp.name + ' <span class="param-type-badge">bool</span></span>' +
              '<input type="checkbox" id="pinput_' + sp.name + '" ' + checked + ' onchange="onBoolParamInput(\'' + sp.name + '\', this.checked)" />' +
            '</div>' +
          '</div>';
        }

        const rng = computeSliderRange(sp);
        return '<div class="param-item">' +
          '<div class="param-top">' +
            '<span class="param-name">' + sp.name + ' <span class="param-type-badge">' + sp.type + '</span></span>' +
            '<button type="button" class="param-def-btn" onclick="resetSingleParam(\'' + sp.name + '\')" title="点击恢复默认值">默认: ' + sp.default + '</button>' +
          '</div>' +
          '<div class="param-control-row">' +
            '<input type="range" class="range-slider" id="pslider_' + sp.name + '" min="' + rng.min + '" max="' + rng.max + '" step="' + rng.step + '" value="' + val + '" oninput="onNumericParamSlide(\'' + sp.name + '\', \'' + sp.type + '\', this.value)" />' +
            '<input type="number" class="num-input param-num-box" id="pnum_' + sp.name + '" step="' + rng.step + '" value="' + val + '" oninput="onNumericParamBox(\'' + sp.name + '\', \'' + sp.type + '\', this.value)" />' +
          '</div>' +
        '</div>';
      }).join("");
    }

    function onNumericParamSlide(name, type, rawVal) {
      const v = (type === "int" || type === "uint") ? parseInt(rawVal, 10) : parseFloat(rawVal);
      if (isNaN(v)) return;
      currentParams[name] = v;
      const numBox = document.getElementById("pnum_" + name);
      if (numBox) numBox.value = v;
      scheduleRender();
    }

    function onNumericParamBox(name, type, rawVal) {
      const v = (type === "int" || type === "uint") ? parseInt(rawVal, 10) : parseFloat(rawVal);
      if (isNaN(v)) return;
      currentParams[name] = v;
      const slider = document.getElementById("pslider_" + name);
      if (slider) {
        if (v < parseFloat(slider.min)) slider.min = v;
        if (v > parseFloat(slider.max)) slider.max = v;
        slider.value = v;
      }
      scheduleRender();
    }

    function onStringParamInput(name, val) {
      currentParams[name] = val;
      scheduleRender();
    }

    function setStringParam(name, val) {
      currentParams[name] = val;
      const inp = document.getElementById("pinput_" + name);
      if (inp) inp.value = val;
      scheduleRender();
    }

    function onBoolParamInput(name, checked) {
      currentParams[name] = checked;
      scheduleRender();
    }

    function resetSingleParam(name) {
      if (!currentEngine) return;
      const sp = (currentEngine.parameters || []).find(p => p.name === name);
      if (!sp) return;
      currentParams[name] = sp.default;
      syncParamInputsFromState();
      scheduleRender();
    }

    function resetEngineParams() {
      if (!currentEngine) return;
      (currentEngine.parameters || []).forEach(sp => {
        currentParams[sp.name] = sp.default;
      });
      syncParamInputsFromState();
      scheduleRender();
    }

    function syncParamInputsFromState() {
      if (!currentEngine) return;
      (currentEngine.parameters || []).forEach(sp => {
        const val = currentParams[sp.name];
        if (val === undefined) return;
        const slider = document.getElementById("pslider_" + sp.name);
        const numBox = document.getElementById("pnum_" + sp.name);
        const strInp = document.getElementById("pinput_" + sp.name);
        if (slider) {
          const numV = Number(val);
          if (numV < parseFloat(slider.min)) slider.min = numV;
          if (numV > parseFloat(slider.max)) slider.max = numV;
          slider.value = numV;
        }
        if (numBox) numBox.value = val;
        if (strInp) {
          if (sp.type === "bool") strInp.checked = !!val;
          else strInp.value = val;
        }
      });
    }

    function buildPostFXSliders() {
      const container = document.getElementById("postFxContainer");
      container.innerHTML = POST_FX_SPECS.map(sp => {
        const val = currentPostFX[sp.key];
        return '<div class="param-item">' +
          '<div class="param-top">' +
            '<span class="param-name">' + sp.label + '</span>' +
            '<button type="button" class="param-def-btn" onclick="resetSinglePostFX(\'' + sp.key + '\')">默认: ' + sp.def + '</button>' +
          '</div>' +
          '<div class="param-control-row">' +
            '<input type="range" class="range-slider" id="fxslider_' + sp.key + '" min="' + sp.min + '" max="' + sp.max + '" step="' + sp.step + '" value="' + val + '" oninput="onPostFXSlide(\'' + sp.key + '\', this.value)" />' +
            '<input type="number" class="num-input param-num-box" id="fxnum_' + sp.key + '" step="' + sp.step + '" value="' + val + '" oninput="onPostFXSlide(\'' + sp.key + '\', this.value)" />' +
          '</div>' +
        '</div>';
      }).join("");
    }

    function onPostFXSlide(key, rawVal) {
      const sp = POST_FX_SPECS.find(s => s.key === key);
      if (!sp) return;
      const v = sp.step === 1 ? parseInt(rawVal, 10) : parseFloat(rawVal);
      if (isNaN(v)) return;
      currentPostFX[key] = v;
      const sEl = document.getElementById("fxslider_" + key);
      const nEl = document.getElementById("fxnum_" + key);
      if (sEl && document.activeElement !== sEl) sEl.value = v;
      if (nEl && document.activeElement !== nEl) nEl.value = v;
      scheduleRender();
    }

    function resetSinglePostFX(key) {
      const sp = POST_FX_SPECS.find(s => s.key === key);
      if (!sp) return;
      currentPostFX[key] = sp.def;
      const sEl = document.getElementById("fxslider_" + key);
      const nEl = document.getElementById("fxnum_" + key);
      if (sEl) sEl.value = sp.def;
      if (nEl) nEl.value = sp.def;
      scheduleRender();
    }

    function resetPostFX() {
      POST_FX_SPECS.forEach(sp => {
        currentPostFX[sp.key] = sp.def;
        const sEl = document.getElementById("fxslider_" + sp.key);
        const nEl = document.getElementById("fxnum_" + sp.key);
        if (sEl) sEl.value = sp.def;
        if (nEl) nEl.value = sp.def;
      });
      scheduleRender();
    }

    function stepSeed(delta) {
      const inp = document.getElementById("seedInput");
      const cur = parseInt(inp.value, 10) || 20261105;
      inp.value = cur + delta;
      scheduleRender();
    }

    function rollNewSeed() {
      const nextSeed = Math.floor(100000 + Math.random() * 90000000);
      document.getElementById("seedInput").value = nextSeed;
      scheduleRender();
    }

    function onResolutionChange() {
      triggerRender(false);
    }

    function scheduleRender() {
      if (!document.getElementById("autoRenderToggle").checked) return;
      if (renderTimer) clearTimeout(renderTimer);
      renderTimer = setTimeout(() => {
        triggerRender(false);
      }, 130);
    }

    function buildRenderQuery(useRandom, overrideSize, isDownload) {
      const artID = currentEngine ? currentEngine.id : "muranoglassnautilus";
      const size = overrideSize || parseInt(document.getElementById("sizeSelect").value, 10) || 480;
      const seed = parseInt(document.getElementById("seedInput").value, 10) || 20261105;
      const palette = document.getElementById("paletteSelect").value || (currentEngine ? currentEngine.default_palette : "dunhuang");

      const qs = new URLSearchParams();
      qs.set("art", artID);
      qs.set("w", String(size));
      qs.set("h", String(size));
      qs.set("seed", String(seed));
      qs.set("palette", palette);

      if (useRandom) {
        qs.set("random", "1");
      } else if (currentEngine && currentEngine.parameters) {
        currentEngine.parameters.forEach(sp => {
          if (currentParams[sp.name] !== undefined) {
            qs.set("p." + sp.name, String(currentParams[sp.name]));
          }
        });
      }

      POST_FX_SPECS.forEach(sp => {
        const val = currentPostFX[sp.key];
        if (val !== undefined && Math.abs(val - sp.def) > 1e-5) {
          qs.set(sp.key, String(val));
        }
      });

      if (isDownload) {
        qs.set("download", "1");
      }
      return qs;
    }

    async function triggerRender(useRandom) {
      if (!currentEngine) return;
      if (renderTimer) {
        clearTimeout(renderTimer);
        renderTimer = null;
      }

      if (activeAbortCtrl) {
        activeAbortCtrl.abort();
      }
      const ctrl = new AbortController();
      activeAbortCtrl = ctrl;

      const overlay = document.getElementById("loadingOverlay");
      const statusBadge = document.getElementById("renderStatusBadge");
      overlay.classList.add("active");
      statusBadge.textContent = "⏳ 正在渲染...";

      const qs = buildRenderQuery(useRandom, null, false);
      try {
        const resp = await fetch("/api/render?" + qs.toString(), { signal: ctrl.signal });
        if (!resp.ok) {
          const msg = await resp.text();
          throw new Error(msg || ("HTTP " + resp.status));
        }
        const data = await resp.json();

        // Sync state if random=1 was used
        if (useRandom) {
          if (data.palette) {
            updatePaletteUI(data.palette);
          }
          if (data.params) {
            currentParams = Object.assign({}, data.params);
            syncParamInputsFromState();
          }
        }

        // Draw onto canvas
        currentImageBase64 = data.image_base64;
        await drawDataURLToCanvas(data.image_base64, data.width, data.height);

        // Update Aesthetic Metrics & Code Boxes
        updateAestheticPanel(data.aesthetic, data.render_ms);
        document.getElementById("cliCommandBox").textContent = data.cli_command || "";
        document.getElementById("goSnippetBox").textContent = buildGoSnippet(data);

        statusBadge.textContent = "✓ 渲染完成 (" + (data.render_ms || 0) + " ms)";
        overlay.classList.remove("active");

        pushRenderHistory(data);
      } catch (err) {
        if (err.name === "AbortError") return;
        overlay.classList.remove("active");
        statusBadge.textContent = "⚠️ 渲染错误: " + err.message;
      }
    }

    function drawDataURLToCanvas(dataURL, w, h) {
      return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => {
          const canvas = document.getElementById("renderCanvas");
          canvas.width = w || img.width;
          canvas.height = h || img.height;
          const ctx = canvas.getContext("2d");
          ctx.clearRect(0, 0, canvas.width, canvas.height);
          ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
          resolve();
        };
        img.onerror = reject;
        img.src = dataURL;
      });
    }

    function randomizeCurrentArt() {
      const nextSeed = Math.floor(100000 + Math.random() * 90000000);
      document.getElementById("seedInput").value = nextSeed;
      triggerRender(true);
    }

    function surpriseRandomArt() {
      const pool = FILTERED_ENGINES.length > 0 ? FILTERED_ENGINES : ALL_ENGINES;
      if (!pool.length) return;
      const pick = pool[Math.floor(Math.random() * pool.length)];
      const nextSeed = Math.floor(100000 + Math.random() * 90000000);
      document.getElementById("seedInput").value = nextSeed;
      selectEngineById(pick.id, false);
      triggerRender(true);
    }

    function resetAllAndRender() {
      if (!currentEngine) return;
      document.getElementById("seedInput").value = 20261105;
      updatePaletteUI(currentEngine.default_palette);
      (currentEngine.parameters || []).forEach(sp => {
        currentParams[sp.name] = sp.default;
      });
      syncParamInputsFromState();
      POST_FX_SPECS.forEach(sp => {
        currentPostFX[sp.key] = sp.def;
        const sEl = document.getElementById("fxslider_" + sp.key);
        const nEl = document.getElementById("fxnum_" + sp.key);
        if (sEl) sEl.value = sp.def;
        if (nEl) nEl.value = sp.def;
      });
      triggerRender(false);
    }

    function updateAestheticPanel(aes, renderMs) {
      document.getElementById("renderTimeTag").textContent = (renderMs !== undefined ? renderMs : "--") + " ms";
      if (!aes) return;
      const score = Number(aes.total_score || 0);
      document.getElementById("scoreTotalVal").textContent = score.toFixed(1) + " / 100";

      const gradeEl = document.getElementById("scoreGradeBadge");
      if (score >= 82) {
        gradeEl.textContent = "S 级 · 典藏佳作";
      } else if (score >= 72) {
        gradeEl.textContent = "A+ 级 · 视觉均衡";
      } else if (score >= 60) {
        gradeEl.textContent = "A 级 · 结构清晰";
      } else {
        gradeEl.textContent = "B 级 · 实验探索";
      }

      setMetricBar("mDyn", aes.dynamic_range);
      setMetricBar("mCon", aes.contrast_std);
      setMetricBar("mHar", aes.color_harmony);
      setMetricBar("mStr", aes.structural_entropy);
      setMetricBar("mFoc", aes.focal_saliency);
    }

    function setMetricBar(prefix, val) {
      const v = Math.max(0, Math.min(1, Number(val || 0)));
      document.getElementById(prefix + "Val").textContent = v.toFixed(2);
      document.getElementById(prefix + "Bar").style.width = Math.round(v * 100) + "%";
    }

    function buildGoSnippet(data) {
      const structName = (currentEngine && currentEngine.struct_name) || data.struct_name || data.art;
      const ctorName = structName.charAt(0).toUpperCase() + structName.slice(1);
      const argsList = [];
      if (currentEngine && currentEngine.parameters) {
        currentEngine.parameters.forEach(sp => {
          const v = data.params && data.params[sp.name] !== undefined ? data.params[sp.name] : sp.default;
          if (sp.type === "string") argsList.push('"' + v + '"');
          else argsList.push(String(v));
        });
      }
      const palConst = capitalizePalette(data.palette || "dunhuang");
      return 'c := generativeart.NewCanva(' + data.width + ', ' + data.height + ')\n' +
        'c.SetSeedLocal(' + data.seed + ')\n' +
        'c.SetColorSchema(common.' + palConst + ')\n' +
        'c.FillBackground()\n' +
        'c.Draw(arts.New' + ctorName + '(' + argsList.join(", ") + '))\n' +
        'c.ToPNG("' + data.art + '.png")';
    }

    function capitalizePalette(p) {
      if (p === "ukiyoe") return "UkiyoE";
      if (p === "pasteldream") return "PastelDream";
      return p.charAt(0).toUpperCase() + p.slice(1);
    }

    function pushRenderHistory(data) {
      const entry = {
        id: Date.now() + "_" + Math.random().toString(36).slice(2, 6),
        art: data.art,
        palette: data.palette,
        seed: data.seed,
        width: data.width,
        height: data.height,
        params: Object.assign({}, data.params || {}),
        postFX: Object.assign({}, currentPostFX),
        aesthetic: data.aesthetic,
        render_ms: data.render_ms,
        cli_command: data.cli_command,
        image_base64: data.image_base64
      };
      renderHistory.unshift(entry);
      if (renderHistory.length > 18) renderHistory.pop();
      renderHistoryFilmstrip(entry.id);
    }

    function renderHistoryFilmstrip(activeId) {
      document.getElementById("historyCountLabel").textContent = renderHistory.length + " 幅记录";
      const strip = document.getElementById("historyStrip");
      strip.innerHTML = renderHistory.map(item => {
        const cls = item.id === activeId ? "hist-thumb active" : "hist-thumb";
        const scoreTxt = item.aesthetic && item.aesthetic.total_score ? item.aesthetic.total_score.toFixed(0) : "";
        return '<div class="' + cls + '" onclick="restoreHistoryEntry(\'' + item.id + '\')" title="' + item.art + ' · palette=' + item.palette + ' · seed=' + item.seed + (scoreTxt ? ' · Score ' + scoreTxt : '') + '">' +
          '<img src="' + item.image_base64 + '" alt="' + item.art + '" />' +
        '</div>';
      }).join("");
    }

    async function restoreHistoryEntry(entryId) {
      const item = renderHistory.find(x => x.id === entryId);
      if (!item) return;
      selectEngineById(item.art, false);
      document.getElementById("seedInput").value = item.seed;
      updatePaletteUI(item.palette);
      currentParams = Object.assign({}, item.params);
      syncParamInputsFromState();

      currentPostFX = Object.assign({}, item.postFX);
      POST_FX_SPECS.forEach(sp => {
        const val = currentPostFX[sp.key] !== undefined ? currentPostFX[sp.key] : sp.def;
        const sEl = document.getElementById("fxslider_" + sp.key);
        const nEl = document.getElementById("fxnum_" + sp.key);
        if (sEl) sEl.value = val;
        if (nEl) nEl.value = val;
      });

      currentImageBase64 = item.image_base64;
      await drawDataURLToCanvas(item.image_base64, item.width, item.height);
      updateAestheticPanel(item.aesthetic, item.render_ms);
      document.getElementById("cliCommandBox").textContent = item.cli_command || "";
      document.getElementById("goSnippetBox").textContent = buildGoSnippet(item);
      renderHistoryFilmstrip(item.id);
    }

    function downloadCurrentPNG() {
      if (!currentImageBase64 || !currentEngine) return;
      const seed = document.getElementById("seedInput").value || "2026";
      const pal = document.getElementById("paletteSelect").value || "default";
      const a = document.createElement("a");
      a.href = currentImageBase64;
      a.download = currentEngine.id + "_s" + seed + "_" + pal + ".png";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }

    function downloadHighResPNG(size) {
      if (!currentEngine) return;
      const qs = buildRenderQuery(false, size, true);
      const a = document.createElement("a");
      a.href = "/api/render?" + qs.toString();
      a.download = currentEngine.id + "_" + size + "x" + size + ".png";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }

    function copyTextBox(boxId, btnId) {
      const txt = document.getElementById(boxId).textContent;
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(txt);
      }
      const btn = document.getElementById(btnId);
      const old = btn.textContent;
      btn.textContent = "已复制 ✓";
      setTimeout(() => { btn.textContent = old; }, 1400);
    }

    function toggleCanvasMat() {
      const frame = document.getElementById("canvasFrame");
      frame.classList.toggle("mat-light");
    }

    function toggleCanvasZoom() {
      const frame = document.getElementById("canvasFrame");
      const isActual = frame.classList.toggle("zoom-actual");
      document.getElementById("zoomBtn").textContent = isActual ? "🔍 1:1 原始像素" : "🔍 适应窗口";
    }

    function toggleTheme() {
      const html = document.documentElement;
      const next = html.getAttribute("data-theme") === "dark" ? "light" : "dark";
      html.setAttribute("data-theme", next);
      document.getElementById("themeBtn").textContent = next === "dark" ? "☀️ 浅色" : "🌙 深色";
    }

    function setupDropzone() {
      const dz = document.getElementById("recipeDropzone");
      dz.addEventListener("dragover", e => {
        e.preventDefault();
        dz.classList.add("dragover");
      });
      dz.addEventListener("dragleave", () => dz.classList.remove("dragover"));
      dz.addEventListener("drop", e => {
        e.preventDefault();
        dz.classList.remove("dragover");
        if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0]) {
          inspectPNGFile(e.dataTransfer.files[0]);
        }
      });
    }

    function onRecipeFilePicked(e) {
      if (e.target.files && e.target.files[0]) {
        inspectPNGFile(e.target.files[0]);
      }
    }

    async function inspectPNGFile(file) {
      const msgEl = document.getElementById("inspectResultMsg");
      msgEl.style.display = "block";
      msgEl.textContent = "正在解析 PNG 内嵌配方...";
      try {
        const buf = await file.arrayBuffer();
        const resp = await fetch("/api/inspect", {
          method: "POST",
          body: buf
        });
        if (!resp.ok) {
          throw new Error(await resp.text());
        }
        const meta = await resp.json();
        const engineName = (meta.engine_name || "").toLowerCase();
        const matched = ALL_ENGINES.find(e =>
          e.id === engineName || (e.struct_name && e.struct_name.toLowerCase() === engineName)
        );
        if (matched) {
          selectEngineById(matched.id, false);
        }
        if (meta.seed) {
          document.getElementById("seedInput").value = meta.seed;
        }
        if (meta.parameters) {
          Object.keys(meta.parameters).forEach(k => {
            currentParams[k] = meta.parameters[k];
          });
          syncParamInputsFromState();
        }
        msgEl.style.color = "var(--ok)";
        msgEl.textContent = "✓ 成功解析配方: " + (meta.engine_name || matched?.id) + " (Seed " + meta.seed + ")，已还原滑块并重新渲染！";
        await triggerRender(false);
      } catch (err) {
        msgEl.style.color = "var(--accent-3)";
        msgEl.textContent = "解析失败: " + err.message;
      }
    }

    initStudio();
  </script>
</body>
</html>
`
