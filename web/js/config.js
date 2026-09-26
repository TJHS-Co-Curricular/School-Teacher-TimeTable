/**
 * 循人课表 — 网页默认设置
 *
 * 由 Teacher-TimeTable.exe 打开时，会再读取运行时文件夹的 config/app.json 覆盖这里的值
 * （学校名称、GitHub 网址等请改 config/app.json，不用重新打包）。
 * 直接双击 index.html 打开时，只会用这里的默认值。
 */
window.TT = window.TT || {};

TT.config = {
  /** 版本号（由 exe 的 /config.json 提供） */
  version: '',
  /** 浏览器分页标题 */
  appName: '循人课表',
  /** 顶栏标题 */
  headerTitle: '循人中学 · 课表',
  /** 页脚学校名称 */
  schoolName: '循人中学',
  schoolNameEn: 'Tsun Jin High School',
  /** GitHub 项目网址（留空则隐藏页脚的 GitHub 按钮） */
  githubUrl: 'https://github.com/TJHS-Co-Curricular/School-Teacher-TimeTable',
  /** 依序尝试读取的课表来源（source*.html 由 Teacher-TimeTable.exe 提供） */
  sources: {
    /** 教务处「班级课表」（必需） */
    class: ['source.html', '班级课表.html'],
    /** 英文系统「场地课表_English」（可选，有就合并） */
    english: ['source_en.html', '场地课表_English.html'],
  },
  /** 心跳间隔（毫秒）：告诉 exe 页面仍开着 */
  heartbeatMs: 20000,
};
