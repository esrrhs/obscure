(function () {
  const IMAGE_EXTS = ['png', 'jpg', 'webp', 'jpeg', 'gif', 'svg'];
  const CONTENT = '../content';

  const bgA = document.getElementById('bg-a');
  const bgB = document.getElementById('bg-b');
  const textBox = document.getElementById('text-box');
  const choicesBox = document.getElementById('choices');
  const overlay = document.getElementById('overlay');

  let meta = { title: 'Game', start: '1', typeSpeed: 40 };
  let currentBg = bgA;
  let nextBg = bgB;
  let typingTimer = null;
  let typingDone = true;
  let fullText = '';
  let onSkip = null;

  let audioCtx = null;
  function getAudioCtx() {
    if (!audioCtx) {
      const AC = window.AudioContext || window.webkitAudioContext;
      if (AC) audioCtx = new AC();
    }
    if (audioCtx && audioCtx.state === 'suspended') audioCtx.resume();
    return audioCtx;
  }

  // 短促的"木鱼/墨点"感音效：正弦+快速衰减包络
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
      fullText = text;
      typingDone = false;

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

  async function goto(sceneId) {
    choicesBox.innerHTML = '';
    clearTyping();

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
      return;
    }

    swapBg(bgSrc);
    sfxScene();
    await typeText(text.replace(/\r\n/g, '\n').replace(/\s+$/,''));
    renderChoices(parseChoices(choicesTxt));
  }

  overlay.addEventListener('click', () => {
    if (!typingDone && onSkip) { sfxSkip(); onSkip(); }
  });

  (async function main() {
    try {
      const metaTxt = await fetchText(`${CONTENT}/meta.txt`);
      meta = Object.assign(meta, parseMeta(metaTxt));
      if (meta.title) document.title = meta.title;
      await goto(String(meta.start));
    } catch (e) {
      textBox.classList.add('error');
      textBox.textContent = `启动失败：${e.message}\n请通过静态服务器访问（例：python3 -m http.server 8000）。`;
    }
  })();
})();
