/**
 * 循人课表 — 启动
 * 1. 由 exe 提供的网页服务读取「班级课表.html」和（可选的）「场地课表_English.html」，每次打开都重新读取
 * 2. 直接双击 index.html 打开时，改为让使用者选择文件（可一次选两个）
 * 3. 定时心跳，让 exe 知道页面仍开着
 */
(function (TT) {
  'use strict';

  const { config, parser, app, util } = TT;
  const { esc, $ } = util;
  const LOCAL = new Error('local');

  /** 依序尝试 urls，回传第一个读得到的解析结果；全部失败则抛出最后的错误 */
  async function fetchFirst(urls) {
    let lastErr = LOCAL;
    if (!location.protocol.startsWith('http')) throw lastErr;
    for (const src of urls) {
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
    throw lastErr;
  }

  async function autoLoad() {
    const [cls, eng] = await Promise.allSettled([fetchFirst(config.sources.class), fetchFirst(config.sources.english)]);
    if (cls.status === 'rejected') throw cls.reason;
    return { data: cls.value, english: eng.status === 'fulfilled' && eng.value.kind === 'venue' ? eng.value : null };
  }

  /** 把使用者选的文件按内容分成 班级课表 / 英文场地课表 */
  async function readPicked(files) {
    const out = { data: null, english: null };
    for (const f of files) {
      const parsed = parser.parseTimetable(parser.decodeHtml(await f.arrayBuffer()));
      if (parsed.kind === 'venue') out.english = parsed;
      else out.data = parsed;
    }
    if (!out.data) throw new Error('没有选到「班级课表.html」。英文的「场地课表_English.html」要和它一起选。');
    return out;
  }

  function showPicker(err) {
    const msg = err && err !== LOCAL && err.message !== 'local' ? `<p class="err">${esc(err.message)}</p>` : '';
    $('out').innerHTML = `<div class="load">
      <h1>读取课表</h1>
      <p>请把「班级课表.html」（和「场地课表_English.html」）放进 Teacher-TimeTable.exe 旁边的 data 文件夹，再按 F5。也可以直接在这里选择文件。</p>
      ${msg}
      <label class="drop" id="drop">
        <input type="file" id="file" accept=".html,.htm" multiple>
        <b>选择或拖入课表文件</b>
        <span>「班级课表.html」必选；「场地课表_English.html」可一起选</span>
      </label>
    </div>`;
    const drop = $('drop');
    const take = (files) =>
      files &&
      files.length &&
      readPicked([...files])
        .then(launch)
        .catch(showPicker);
    $('file').onchange = (e) => take(e.target.files);
    drop.ondragover = (e) => {
      e.preventDefault();
      drop.classList.add('over');
    };
    drop.ondragleave = () => drop.classList.remove('over');
    drop.ondrop = (e) => {
      e.preventDefault();
      drop.classList.remove('over');
      take(e.dataTransfer.files);
    };
  }

  function launch({ data, english }) {
    app.start(data, english);
  }

  function heartbeat() {
    if (!location.protocol.startsWith('http')) return;
    const ping = () => fetch('/__ping', { cache: 'no-store' }).catch(() => {});
    ping();
    setInterval(ping, config.heartbeatMs);
  }

  /** 由 exe 打开时，读取 config/app.json（经 /config.json）覆盖默认设置 */
  async function loadRemoteConfig() {
    if (!location.protocol.startsWith('http')) return;
    try {
      const r = await fetch('config.json?t=' + Date.now(), { cache: 'no-store' });
      if (r.ok) Object.assign(config, await r.json());
    } catch (e) {
      /* 读不到就用默认值 */
    }
  }

  function initChrome() {
    document.title = config.appName;
    $('brandTitle').textContent = config.headerTitle;
    $('year').textContent = new Date().getFullYear();
    $('schoolName').textContent = config.schoolName;
    $('schoolNameEn').textContent = config.schoolNameEn;
    $('version').textContent = config.version ? 'v' + config.version : '';
    const gh = $('gh');
    if (config.githubUrl) gh.href = config.githubUrl;
    else gh.hidden = true;
  }

  $('out').innerHTML = '<div class="load"><div class="spin"></div><p>正在读取课表…</p></div>';
  loadRemoteConfig().then(initChrome).then(autoLoad).then(launch).catch(showPicker);
  heartbeat();
})(window.TT);
