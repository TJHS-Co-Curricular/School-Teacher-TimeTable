/**
 * 循人课表 — 解析 eSchool 列印页面（GBK 编码的 HTML）
 * 支持两种：
 *   班级课表（教务处系统，每页「班级：…」）        kind = 'class'
 *   场地课表_English（英文系统，每页「场地：…」）  kind = 'venue'
 * 输出：{ kind, periods, recess, times, pre, classes, printed, sem, note }
 *   classes[i] = { n: 班级/场地名, hr: 第二行资料（班导师 / 课时总数）, g: 每天每节的文字, p: 统计表 }
 */
(function (TT) {
  'use strict';

  function decodeHtml(buf) {
    const u8 = new Uint8Array(buf);
    const probe = new TextDecoder('latin1').decode(u8.slice(0, 4096));
    const m = probe.match(/charset\s*=\s*["']?([\w-]+)/i);
    let cs = (m ? m[1] : 'utf-8').toLowerCase();
    if (cs === 'gb2312' || cs === 'gbk') cs = 'gb18030';
    try {
      return new TextDecoder(cs).decode(u8);
    } catch (e) {
      return new TextDecoder('utf-8').decode(u8);
    }
  }
  const texts = (el) => {
    const out = [];
    const w = document.createTreeWalker(el, NodeFilter.SHOW_TEXT);
    let n;
    while ((n = w.nextNode())) {
      const s = n.nodeValue.replace(/\s+/g, ' ').trim();
      if (s) out.push(s);
    }
    return out;
  };
  const isRecess = (td) => td.classList.contains('print-tt__recess');
  function parseTimetable(html) {
    const doc = new DOMParser().parseFromString(html, 'text/html');
    const areas = [...doc.querySelectorAll('div.printarea')];
    if (!areas.length) throw new Error('文件里找不到课表（没有 printarea）。请确认是从 eSchool 列印页面另存的 HTML。');
    const firstLabel = (areas[0].querySelector('.print-tt__info')?.textContent || '').split(/[：:]/)[0].trim();
    const kind = firstLabel === '场地' ? 'venue' : 'class';
    const sem = (html.match(/[?&]sem=(\d+)/) || [])[1];
    let periods,
      recess,
      times,
      pre,
      printed = '',
      note = '';
    const classes = areas.map((pa, idx) => {
      const infos = [...pa.querySelectorAll('.print-tt__info')].map((e) => e.textContent.trim());
      const val = (s) => (s || '').split(/[：:]/).slice(1).join('：').trim();
      const tbl = pa.querySelector('.print-tt__timetable-box table');
      const rows = [...tbl.querySelectorAll('tr')];
      if (!periods) {
        const h0 = [...rows[0].children].slice(3);
        periods = [];
        recess = [];
        h0.forEach((td) => {
          if (isRecess(td)) recess.push(periods.length);
          else periods.push(td.textContent.replace(/\s+/g, ''));
        });
        const h1 = [...rows[1].children];
        pre = h1.slice(1, 4).map((td) => texts(td).join(''));
        times = h1
          .slice(4)
          .filter((td) => !isRecess(td))
          .map((td) => texts(td));
        printed = infos[2] || '';
        const box = pa.querySelector('.print-tt__rmk-box');
        if (box) {
          const groups = [];
          let g = null;
          box.querySelectorAll('tr').forEach((tr) => {
            const tds = [...tr.children];
            if (tds.length > 1) {
              g = { who: tds[1].textContent.trim(), lines: [tds[0].textContent.trim()] };
              groups.push(g);
            } else if (g && tds[0]) g.lines.push(tds[0].textContent.trim());
          });
          const clean = (s) =>
            s
              .replace(/[A-Za-z]+\s*:\s*/g, '')
              .replace(/~/g, '–')
              .replace(/\s+/g, ' ');
          const title = (texts(box)[0] || '').replace(/\s*[A-Za-z0-9 ]*period\s*$/i, '');
          if (groups.length)
            note =
              `${title || '*'}：` + groups.map((x) => `${x.who} ${x.lines.map(clean).join('，')}`).join('；') + '。';
        }
      }
      const grid = rows
        .slice(2)
        .map((tr) =>
          [...tr.children]
            .slice(1)
            .filter((td) => !isRecess(td) && !td.hasAttribute('rowspan'))
            .map((td) => texts(td)),
        )
        .filter((r) => r.length);
      const p = [...pa.querySelectorAll('.print-tt__plan table')].map((t) =>
        [...t.querySelectorAll('tr')]
          .filter((tr) => tr.querySelector('td'))
          .map((tr) => [...tr.querySelectorAll('td')].map((td) => td.textContent.trim())),
      );
      return { n: val(infos[0]) || firstLabel + (idx + 1), hr: val(infos[1]), g: grid, p };
    });
    return { kind, periods, recess, times, pre, classes, printed, sem, note };
  }

  TT.parser = { decodeHtml, parseTimetable };
})((window.TT = window.TT || {}));
