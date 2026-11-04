(function () {
  'use strict';

  const lib = window.GoMemTreap;
  const svg = document.getElementById('treeSvg');
  if (!lib || !svg) return;
  const TreapAllocator = lib.TreapAllocator;
  const SNIPPETS = lib.SNIPPETS;

  /* ------------------------------------------------------------------ */
  /* constants                                                          */
  /* ------------------------------------------------------------------ */
  const NS = 'http://www.w3.org/2000/svg';
  const TL = 60, TR = 740, TOP = 84, BOTTOM = 418;
  const FX = 862, FY = 132, FGAP = 74;
  const POOL = { x: 862, y: 430 };
  const NW = 112, NH = 58;
  const TOTAL = 1024;
  const PALETTE = ['#f472b6', '#a78bfa', '#60a5fa', '#fbbf24', '#fb923c', '#f87171', '#c084fc', '#38bdf8', '#facc15', '#e879f9'];
  const DUR = { struct: 900, visit: 520, info: 520 };
  const HOLD = { struct: 380, visit: 180, info: 520 };
  const MAX_QUEUE = 4000;
  const HOT = { path: 1, visit: 1, cand: 1, target: 1, merge: 1 };

  const $ = (id) => document.getElementById(id);
  const ui = {
    stepCard: document.querySelector('.step-card'),
    stepOp: $('stepOp'),
    stepIdx: $('stepIdx'),
    stepDesc: $('stepDesc'),
    codeFn: $('codeFn'),
    codeBody: $('codeBody'),
    memBar: $('memBar'),
    allocList: $('allocList'),
    opLog: $('opLog'),
    btnPlay: $('btnPlay'),
    btnStep: $('btnStep'),
    btnSkip: $('btnSkip'),
    speed: $('speed'),
    speedOut: $('speedOut'),
    queueInfo: $('queueInfo'),
    sizeInput: $('sizeInput'),
    allocForm: $('allocForm'),
  };

  function el(tag, attrs, parent) {
    const e = document.createElementNS(NS, tag);
    if (attrs) for (const k in attrs) e.setAttribute(k, attrs[k]);
    if (parent) parent.appendChild(e);
    return e;
  }

  /* ------------------------------------------------------------------ */
  /* static svg scene                                                   */
  /* ------------------------------------------------------------------ */
  const defs = el('defs', {}, svg);
  const grad = el('linearGradient', { id: 'nodeGrad', x1: '0', y1: '0', x2: '0', y2: '1' }, defs);
  el('stop', { offset: '0%', 'stop-color': '#162a44' }, grad);
  el('stop', { offset: '100%', 'stop-color': '#0b1524' }, grad);
  const glow = el('filter', { id: 'glow', x: '-50%', y: '-50%', width: '200%', height: '200%' }, defs);
  el('feGaussianBlur', { stdDeviation: '6', result: 'b' }, glow);

  const zone = el('g', { class: 'zone' }, svg);
  el('rect', { class: 'zone-rect', x: 790, y: 62, width: 144, height: 326, rx: 16 }, zone);
  el('text', { class: 'zone-label', x: FX, y: 86 }, zone).textContent = '游离节点';
  const poolG = el('g', { class: 'pool' }, svg);
  el('rect', { class: 'pool-rect', x: 800, y: 404, width: 124, height: 52, rx: 12 }, poolG);
  el('text', { class: 'pool-label', x: 846, y: 435 }, poolG).textContent = '对象池';
  const poolCount = el('text', { class: 'pool-count', x: 898, y: 435 }, poolG);

  const gEdges = el('g', { class: 'edges' }, svg);
  const rootLabel = el('g', { class: 'root-label' }, svg);
  el('text', { x: 0, y: 0 }, rootLabel).textContent = 'sizeTree';
  el('path', { d: 'M0 6 L0 16', class: 'root-arrow' }, rootLabel);
  const emptyText = el('text', { class: 'empty-text', x: (TL + TR) / 2, y: 220 }, svg);
  emptyText.textContent = 'sizeTree = nil（没有空闲块）';
  const gNodes = el('g', { class: 'nodes' }, svg);

  const nodeEls = new Map();
  const edgeEls = new Map();

  /* ------------------------------------------------------------------ */
  /* layout                                                             */
  /* ------------------------------------------------------------------ */
  function layout(snap) {
    const pos = {};
    const order = [];
    let maxD = 0;
    (function walk(id, d) {
      if (id == null || d > 64) return;
      const n = snap.nodes[id];
      walk(n.left, d + 1);
      order.push([id, d]);
      if (d > maxD) maxD = d;
      walk(n.right, d + 1);
    })(snap.root, 0);

    const n = order.length;
    const spacing = n > 1 ? Math.min(128, (TR - TL) / (n - 1)) : 0;
    const scale = n > 1 ? Math.max(0.5, Math.min(1, spacing / 120)) : 1;
    const x0 = (TL + TR) / 2 - (spacing * (n - 1)) / 2;
    const gap = maxD ? Math.min(86, (BOTTOM - TOP) / maxD) : 0;
    order.forEach(function (item, i) {
      pos[item[0]] = { x: x0 + i * spacing, y: TOP + item[1] * gap, s: scale, o: 1 };
    });
    snap.floating.forEach(function (f, i) {
      pos[f.id] = { x: FX, y: FY + i * FGAP, s: 1, o: 1 };
    });
    return pos;
  }

  /* ------------------------------------------------------------------ */
  /* drawing                                                            */
  /* ------------------------------------------------------------------ */
  const ease = (t) => (t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2);
  const lerp = (a, b, t) => a + (b - a) * t;

  function makeNode(id) {
    const g = el('g', { class: 'node' }, gNodes);
    el('rect', { class: 'halo', x: -NW / 2 - 6, y: -NH / 2 - 6, width: NW + 12, height: NH + 12, rx: 18 }, g);
    el('rect', { class: 'box', x: -NW / 2, y: -NH / 2, width: NW, height: NH, rx: 13 }, g);
    const size = el('text', { class: 'nsize', x: 0, y: -3 }, g);
    const sizeNum = el('tspan', {}, size);
    const sizeUnit = el('tspan', { class: 'nunit', dx: '2' }, size);
    sizeUnit.textContent = 'B';
    const range = el('text', { class: 'nrange', x: 0, y: 15 }, g);
    el('rect', { class: 'nbar-bg', x: -42, y: 21, width: 84, height: 3.5, rx: 1.75 }, g);
    const bar = el('rect', { class: 'nbar', x: -42, y: 21, width: 0, height: 3.5, rx: 1.75 }, g);
    g.addEventListener('mouseenter', function () { linkHover(id, true); });
    g.addEventListener('mouseleave', function () { linkHover(id, false); });
    const rec = { g: g, sizeNum: sizeNum, range: range, bar: bar, key: '' };
    nodeEls.set(id, rec);
    return rec;
  }

  function linkHover(id, on) {
    const seg = ui.memBar.querySelector('[data-node="' + Number(id) + '"]');
    if (seg) seg.classList.toggle('hover', on);
    const n = nodeEls.get(String(id));
    if (n) n.g.classList.toggle('hover', on);
  }

  function edgesOf(snap) {
    const out = {};
    for (const id in snap.nodes) {
      const n = snap.nodes[id];
      if (n.left != null) out[id + '-' + n.left] = [String(id), String(n.left)];
      if (n.right != null) out[id + '-' + n.right] = [String(id), String(n.right)];
    }
    return out;
  }

  function draw(from, to, t) {
    const e = ease(t);
    const fs = from.step.snap, ts = to.step.snap;
    const hl = to.step.hl || {};
    const fromPool = ts.pool < fs.pool;
    const toPool = ts.pool > fs.pool;
    const cur = {};
    const floatSet = {};
    ts.floating.forEach(function (f) { floatSet[f.id] = 1; });

    const ids = new Set(Object.keys(from.pos).concat(Object.keys(to.pos)));
    ids.forEach(function (id) {
      let a = from.pos[id], b = to.pos[id];
      if (!a) a = fromPool ? { x: POOL.x, y: POOL.y, s: 0.3, o: 0 } : { x: b.x, y: b.y, s: b.s * 0.4, o: 0 };
      if (!b) b = toPool ? { x: POOL.x, y: POOL.y, s: 0.3, o: 0 } : { x: a.x, y: a.y, s: a.s * 0.4, o: 0 };
      const p = { x: lerp(a.x, b.x, e), y: lerp(a.y, b.y, e), s: lerp(a.s, b.s, e), o: lerp(a.o, b.o, e) };
      cur[id] = p;

      const rec = nodeEls.get(id) || makeNode(id);
      const data = ts.all[id] || fs.all[id];
      const key = data.start + ':' + data.end;
      if (rec.key !== key) {
        rec.key = key;
        const size = data.end - data.start;
        rec.sizeNum.textContent = String(size);
        rec.range.textContent = '[' + data.start + ', ' + data.end + ')';
        rec.bar.setAttribute('width', String(Math.max(1.5, (84 * size) / TOTAL)));
      }
      rec.g.setAttribute('transform', 'translate(' + p.x.toFixed(2) + ',' + p.y.toFixed(2) + ') scale(' + p.s.toFixed(3) + ')');
      rec.g.setAttribute('opacity', p.o.toFixed(3));
      const cls = 'node' + (hl[id] ? ' ' + hl[id] : '') + (floatSet[id] ? ' floating' : '') +
        (rec.g.classList.contains('hover') ? ' hover' : '');
      if (rec.g.getAttribute('class') !== cls) rec.g.setAttribute('class', cls);
    });

    const fe = edgesOf(fs), te = edgesOf(ts);
    const ekeys = new Set(Object.keys(fe).concat(Object.keys(te)));
    ekeys.forEach(function (k) {
      const pair = te[k] || fe[k];
      const pp = cur[pair[0]], cp = cur[pair[1]];
      if (!pp || !cp) return;
      let path = edgeEls.get(k);
      if (!path) {
        path = el('path', { class: 'edge' }, gEdges);
        edgeEls.set(k, path);
      }
      const op = fe[k] && te[k] ? 1 : te[k] ? e : 1 - e;
      const y1 = pp.y + (NH / 2) * pp.s, y2 = cp.y - (NH / 2) * cp.s;
      const my = (y1 + y2) / 2;
      path.setAttribute('d', 'M' + pp.x.toFixed(1) + ' ' + y1.toFixed(1) + ' C' + pp.x.toFixed(1) + ' ' + my.toFixed(1) +
        ' ' + cp.x.toFixed(1) + ' ' + my.toFixed(1) + ' ' + cp.x.toFixed(1) + ' ' + y2.toFixed(1));
      path.setAttribute('opacity', (op * Math.min(pp.o, cp.o)).toFixed(3));
      const hot = te[k] && HOT[hl[pair[0]]] && HOT[hl[pair[1]]];
      path.setAttribute('class', hot ? 'edge hot' : 'edge');
    });

    const rid = ts.root != null ? String(ts.root) : null;
    if (rid && cur[rid]) {
      const p = cur[rid];
      rootLabel.setAttribute('transform', 'translate(' + p.x.toFixed(1) + ',' + (p.y - NH / 2 * p.s - 22).toFixed(1) + ')');
      rootLabel.setAttribute('opacity', fs.root != null ? '1' : e.toFixed(3));
    } else {
      rootLabel.setAttribute('opacity', fs.root != null ? (1 - e).toFixed(3) : '0');
    }
    emptyText.setAttribute('opacity', ts.root == null ? e.toFixed(3) : fs.root == null ? (1 - e).toFixed(3) : '0');
  }

  function cleanup(to) {
    nodeEls.forEach(function (rec, id) {
      if (!to.pos[id]) {
        rec.g.remove();
        nodeEls.delete(id);
      }
    });
    const te = edgesOf(to.step.snap);
    edgeEls.forEach(function (path, k) {
      if (!te[k]) {
        path.remove();
        edgeEls.delete(k);
      }
    });
  }

  /* ------------------------------------------------------------------ */
  /* code panel                                                         */
  /* ------------------------------------------------------------------ */
  const parsed = {};
  let codeShown = null;

  function escapeHTML(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  function highlight(s) {
    return s.replace(/(\/\/.*$)|\b(func|if|else|for|return|switch|case|true|nil)\b|\b(int)\b|\b(Block|Allocator)\b|(-?\b\d+\b)/g,
      function (m, com, kw, ty, tp, num) {
        if (com) return '<span class="c-com">' + com + '</span>';
        if (kw) return '<span class="c-kw">' + kw + '</span>';
        if (ty) return '<span class="c-ty">' + ty + '</span>';
        if (tp) return '<span class="c-tp">' + tp + '</span>';
        return '<span class="c-num">' + num + '</span>';
      });
  }

  function parseSnippet(name) {
    if (parsed[name]) return parsed[name];
    const marks = {};
    const html = (SNIPPETS[name] || []).map(function (raw, i) {
      const m = raw.match(/\s*\/\/@(\w+)\s*$/);
      let text = m ? raw.slice(0, m.index) : raw;
      if (m && !(m[1] in marks)) marks[m[1]] = i;
      text = text.replace(/\t/g, '    ');
      return '<div class="cl"><span class="ln">' + (i + 1) + '</span><span class="ct">' + highlight(escapeHTML(text)) + '</span></div>';
    }).join('');
    parsed[name] = { marks: marks, html: html };
    return parsed[name];
  }

  function renderCode(name, mark) {
    const sn = parseSnippet(name);
    if (codeShown !== name) {
      ui.codeBody.innerHTML = sn.html; // static, escaped source only
      codeShown = name;
      ui.codeFn.textContent = 'allocator.go · ' + name + '()';
    }
    const lines = ui.codeBody.children;
    const idx = mark && mark in sn.marks ? sn.marks[mark] : -1;
    for (let i = 0; i < lines.length; i++) lines[i].classList.toggle('active', i === idx);
    if (idx >= 0) {
      const line = lines[idx];
      const body = ui.codeBody;
      const top = line.offsetTop - body.clientHeight / 2 + line.offsetHeight / 2;
      body.scrollTo({ top: Math.max(0, top), behavior: 'smooth' });
    }
  }

  /* ------------------------------------------------------------------ */
  /* memory bar                                                         */
  /* ------------------------------------------------------------------ */
  const segEls = new Map();
  const memHL = document.createElement('div');
  memHL.className = 'memhl';
  ui.memBar.appendChild(memHL);

  function pct(v) {
    return ((v / TOTAL) * 100).toFixed(4) + '%';
  }

  function renderMem(step) {
    const snap = step.snap;
    const floatSet = {};
    snap.floating.forEach(function (f) { floatSet[f.id] = 1; });
    const frees = Object.keys(snap.all).map(function (k) { return snap.all[k]; })
      .sort(function (a, b) { return a.start - b.start || (b.end - b.start) - (a.end - a.start); });
    const items = [];
    let maxEnd = -1;
    frees.forEach(function (f) {
      if (f.end <= maxEnd) return;
      maxEnd = Math.max(maxEnd, f.end);
      items.push({ key: 'f' + f.id, start: f.start, end: f.end, type: 'free', node: f.id, floating: !!floatSet[f.id] });
    });
    step.allocs.forEach(function (a) {
      items.push({ key: 'a' + a.id, start: a.offset, end: a.offset + a.size, type: 'alloc', alloc: a });
    });
    const occ = items.slice().sort(function (a, b) { return a.start - b.start; });
    let cursor = 0;
    occ.forEach(function (it) {
      if (it.start > cursor) items.push({ key: 'p' + cursor, start: cursor, end: it.start, type: 'pending' });
      cursor = Math.max(cursor, it.end);
    });
    if (cursor < TOTAL) items.push({ key: 'p' + cursor, start: cursor, end: TOTAL, type: 'pending' });

    const seen = new Set();
    items.forEach(function (it) {
      seen.add(it.key);
      let seg = segEls.get(it.key);
      if (!seg) {
        seg = document.createElement('div');
        seg.appendChild(document.createElement('span'));
        segEls.set(it.key, seg);
        ui.memBar.insertBefore(seg, memHL);
        if (it.type === 'alloc') {
          seg.setAttribute('role', 'button');
          seg.tabIndex = 0;
          const aid = it.alloc.id;
          seg.addEventListener('click', function () { requestFree(aid); });
          seg.addEventListener('keydown', function (ev) {
            if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); requestFree(aid); }
          });
        } else if (it.type === 'free') {
          const nid = it.node;
          seg.dataset.node = String(nid);
          seg.addEventListener('mouseenter', function () { linkHover(nid, true); });
          seg.addEventListener('mouseleave', function () { linkHover(nid, false); });
        }
      }
      const size = it.end - it.start;
      seg.className = 'seg ' + it.type + (it.floating ? ' floating' : '') + (seg.classList.contains('hover') ? ' hover' : '');
      seg.style.left = pct(it.start);
      seg.style.width = pct(size);
      const wide = size / TOTAL > 0.045;
      let label = '';
      if (it.type === 'free') label = wide ? String(size) : '';
      else if (it.type === 'alloc') {
        seg.style.setProperty('--c', it.alloc.color);
        seg.title = '#' + it.alloc.id + ' [' + it.start + ', ' + it.end + ') · 点击释放';
        label = wide ? '#' + it.alloc.id + ' · ' + size : '';
      }
      seg.firstChild.textContent = label;
    });
    segEls.forEach(function (seg, k) {
      if (!seen.has(k)) {
        seg.remove();
        segEls.delete(k);
      }
    });

    const m = step.memHL;
    if (m) {
      memHL.className = 'memhl show ' + m.cls;
      memHL.style.left = pct(m.offset);
      memHL.style.width = pct(m.size);
    } else {
      memHL.className = 'memhl';
    }
  }

  /* ------------------------------------------------------------------ */
  /* side info: alloc list & log                                        */
  /* ------------------------------------------------------------------ */
  function renderAllocList(list) {
    ui.allocList.textContent = '';
    if (!list.length) {
      const s = document.createElement('span');
      s.className = 'muted';
      s.textContent = '暂无，点击上方 Allocate 分配内存';
      ui.allocList.appendChild(s);
      return;
    }
    list.forEach(function (a) {
      const b = document.createElement('button');
      b.type = 'button';
      b.className = 'achip';
      b.style.setProperty('--c', a.color);
      b.title = '点击释放 Free(' + a.offset + ', ' + a.size + ')';
      const dot = document.createElement('i');
      const txt = document.createElement('span');
      txt.textContent = '#' + a.id + '  [' + a.offset + ', ' + (a.offset + a.size) + ')';
      const x = document.createElement('em');
      x.textContent = 'Free';
      b.appendChild(dot);
      b.appendChild(txt);
      b.appendChild(x);
      b.addEventListener('click', function () { requestFree(a.id); });
      ui.allocList.appendChild(b);
    });
  }

  function addLog(op) {
    if (op.li) return;
    const li = document.createElement('li');
    li.className = 'running ' + op.type;
    const l = document.createElement('code');
    l.textContent = op.label;
    const r = document.createElement('span');
    r.textContent = '执行中…';
    li.appendChild(l);
    li.appendChild(r);
    op.li = li;
    op.resEl = r;
    ui.opLog.insertBefore(li, ui.opLog.firstChild);
    while (ui.opLog.children.length > 40) ui.opLog.removeChild(ui.opLog.lastChild);
  }

  function finishLog(op) {
    addLog(op);
    if (op.done) return;
    op.done = true;
    op.li.classList.remove('running');
    if (op.fail) op.li.classList.add('fail');
    op.resEl.textContent = op.result;
  }

  /* ------------------------------------------------------------------ */
  /* player                                                             */
  /* ------------------------------------------------------------------ */
  const st = { shown: null, anim: null, queue: [], playing: true, speed: 1, holdUntil: 0 };

  function applyStep(step) {
    const op = step.op;
    ui.stepOp.textContent = op ? op.label : '就绪';
    ui.stepIdx.textContent = op ? step.idx + ' / ' + step.count : '';
    ui.stepDesc.textContent = step.desc;
    ui.stepCard.classList.toggle('fail', !!step.fail);
    ui.stepCard.classList.remove('flash');
    void ui.stepCard.offsetWidth;
    ui.stepCard.classList.add('flash');
    renderCode(step.code, step.mark);
    renderMem(step);
    poolCount.textContent = '× ' + step.snap.pool;
    renderAllocList(step.allocs);
    if (op && step.idx === 1) addLog(op);
    if (op && step.idx === step.count) finishLog(op);
    updateQueueInfo();
  }

  function startStep(step, now, dur) {
    const to = { step: step, pos: layout(step.snap) };
    st.anim = { from: st.shown, to: to, start: now, dur: dur != null ? dur : DUR[step.kind] / st.speed };
    applyStep(step);
  }

  function finishAnim(now) {
    const a = st.anim;
    st.anim = null;
    draw(a.to, a.to, 1);
    cleanup(a.to);
    st.shown = a.to;
    st.holdUntil = now + HOLD[a.to.step.kind] / st.speed;
    updateQueueInfo();
  }

  function frame(now) {
    if (st.anim) {
      const a = st.anim;
      const t = a.dur <= 0 ? 1 : Math.min(1, (now - a.start) / a.dur);
      draw(a.from, a.to, t);
      if (t >= 1) finishAnim(now);
    } else if (st.playing && st.queue.length && now >= st.holdUntil) {
      startStep(st.queue.shift(), now);
    }
    requestAnimationFrame(frame);
  }

  function updateQueueInfo() {
    let ops = 0;
    for (let i = 0; i < st.queue.length; i++) if (st.queue[i].idx === 1) ops++;
    ui.queueInfo.textContent = ops ? '队列中还有 ' + ops + ' 个操作' : '';
  }

  function setPlaying(p) {
    st.playing = p;
    ui.btnPlay.textContent = p ? '暂停' : '播放';
    ui.btnPlay.setAttribute('aria-pressed', String(p));
  }

  function stepOnce() {
    const now = performance.now();
    if (st.anim) finishAnim(now);
    else if (st.queue.length) startStep(st.queue.shift(), now);
  }

  function skipToEnd() {
    const now = performance.now();
    if (st.anim) finishAnim(now);
    if (!st.queue.length) return;
    st.queue.forEach(function (s) {
      if (s.idx === 1) addLog(s.op);
      if (s.idx === s.count) finishLog(s.op);
    });
    const last = st.queue[st.queue.length - 1];
    st.queue = [];
    startStep(last, now, 650 / st.speed);
  }

  /* ------------------------------------------------------------------ */
  /* model / controller                                                 */
  /* ------------------------------------------------------------------ */
  let model, allocs, nextAllocId;

  function initialStep() {
    model.allocsView = [];
    model.begin();
    model.emit('Allocate', null, '初始状态：整块 1024 字节都是空闲的，树中只有一个节点。点击上方按钮开始分配，或直接点「自动演示」。');
    return model.end()[0];
  }

  function reset() {
    model = new TreapAllocator(TOTAL);
    allocs = [];
    nextAllocId = 1;
    st.queue = [];
    st.anim = null;
    nodeEls.forEach(function (r) { r.g.remove(); });
    nodeEls.clear();
    edgeEls.forEach(function (p) { p.remove(); });
    edgeEls.clear();
    ui.opLog.textContent = '';
    const s0 = initialStep();
    st.shown = { step: s0, pos: layout(s0.snap) };
    draw(st.shown, st.shown, 1);
    applyStep(s0);
  }

  function runOp(op, fn) {
    if (st.queue.length > MAX_QUEUE) return null;
    model.allocsView = allocs.map(function (a) { return Object.assign({}, a); });
    model.begin();
    let res;
    try {
      res = fn();
    } catch (err) {
      model.end();
      console.error(err);
      return null;
    }
    const steps = model.end();
    steps.forEach(function (s, i) {
      s.op = op;
      s.idx = i + 1;
      s.count = steps.length;
    });
    st.queue.push.apply(st.queue, steps);
    updateQueueInfo();
    return res;
  }

  function requestAlloc(size) {
    size = Math.floor(Number(size));
    if (!Number.isFinite(size) || size < 1 || size > TOTAL) return null;
    const op = { label: 'Allocate(' + size + ')', type: 'op-alloc' };
    return runOp(op, function () {
      const off = model.allocate(size);
      if (off < 0) {
        op.fail = true;
        op.result = '→ -1 无可用块';
        return null;
      }
      const a = { id: nextAllocId, offset: off, size: size, color: PALETTE[(nextAllocId - 1) % PALETTE.length] };
      nextAllocId++;
      allocs.push(a);
      model.allocsView = allocs.map(function (x) { return Object.assign({}, x); });
      model.emit('Allocate', 'done', '分配完成：返回 offset = ' + off + '，调用方拿到 [' + off + ', ' + (off + size) + ')，即 #' + a.id);
      op.result = '→ offset ' + off + '（#' + a.id + '）';
      return a.id;
    });
  }

  const FREE_RESULT = { both: '三块合并', none: '新建节点', left: '与左邻合并', right: '与右邻合并' };

  function requestFree(id) {
    const a = allocs.find(function (x) { return x.id === id; });
    if (!a) return;
    allocs = allocs.filter(function (x) { return x !== a; });
    const op = { label: 'Free(' + a.offset + ', ' + a.size + ')', type: 'op-free' };
    runOp(op, function () {
      model.allocsView = allocs.map(function (x) { return Object.assign({}, x); });
      const kind = model.free(a.offset, a.size);
      model.curMemHL = null;
      model.emit('Free', 'end', '回收完成：#' + a.id + ' 已归还，' + FREE_RESULT[kind] + '。');
      op.result = '#' + a.id + ' · ' + FREE_RESULT[kind];
    });
  }

  function randomOp() {
    const used = allocs.reduce(function (s, a) { return s + a.size; }, 0);
    const sizes = [32, 48, 64, 96, 128, 160, 200, 256];
    if (allocs.length && (Math.random() < 0.45 || TOTAL - used < 64)) {
      requestFree(allocs[Math.floor(Math.random() * allocs.length)].id);
    } else {
      requestAlloc(sizes[Math.floor(Math.random() * sizes.length)]);
    }
  }

  function runDemo() {
    reset();
    const ids = [];
    [200, 120, 300, 64].forEach(function (s) { ids.push(requestAlloc(s)); });
    requestFree(ids[1]);          // isolated hole -> new node, rotates up
    ids.push(requestAlloc(100));  // best-fit into the hole
    requestFree(ids[0]);          // no neighbour
    requestFree(ids[2]);          // merge with left neighbour
    requestFree(ids[3]);          // merge both sides
    requestFree(ids[4]);          // merge both sides -> single block
    const whole = requestAlloc(TOTAL); // exact fit -> node back to pool, empty tree
    requestFree(whole);           // node reused from pool
    setPlaying(true);
  }

  /* ------------------------------------------------------------------ */
  /* wiring                                                             */
  /* ------------------------------------------------------------------ */
  ui.allocForm.addEventListener('submit', function (ev) {
    ev.preventDefault();
    const v = Number(ui.sizeInput.value);
    if (!Number.isInteger(v) || v < 1 || v > TOTAL) {
      ui.sizeInput.setCustomValidity('请输入 1 ~ ' + TOTAL + ' 之间的整数');
      ui.sizeInput.reportValidity();
      return;
    }
    ui.sizeInput.setCustomValidity('');
    requestAlloc(v);
  });
  ui.sizeInput.addEventListener('input', function () { ui.sizeInput.setCustomValidity(''); });
  document.querySelectorAll('#quickSizes [data-size]').forEach(function (b) {
    b.addEventListener('click', function () {
      ui.sizeInput.value = b.dataset.size;
      requestAlloc(Number(b.dataset.size));
    });
  });
  $('btnRandom').addEventListener('click', randomOp);
  $('btnDemo').addEventListener('click', runDemo);
  $('btnReset').addEventListener('click', reset);
  ui.btnPlay.addEventListener('click', function () { setPlaying(!st.playing); });
  ui.btnStep.addEventListener('click', function () {
    setPlaying(false);
    stepOnce();
  });
  ui.btnSkip.addEventListener('click', skipToEnd);
  ui.speed.addEventListener('input', function () {
    st.speed = Number(ui.speed.value) || 1;
    ui.speedOut.textContent = st.speed + '×';
  });

  document.addEventListener('keydown', function (ev) {
    const tag = (ev.target && ev.target.tagName) || '';
    if (/INPUT|TEXTAREA|SELECT|BUTTON/.test(tag) || ev.metaKey || ev.ctrlKey || ev.altKey) return;
    const viz = $('viz').getBoundingClientRect();
    if (viz.bottom < 0 || viz.top > window.innerHeight) return;
    if (ev.key === ' ') {
      ev.preventDefault();
      setPlaying(!st.playing);
    } else if (ev.key === 'ArrowRight') {
      ev.preventDefault();
      setPlaying(false);
      stepOnce();
    }
  });

  reset();
  requestAnimationFrame(frame);

  // Autoplay the demo once the visualizer scrolls into view for the first time.
  if ('IntersectionObserver' in window) {
    const io = new IntersectionObserver(function (entries) {
      if (entries.some(function (en) { return en.isIntersecting; })) {
        io.disconnect();
        if (!st.queue.length && !allocs.length) runDemo();
      }
    }, { threshold: 0.35 });
    io.observe($('viz'));
  }
})();
