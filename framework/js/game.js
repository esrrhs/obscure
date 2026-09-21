(function () {
  const IMAGE_EXTS = ['png', 'jpg', 'webp', 'jpeg', 'gif', 'svg'];
  const CONTENT = '../content';

  const bgA = document.getElementById('bg-a');
  const bgB = document.getElementById('bg-b');
  const portraitEl = document.getElementById('portrait');
  const textBox = document.getElementById('text-box');
  const choicesBox = document.getElementById('choices');
  const overlay = document.getElementById('overlay');

  let meta = { title: 'Game', start: '1', typeSpeed: 40 };
  let currentBg = bgA;
  let nextBg = bgB;
  let typingTimer = null;
  let typingDone = true;
  let onSkip = null;

  let lines = [];
  let lineIndex = 0;
  let awaitingAdvance = false;
  let sceneChoicesTxt = null;
  let busy = false;

  // 角色名 -> 立绘目录 id；立绘路径缓存
  let speakerMap = {};
  const portraitCache = {}; // id -> src | null
  let currentPortraitId = null;

  let audioCtx = null;
  function getAudioCtx() {
    if (!audioCtx) {
      const AC = window.AudioContext || window.webkitAudioContext;
      if (AC) audioCtx = new AC();
    }
    if (audioCtx && audioCtx.state === 'suspended') audioCtx.resume();
    return audioCtx;
  }

  function playTone({ freq = 660, dur = 0.09, type = 'sine', gain = 0.15 } = {}) {
    const ctx = getAudioCtx();
    if (!ctx) return;
    const t0 = ctx.currentTime;
    const osc = ctx.createOscillator();
    const g = ctx.createGain();
    osc.type = type;
    osc.frequency.setValueAtTime(freq, t0);
    g.gain.setValueAtTime(0, t0);
    g.gain.linearRampToValueAtTime(gain, t0 + 0.005);
    g.gain.exponentialRampToValueAtTime(0.0001, t0 + dur);
    osc.connect(g).connect(ctx.destination);
    osc.start(t0);
    osc.stop(t0 + dur + 0.02);
  }

  function sfxClick()  { playTone({ freq: 880, dur: 0.08, type: 'triangle', gain: 0.18 }); }
  function sfxSkip()   { playTone({ freq: 520, dur: 0.06, type: 'sine',     gain: 0.10 }); }
  function sfxScene()  { playTone({ freq: 330, dur: 0.20, type: 'sine',     gain: 0.14 }); }
  function sfxRestart(){ playTone({ freq: 220, dur: 0.30, type: 'sine',     gain: 0.16 }); }
  function sfxAdvance(){ playTone({ freq: 700, dur: 0.05, type: 'sine',     gain: 0.08 }); }

  async function fetchText(url) {
    const r = await fetch(url);
    if (!r.ok) throw new Error(`HTTP ${r.status} on ${url}`);
    return await r.text();
  }

  async function tryFetchText(url) {
    try {
      const r = await fetch(url);
      if (!r.ok) return null;
      return await r.text();
    } catch {
      return null;
    }
  }

  function parseMeta(txt) {
    const out = {};
    txt.split(/\r?\n/).forEach(line => {
      const m = line.match(/^\s*([^#=\s][^=]*?)\s*=\s*(.*?)\s*$/);
      if (m) out[m[1]] = m[2];
    });
    return out;
  }

  function parseSpeakerMap(txt) {
    const out = {};
    if (!txt) return out;
    txt.split(/\r?\n/).forEach(line => {
      const m = line.match(/^\s*([^#=\s][^=]*?)\s*=\s*(.*?)\s*$/);
      if (m) out[m[1].trim()] = m[2].trim();
    });
    return out;
  }

  function parseChoices(txt) {
    if (!txt) return [];
    return txt.split(/\r?\n/).map(l => l.trim()).filter(Boolean).map(line => {
      const idx = line.lastIndexOf('|');
      if (idx < 0) return null;
      const text = line.slice(0, idx).trim();
      const next = line.slice(idx + 1).trim();
      if (!text || !next) return null;
      return { text, next };
    }).filter(Boolean);
  }

  function splitLines(text) {
    return text
      .replace(/\r\n/g, '\n')
      .replace(/\s+$/, '')
      .split('\n')
      .map(l => l.trim())
      .filter(Boolean);
  }

  // 从「角色："台词"」解析说话人；旁白返回 null
  function parseSpeaker(line) {
    const idx = line.search(/[：:]/);
    if (idx <= 0) return null;
    const q = line.charAt(idx + 1);
    if (q !== '"' && q !== '\u201c' && q !== '\u300c') return null;
    return line.slice(0, idx).trim();
  }

  function preloadImage(src) {
    return new Promise(resolve => {
      const img = new Image();
      img.onload = () => resolve(src);
      img.onerror = () => resolve(null);
      img.src = src;
    });
  }

  async function findBg(sceneId) {
    for (const ext of IMAGE_EXTS) {
      const src = `${CONTENT}/scenes/${sceneId}/bg.${ext}`;
      const ok = await preloadImage(src);
      if (ok) return ok;
    }
    return null;
  }

  async function findPortrait(charId) {
    if (charId in portraitCache) return portraitCache[charId];
    for (const ext of IMAGE_EXTS) {
      const src = `${CONTENT}/characters/${charId}/portrait.${ext}`;
      const ok = await preloadImage(src);
      if (ok) {
        portraitCache[charId] = ok;
        return ok;
      }
    }
    portraitCache[charId] = null;
    return null;
  }

  function hidePortrait() {
    portraitEl.classList.remove('show');
    currentPortraitId = null;
  }

  async function showPortraitForLine(line) {
    const speaker = parseSpeaker(line);
    if (!speaker) {
      // 旁白：收起立绘
      hidePortrait();
      return;
    }
    const charId = speakerMap[speaker];
    if (!charId) {
      hidePortrait();
      return;
    }
    if (charId === currentPortraitId) {
      portraitEl.classList.add('show');
      return;
    }
    const src = await findPortrait(charId);
    if (!src) {
      hidePortrait();
      return;
    }
    // 先淡出再换图，避免闪一下
    portraitEl.classList.remove('show');
    await new Promise(r => setTimeout(r, 120));
    portraitEl.style.backgroundImage = `url("${src}")`;
    currentPortraitId = charId;
    // 强制回流后再淡入
    void portraitEl.offsetWidth;
    portraitEl.classList.add('show');
  }

  function swapBg(src) {
    if (src) {
      nextBg.style.backgroundImage = `url("${src}")`;
    } else {
      nextBg.style.backgroundImage = 'none';
    }
    nextBg.classList.add('active');
    currentBg.classList.remove('active');
    [currentBg, nextBg] = [nextBg, currentBg];
  }

  function clearTyping() {
    if (typingTimer) { clearInterval(typingTimer); typingTimer = null; }
  }

  function typeText(text) {
    return new Promise(resolve => {
      clearTyping();
      textBox.classList.remove('error');
      textBox.textContent = '';
      typingDone = false;
      awaitingAdvance = false;

      const caret = document.createElement('span');
      caret.className = 'caret';
      caret.textContent = '▌';

      let i = 0;
      const speed = Math.max(5, Number(meta.typeSpeed) || 40);

      function finish() {
        clearTyping();
        textBox.textContent = text;
        textBox.appendChild(caret);
        typingDone = true;
        onSkip = null;
        resolve();
      }

      onSkip = finish;

      typingTimer = setInterval(() => {
        i++;
        textBox.textContent = text.slice(0, i);
        textBox.appendChild(caret);
        if (i >= text.length) finish();
      }, speed);
    });
  }

  function renderChoices(choices) {
    choicesBox.innerHTML = '';
    awaitingAdvance = false;
    if (!choices.length) {
      const btn = document.createElement('button');
      btn.className = 'choice restart';
      btn.textContent = '重新开始';
      btn.onclick = e => { e.stopPropagation(); sfxRestart(); goto(meta.start); };
      choicesBox.appendChild(btn);
      requestAnimationFrame(() => btn.classList.add('show'));
      return;
    }
    choices.forEach((c, i) => {
      const btn = document.createElement('button');
      btn.className = 'choice';
      btn.textContent = c.text;
      btn.onclick = e => { e.stopPropagation(); sfxClick(); goto(c.next); };
      choicesBox.appendChild(btn);
      setTimeout(() => btn.classList.add('show'), 80 * i + 30);
    });
  }

  async function playCurrentLine() {
    if (lineIndex >= lines.length) {
      renderChoices(parseChoices(sceneChoicesTxt));
      return;
    }
    const line = lines[lineIndex];
    await showPortraitForLine(line);
    await typeText(line);
    if (lineIndex >= lines.length - 1) {
      renderChoices(parseChoices(sceneChoicesTxt));
    } else {
      awaitingAdvance = true;
    }
  }

  async function advance() {
    if (busy) return;
    if (!typingDone && onSkip) {
      sfxSkip();
      onSkip();
      return;
    }
    if (awaitingAdvance && lineIndex < lines.length - 1) {
      busy = true;
      try {
        sfxAdvance();
        lineIndex++;
        awaitingAdvance = false;
        await playCurrentLine();
      } finally {
        busy = false;
      }
    }
  }

  async function goto(sceneId) {
    busy = true;
    choicesBox.innerHTML = '';
    clearTyping();
    awaitingAdvance = false;
    onSkip = null;
    typingDone = true;
    lines = [];
    lineIndex = 0;
    sceneChoicesTxt = null;
    hidePortrait();
    portraitEl.style.backgroundImage = 'none';

    const base = `${CONTENT}/scenes/${sceneId}`;
    let text, choicesTxt, bgSrc;
    try {
      [text, choicesTxt, bgSrc] = await Promise.all([
        fetchText(`${base}/text.txt`),
        tryFetchText(`${base}/choices.txt`),
        findBg(sceneId),
      ]);
    } catch (e) {
      textBox.classList.add('error');
      textBox.textContent = `无法加载幕 "${sceneId}"：${e.message}`;
      busy = false;
      return;
    }

    sceneChoicesTxt = choicesTxt;
    lines = splitLines(text);
    if (!lines.length) lines = [''];

    swapBg(bgSrc);
    sfxScene();
    try {
      await playCurrentLine();
    } finally {
      busy = false;
    }
  }

  overlay.addEventListener('click', () => { advance(); });

  (async function main() {
    try {
      const [metaTxt, mapTxt] = await Promise.all([
        fetchText(`${CONTENT}/meta.txt`),
        tryFetchText(`${CONTENT}/characters/index.txt`),
      ]);
      meta = Object.assign(meta, parseMeta(metaTxt));
      speakerMap = parseSpeakerMap(mapTxt);
      if (meta.title) document.title = meta.title;
      await goto(String(meta.start));
    } catch (e) {
      textBox.classList.add('error');
      textBox.textContent = `启动失败：${e.message}\n请通过静态服务器访问（例：python3 -m http.server 8000）。`;
    }
  })();
})();
