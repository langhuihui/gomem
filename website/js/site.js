(function () {
  'use strict';

  // sticky nav shadow
  const nav = document.getElementById('nav');
  const onScroll = function () { nav.classList.toggle('scrolled', window.scrollY > 8); };
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();

  // copy install command
  const copyBtn = document.getElementById('copyInstall');
  const tip = document.getElementById('copyTip');
  copyBtn.addEventListener('click', function () {
    const text = document.getElementById('installCmd').textContent;
    const done = function (ok) {
      tip.textContent = ok ? '已复制' : '复制失败';
      copyBtn.classList.add('copied');
      setTimeout(function () {
        tip.textContent = '复制';
        copyBtn.classList.remove('copied');
      }, 1600);
    };
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(function () { done(true); }, function () { done(false); });
    } else {
      const range = document.createRange();
      range.selectNodeContents(document.getElementById('installCmd'));
      const sel = window.getSelection();
      sel.removeAllRanges();
      sel.addRange(range);
      let ok = false;
      try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
      sel.removeAllRanges();
      done(ok);
    }
  });

  // tabs
  const tabs = document.querySelectorAll('.tab');
  tabs.forEach(function (t) {
    t.addEventListener('click', function () {
      tabs.forEach(function (x) { x.classList.toggle('active', x === t); });
      document.querySelectorAll('.tab-panel').forEach(function (p) {
        p.classList.toggle('active', p.id === t.dataset.tab);
      });
    });
  });

  // benchmark bars
  document.querySelectorAll('.bars').forEach(function (bars) {
    const max = Number(bars.dataset.max) || 1;
    bars.querySelectorAll('.bf').forEach(function (bf) {
      bf.style.setProperty('--w', Math.min(100, (Number(bf.dataset.v) / max) * 100).toFixed(2) + '%');
    });
  });

  // reveal on scroll
  const items = document.querySelectorAll('.reveal');
  if ('IntersectionObserver' in window) {
    const io = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        if (en.isIntersecting) {
          en.target.classList.add('in');
          io.unobserve(en.target);
        }
      });
    }, { threshold: 0.12 });
    items.forEach(function (it) { io.observe(it); });
  } else {
    items.forEach(function (it) { it.classList.add('in'); });
  }
})();
