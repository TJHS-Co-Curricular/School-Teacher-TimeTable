/**
 * 循人课表 — 共用小工具
 */
(function (TT) {
  'use strict';

  /** HTML 转义 */
  const esc = (s) =>
    String(s).replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);

  /** 中文按拼音排序 */
  const zh = (a, b) => a.localeCompare(b, 'zh-Hans-CN');

  /** document.getElementById 简写 */
  const $ = (id) => document.getElementById(id);

  /** 时间转 24 小时制：「01:20 pm」→「13:20」 */
  const to24 = (s) => {
    const m = s.match(/(\d+):(\d+)\s*(am|pm)/i);
    if (!m) return s;
    let h = +m[1];
    if (/pm/i.test(m[3]) && h < 12) h += 12;
    return String(h).padStart(2, '0') + ':' + m[2];
  };

  TT.util = { esc, zh, $, to24 };
})((window.TT = window.TT || {}));
