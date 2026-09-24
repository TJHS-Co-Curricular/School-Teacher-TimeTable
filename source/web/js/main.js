/**
 * 循人课表 — 启动
 * 1. 由 exe 提供的网页服务读取「班级课表.html」（每次打开都重新读取）
 * 2. 直接双击 index.html 打开时，改为让使用者选择文件
 * 3. 定时心跳，让 exe 知道页面仍开着
 */
(function (TT) {
  'use strict';

  const { config, parser, app, util } = TT;
  const { esc, $ } = util;

  async function autoLoad() {
    let lastErr = null;
    if (location.protocol.startsWith('http')) {
      for (const src of config.sources) {
        try {
          const r = await fetch(src + '?t=' + Date.now(), { cache: 'no-store' });
          if (!r.ok) {
            lastErr = new Error(`找不到「${decodeURIComponent(src)}」（${r.status}）`);
            continue;
          }
          return parser.parseTimetable(parser.decodeHtml(await r.arrayBuffer()));
        } catch (e) {
          lastErr = e;
        }
      }
    }
    throw lastErr || new Error('local');
  }

  function showPicker(err) {
    const msg = err && err.message !== 'local' ? `<p class="err">${esc(err.message)}</p>` : '';
    $('out').innerHTML = `<div class="load">
      <h1>读取班级课表</h1>
      <p>请把「班级课表.html」放在 Teacher-TimeTable.exe 旁边后重新打开。也可以直接在这里选择文件。</p>
      ${msg}
      <label class="drop" id="drop">
        <input type="file" id="file" accept=".html,.htm">
        <b>选择或拖入「班级课表.html」</b>
        <span>从 eSchool 列印页面「另存为」的 HTML 文件</span>
      </label>
    </div>`;
    const drop = $('drop');
    const take = (f) =>
      f &&
      f.arrayBuffer().then((b) => {
        try {
          app.start(parser.parseTimetable(parser.decodeHtml(b)));
        } catch (e) {
          showPicker(e);
        }
      });
    $('file').onchange = (e) => take(e.target.files[0]);
    drop.ondragover = (e) => {
      e.preventDefault();
      drop.classList.add('over');
    };
    drop.ondragleave = () => drop.classList.remove('over');
    drop.ondrop = (e) => {
      e.preventDefault();
      drop.classList.remove('over');
      take(e.dataTransfer.files[0]);
    };
  }

  function heartbeat() {
    if (!location.protocol.startsWith('http')) return;
    const ping = () => fetch('/__ping', { cache: 'no-store' }).catch(() => {});
    ping();
    setInterval(ping, config.heartbeatMs);
  }

  function initChrome() {
    document.title = config.appName;
    $('brandTitle').textContent = config.headerTitle;
    $('year').textContent = new Date().getFullYear();
    $('schoolName').textContent = config.schoolName;
    $('schoolNameEn').textContent = config.schoolNameEn;
    const gh = $('gh');
    if (config.githubUrl) gh.href = config.githubUrl;
    else gh.hidden = true;
  }

  initChrome();
  $('out').innerHTML = '<div class="load"><div class="spin"></div><p>正在读取「班级课表.html」…</p></div>';
  autoLoad().then(app.start).catch(showPicker);
  heartbeat();
})(window.TT);
