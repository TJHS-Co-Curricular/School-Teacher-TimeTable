/**
 * 循人课表 — 界面
 * TT.app.start(data) 接收 parser 的结果，整合出老师课表并渲染三个视图：
 *   #teacher[=名字]  老师列表 / 个人课表
 *   #class[=班级]    班级列表 / 班级课表
 *   #summary         老师节数总表
 */
(function (TT) {
  'use strict';

  const { esc, zh, $, to24 } = TT.util;
  let started = false;

  function start(DATA) {
    if (started) return;
    started = true;

    const DAYS = ['星期一', '星期二', '星期三', '星期四', '星期五', '星期六'];
    const DAYS_EN = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    const DSHORT = ['一', '二', '三', '四', '五', '六'];
    const WEEKDAYS = 5;
    const GROUP_RE = /^(E\d+|CO\w*\d+|Maker\d+|科实[A-Z])$/i;
    const ACT = new Set(['联课活动', '公共选修', '共同备课']);

    const TIMES = DATA.times.map(([a, b]) => [to24(a), to24(b)]);
    const PER = DATA.periods.map((p) => p.replace('*', ''));
    const gapAfter = (i) => DATA.recess.includes(i + 1) && i + 1 < PER.length;

    /* 科目 → 色系 */
    function kind(s) {
      if (ACT.has(s)) return 'act';
      if (/华文|书法/.test(s)) return 'cn';
      if (/英文/.test(s)) return 'en';
      if (/马来/.test(s)) return 'bm';
      if (/数/.test(s) || s === 'S高') return 'ma';
      if (/科|物|化|生/.test(s) && !/生涯/.test(s)) return 'sc';
      if (/历|地理/.test(s)) return 'hi';
      if (/会计|商业|经济/.test(s)) return 'bz';
      if (/电脑/.test(s)) return 'ict';
      if (/美术|音乐|设计/.test(s)) return 'art';
      if (/体育/.test(s)) return 'pe';
      if (/辅导|生涯|道/.test(s)) return 'gd';
      if (/专题/.test(s)) return 'prj';
      return 'oth';
    }
    const kcls = (s) => {
      const k = kind(s);
      return k === 'act' ? 'act' : 'k-' + k;
    };

    /* ---------- 由班级课表整合老师课表 ---------- */
    const classOrder = {};
    DATA.classes.forEach((c, i) => (classOrder[c.n] = i));
    const C = {};
    DATA.classes.forEach((c) => (C[c.n] = c));
    const T = {};
    DATA.classes.forEach((c) =>
      c.g.forEach((row, d) => {
        if (d >= WEEKDAYS) return;
        row.forEach((cell, p) => {
          if (cell.length < 2) return;
          const [s, name] = cell;
          const t = T[name] || (T[name] = { name, slots: {}, pair: {} });
          (t.slots[d + '-' + p] = t.slots[d + '-' + p] || []).push({ s, c: c.n });
          const k = c.n + '\u0001' + s;
          t.pair[k] = (t.pair[k] || 0) + 1;
        });
      }),
    );
    Object.values(T).forEach((t) => {
      t.group = GROUP_RE.test(t.name);
      t.day = Array(WEEKDAYS).fill(0);
      Object.keys(t.slots).forEach((k) => t.day[+k.split('-')[0]]++);
      t.total = t.day.reduce((a, b) => a + b, 0);
      const subjCount = {};
      Object.entries(t.pair).forEach(([k, n]) => {
        const s = k.split('\u0001')[1];
        subjCount[s] = (subjCount[s] || 0) + n;
      });
      t.subj = Object.keys(subjCount).sort((a, b) => subjCount[b] - subjCount[a]);
      t.cls = [...new Set(Object.keys(t.pair).map((k) => k.split('\u0001')[0]))].sort(
        (a, b) => classOrder[a] - classOrder[b],
      );
      t.homeroom = DATA.classes.filter((c) => c.hr === t.name).map((c) => c.n);
    });
    const teacherNames = Object.keys(T).sort((a, b) => T[a].group - T[b].group || zh(a, b));
    const MAXDAY = Math.max(...Object.values(T).flatMap((t) => t.day));
    const MAXTOT = Math.max(...Object.values(T).map((t) => t.total));
    const semester = (DATA.sem ? '第' + DATA.sem + '学期 · ' : '') + DATA.printed.replace('列印日期：', '资料更新 ');

    /* ---------- 共用组件 ---------- */
    const spark = (arr) =>
      `<span class="spark" aria-hidden="true">${arr.map((n) => `<i style="height:${Math.max(2, (n / MAXDAY) * 18)}px"></i>`).join('')}</span>`;
    const chip = (s) => `<span class="chip ${kcls(s)}">${esc(s)}</span>`;
    function headRow(extra) {
      let h = `<tr><th class="day"></th>`;
      PER.forEach((p, i) => {
        h += `<th><span class="p num">${esc(p)}${DATA.periods[i].includes('*') ? '*' : ''}</span><span class="num">${TIMES[i][0]}<br>${TIMES[i][1]}</span></th>`;
        if (gapAfter(i)) h += `<th class="gap"></th>`;
      });
      return h + (extra ? `<th class="cnt">节数</th>` : '') + `</tr>`;
    }
    const dayTh = (d) => `<th class="day"><b>${DAYS[d]}</b><span>${DAYS_EN[d]}</span></th>`;
    const sixthNote = DATA.note ? `<p class="note">${esc(DATA.note)}</p>` : '';

    /* 手机版按天列表：items(d,p) 返回 {s, w} 或 null */
    function dayList(nDays, items, counts) {
      let h = '<div class="days">';
      for (let d = 0; d < nDays; d++) {
        const rows = PER.map((p, i) => ({ i, it: items(d, i) })).filter((x) => x.it);
        h +=
          `<div class="dayc"><h3>${DAYS[d]}<span>${counts ? counts[d] + ' 节' : ''}</span></h3>` +
          (rows.length
            ? `<ol>${rows.map(({ i, it }) => `<li class="${kcls(it.s)}"><span class="p num">${esc(PER[i])}</span><span class="t num">${TIMES[i][0]}–${TIMES[i][1]}</span><span class="w"><b>${esc(it.s)}</b>${it.w ? ' · ' + esc(it.w) : ''}</span></li>`).join('')}</ol>`
            : `<div class="none">没有课</div>`) +
          `</div>`;
      }
      return h + '</div>';
    }
    function weekBars(day, total, label, mx) {
      mx = Math.max(mx || MAXDAY, 1);
      return `<div class="week">${day.map((n, d) => `<div class="d"><div class="bar"><i style="height:${(n / mx) * 100}%"></i></div><b class="num">${n}</b><span>${DSHORT[d]}</span></div>`).join('')}
    <div class="total"><b class="num">${total}</b><span>${label}</span></div></div>`;
    }
    function navBar(list, cur, backLabel) {
      const i = list.indexOf(cur);
      return `<div class="dnav">
    <button class="btn" data-act="back">← ${backLabel}</button><span class="sp"></span>
    <button class="btn" data-act="prev" ${i <= 0 ? 'disabled' : ''} title="键盘 ←">‹ ${i > 0 ? esc(list[i - 1]) : '上一个'}</button>
    <button class="btn" data-act="next" ${i >= list.length - 1 ? 'disabled' : ''} title="键盘 →">${i < list.length - 1 ? esc(list[i + 1]) : '下一个'} ›</button>
    <button class="btn" data-act="print">打印</button></div>`;
    }

    /* ---------- 老师：列表 ---------- */
    function matchT(t, q) {
      return (
        !q ||
        t.name.toLowerCase().includes(q) ||
        t.subj.some((s) => s.toLowerCase().includes(q)) ||
        t.cls.some((c) => c.toLowerCase().includes(q))
      );
    }
    function teacherList(q) {
      const list = teacherNames.filter((n) => matchT(T[n], q));
      const cn = list.filter((n) => !T[n].group && !/^[A-Za-z]/.test(n));
      const en = list.filter((n) => !T[n].group && /^[A-Za-z]/.test(n));
      const gp = list.filter((n) => T[n].group);
      const card = (n) => {
        const t = T[n];
        return `<button class="tcard" data-go="teacher=${esc(n)}">
      <span class="nm">${esc(n)}</span><span class="tot"><b class="num">${t.total}</b>${spark(t.day)}<small>节 / 周</small></span>
      <span class="sj">${esc(t.subj.join('、'))}</span></button>`;
      };
      const sec = (title, arr) =>
        arr.length
          ? `<section class="section"><h2>${title}<span class="count">${arr.length}</span></h2><div class="tgrid">${arr.map(card).join('')}</div></section>`
          : '';
      const people = cn.length + en.length;
      return `<div class="pagehead"><div><h1>老师</h1><p>${people} 位老师，节数为星期一至星期五合计。点击查看个人课表。</p></div></div>
    ${list.length ? '' : `<div class="empty">找不到「${esc(q)}」</div>`}
    ${sec('华文名 · 按拼音', cn)}${sec('英文 / 马来文名', en)}${sec('分组 / 代号', gp)}`;
    }

    /* ---------- 老师：课表 ---------- */
    function teacherDetail(name) {
      const t = T[name];
      let body = '';
      for (let d = 0; d < WEEKDAYS; d++) {
        body += `<tr>${dayTh(d)}`;
        PER.forEach((_, p) => {
          const it = t.slots[d + '-' + p];
          if (it) {
            const s = it[0].s;
            body += `<td class="on ${kcls(s)}"><span class="s">${esc(s)}</span><span class="w">${it.map((x) => esc(x.c)).join('、')}</span></td>`;
          } else body += `<td class="slot"></td>`;
          if (gapAfter(p)) body += `<td class="gap"></td>`;
        });
        body += `<td class="count num">${t.day[d]}</td></tr>`;
      }
      const plan = Object.entries(t.pair)
        .map(([k, n]) => {
          const [c, s] = k.split('\u0001');
          return { c, s, n };
        })
        .sort((a, b) => classOrder[a.c] - classOrder[b.c] || zh(a.s, b.s));
      return `<div class="detail">${navBar(teacherNames, name, '全部老师')}
    <div class="dhead">
      <div><div class="kind">${t.group ? '分组 / 代号' : '老师'}</div><h1>${esc(name)}</h1>
        <div class="chips">${t.subj.map(chip).join('')}</div></div>
      ${weekBars(t.day, t.total, '节 / 周')}
      <div class="meta"><span>任教 <b>${t.cls.length}</b> 班</span>${t.homeroom.length ? `<span>班导师：<b>${t.homeroom.map(esc).join('、')}</b></span>` : ''}<span>最多一天 <b>${Math.max(...t.day)}</b> 节</span></div>
    </div>
    <div class="ttwrap"><table class="tt"><thead>${headRow(true)}</thead><tbody>${body}</tbody></table>${sixthNote}</div>
    ${dayList(
      WEEKDAYS,
      (d, p) => {
        const it = t.slots[d + '-' + p];
        return it ? { s: it[0].s, w: it.map((x) => x.c).join('、') } : null;
      },
      t.day,
    )}
    <div class="plan">${plan.map((r) => `<div class="prow"><span class="chip ${kcls(r.s)}">${esc(r.s)}</span><span class="who"><button class="lnk" data-go="class=${esc(r.c)}">${esc(r.c)}</button></span><b class="num">${r.n}<small>节</small></b></div>`).join('')}</div>
  </div>`;
    }

    /* ---------- 班级：列表 ---------- */
    function classList(q) {
      const by = {};
      DATA.classes
        .filter((c) => !q || c.n.toLowerCase().includes(q) || c.hr.toLowerCase().includes(q))
        .forEach((c) => (by[c.n.slice(0, 2)] = by[c.n.slice(0, 2)] || []).push(c));
      const gs = Object.keys(by);
      return `<div class="pagehead"><div><h1>班级</h1><p>${DATA.classes.length} 个班，按年级排列。</p></div></div>
    ${gs.length ? '' : `<div class="empty">找不到「${esc(q)}」</div>`}
    ${gs
      .map(
        (g) => `<section class="section"><h2>${g}<span class="count">${by[g].length} 班</span></h2><div class="cgrid">
      ${by[g].map((c) => `<button class="ccard" data-go="class=${esc(c.n)}"><b>${esc(c.n.slice(2))}</b><span>班导师 ${esc(c.hr)}</span></button>`).join('')}</div></section>`,
      )
      .join('')}`;
    }

    /* ---------- 班级：课表 ---------- */
    function classDetail(n) {
      const c = C[n];
      const cnt = c.g.map((row) => row.filter((x) => x.length).length);
      let body = '';
      c.g.forEach((row, d) => {
        body += `<tr class="${d >= WEEKDAYS ? 'sat' : ''}">${dayTh(d)}`;
        row.forEach((cell, p) => {
          if (!cell.length) body += `<td class="slot"></td>`;
          else {
            const k = kcls(cell[0]);
            const w = cell[1]
              ? T[cell[1]]
                ? `<button class="lnk" data-go="teacher=${esc(cell[1])}">${esc(cell[1])}</button>`
                : esc(cell[1])
              : '';
            body += `<td class="on ${k}"><span class="s">${esc(cell[0])}</span><span class="w">${w}</span></td>`;
          }
          if (gapAfter(p)) body += `<td class="gap"></td>`;
        });
        body += `</tr>`;
      });
      // 合并原版的「班级/教师 · 节数 · 科目」
      const plan = c.p.flat();
      const subjects = [...new Set(plan.map((r) => r[2]))];
      const weekTotal = cnt.slice(0, WEEKDAYS).reduce((a, b) => a + b, 0);
      return `<div class="detail">${navBar(
        DATA.classes.map((x) => x.n),
        n,
        '全部班级',
      )}
    <div class="dhead">
      <div><div class="kind">${esc(n.slice(0, 2))} · 班级</div><h1>${esc(n)}</h1>
        <div class="chips">${subjects.slice(0, 16).map(chip).join('')}</div></div>
      ${weekBars(cnt.slice(0, WEEKDAYS), weekTotal, '节 / 周', PER.length)}
      <div class="meta"><span>班导师：${T[c.hr] ? `<button class="lnk" data-go="teacher=${esc(c.hr)}"><b>${esc(c.hr)}</b></button>` : `<b>${esc(c.hr)}</b>`}</span><span>科任老师 <b>${new Set(plan.map((r) => r[0])).size}</b> 位</span></div>
    </div>
    <div class="ttwrap"><table class="tt"><thead>${headRow(false)}</thead><tbody>${body}</tbody></table>${sixthNote}</div>
    ${dayList(
      c.g.length,
      (d, p) => {
        const x = c.g[d][p];
        return x.length ? { s: x[0], w: x[1] || '' } : null;
      },
      cnt,
    )}
    <div class="plan">${plan
      .map(
        (r) => `<div class="prow"><span class="chip ${kcls(r[2])}">${esc(r[2])}</span>
      <span class="who">${T[r[0]] ? `<button class="lnk" data-go="teacher=${esc(r[0])}">${esc(r[0])}</button>` : esc(r[0])}</span><b class="num">${esc(r[1])}<small>节</small></b></div>`,
      )
      .join('')}</div>
  </div>`;
    }

    /* ---------- 节数总表 ---------- */
    let sortKey = 'total',
      sortDir = -1;
    function summary(q) {
      const list = teacherNames.map((n) => T[n]).filter((t) => matchT(t, q));
      const val = (t, k) =>
        k === 'name' ? t.name : k === 'total' ? t.total : k === 'subj' ? t.subj[0] || '' : t.day[+k];
      const cmp = (a, b) => {
        const x = val(a, sortKey),
          y = val(b, sortKey);
        const r = typeof x === 'number' ? x - y : zh(x, y);
        return r * sortDir || zh(a.name, b.name);
      };
      const cols = [
        ['name', '老师', ''],
        ['0', '一', 'c'],
        ['1', '二', 'c'],
        ['2', '三', 'c'],
        ['3', '四', 'c'],
        ['4', '五', 'c'],
        ['total', '总节数', ''],
        ['subj', '科目', ''],
        ['cls', '班级', ''],
      ];
      const th = cols
        .map(
          ([k, l, c]) =>
            `<th class="${c}" data-k="${k}" ${k === sortKey ? `aria-sort="${sortDir > 0 ? 'ascending' : 'descending'}"` : ''}>${l}</th>`,
        )
        .join('');
      const row = (t) => `<tr><td><button class="lnk" data-go="teacher=${esc(t.name)}">${esc(t.name)}</button></td>
    ${t.day.map((n) => `<td class="h c num"><span class="${n >= 7 ? 'hi' : ''}" style="--v:${n}">${n}</span></td>`).join('')}
    <td class="tot"><span class="tbar"><b class="num">${t.total}</b><i style="width:${(t.total / MAXTOT) * 110}px"></i></span></td>
    <td class="sj">${t.subj.map(chip).join(' ')}</td><td class="cl">${esc(t.cls.join('、'))}</td></tr>`;
      const people = list.filter((t) => !t.group).sort(cmp),
        groups = list.filter((t) => t.group).sort(cmp);
      const all = teacherNames.map((n) => T[n]).filter((t) => !t.group);
      const avg = (all.reduce((a, t) => a + t.total, 0) / all.length).toFixed(1);
      return `<div class="pagehead"><div><h1>老师节数总表</h1><p>星期一至星期五，每个时段算 1 节。不含早自习、班导师时间、晨读、联课活动、公共选修与共同备课。</p></div>
      <div class="stats"><div class="stat"><b class="num">${all.length}</b><span>位老师</span></div><div class="stat"><b class="num">${avg}</b><span>平均节数</span></div><div class="stat"><b class="num">${Math.max(...all.map((t) => t.total))}</b><span>最多节数</span></div>
      <button class="btn noprint" data-act="print" style="align-self:center">打印</button></div></div>
    <div class="sumwrap"><table class="sum"><thead><tr>${th}</tr></thead><tbody>
      ${people.map(row).join('')}
      ${groups.length ? `<tr class="grp"><td colspan="9">分组 / 代号（E1–E12、COMM、Maker、科实等，课表中以代号代替老师名字）</td></tr>${groups.map(row).join('')}` : ''}
      ${list.length ? '' : `<tr><td colspan="9" class="empty">找不到「${esc(q)}」</td></tr>`}
    </tbody></table></div>`;
    }

    /* ---------- 路由 ---------- */
    const out = $('out'),
      searchEl = $('search');
    let mode = 'teacher',
      cur = null;
    function parse() {
      const h = decodeURIComponent(location.hash.slice(1));
      const m = h.match(/^(teacher|class|summary)(?:=(.+))?$/);
      mode = m ? m[1] : 'teacher';
      cur = m && m[2] ? m[2] : null;
      if (cur && !(mode === 'teacher' ? T[cur] : C[cur])) cur = null;
    }
    function go(h) {
      location.hash = h;
    }
    function render() {
      parse();
      document.querySelectorAll('#tabs button').forEach((b) => b.classList.toggle('on', b.dataset.mode === mode));
      $('searchWrap').hidden = !!cur;
      const q = searchEl.value.trim().toLowerCase();
      if (mode === 'summary') out.innerHTML = summary(q);
      else if (mode === 'teacher') out.innerHTML = cur ? teacherDetail(cur) : teacherList(q);
      else out.innerHTML = cur ? classDetail(cur) : classList(q);
      document.title = (cur ? cur + ' · ' : '') + TT.config.appName;
    }
    function step(d) {
      const L = mode === 'teacher' ? teacherNames : DATA.classes.map((c) => c.n);
      const i = L.indexOf(cur);
      if (i >= 0 && L[i + d]) go(mode + '=' + L[i + d]);
    }
    $('tabs').addEventListener('click', (e) => {
      const b = e.target.closest('button');
      if (b) {
        if (b.dataset.mode === mode && !cur) return;
        go(b.dataset.mode);
      }
    });
    searchEl.addEventListener('input', () => {
      if (!cur) render();
    });
    out.addEventListener('click', (e) => {
      const g = e.target.closest('[data-go]');
      if (g) {
        go(g.dataset.go);
        return;
      }
      const a = e.target.closest('[data-act]');
      if (a) {
        const act = a.dataset.act;
        if (act === 'back') go(mode);
        else if (act === 'prev') step(-1);
        else if (act === 'next') step(1);
        else if (act === 'print') window.print();
        return;
      }
      const th = e.target.closest('th[data-k]');
      if (th && th.dataset.k !== 'cls') {
        const k = th.dataset.k;
        sortDir = sortKey === k ? -sortDir : k === 'name' || k === 'subj' ? 1 : -1;
        sortKey = k;
        render();
      }
    });
    document.addEventListener('keydown', (e) => {
      if (!cur || e.target.tagName === 'INPUT') return;
      if (e.key === 'ArrowLeft') step(-1);
      else if (e.key === 'ArrowRight') step(1);
      else if (e.key === 'Escape') go(mode);
    });
    window.addEventListener('hashchange', () => {
      render();
      window.scrollTo(0, 0);
    });
    $('sub').textContent = semester;
    render();
  }

  TT.app = { start };
})((window.TT = window.TT || {}));
