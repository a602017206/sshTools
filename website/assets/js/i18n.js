export const LANG_STORAGE_KEY = 'sshtools-site-lang';

export const MESSAGES = {
  zh: {
    brand: 'AHaSSHTools',
    tagline: '跨平台运维客户端：SSH、文件、监控、数据与 AI',
    'nav.home': '首页',
    'nav.features': '功能',
    'nav.download': '下载',
    'nav.github': 'GitHub',
    'nav.theme': '主题',
    'nav.lang': '语言',
    'theme.open': '主题',
    'theme.mode': '组合模型',
    'theme.mode.preset': '预设',
    'theme.mode.duo': '双色',
    'theme.mode.trio': '三色',
    'theme.presets': '预设主题',
    'theme.bg': '背景色',
    'theme.fg': '字体色',
    'theme.accent': '强调色',
    'theme.cursor': '鼠标造型',
    'home.ctaFeatures': '查看功能',
    'home.ctaDownload': '立即下载',
    'home.capabilitiesTitle': '一套客户端，覆盖日常运维链路',
    'cap.ssh.title': 'SSH 与终端',
    'cap.ssh.desc': '多标签 SSH、本地 Shell、密钥认证与会话日志，集中处理主机资产。',
    'cap.sftp.title': 'SFTP 文件管理',
    'cap.sftp.desc': '浏览、上传、下载、编辑远程文件，并跟踪每一项传输进度。',
    'cap.monitor.title': '系统监控',
    'cap.monitor.desc': '按会话查看 CPU、内存、磁盘、网络和主机信息，按需实时刷新。',
    'cap.data.title': '数据服务',
    'cap.data.desc': '连接关系库、缓存、搜索和消息服务，在同一工作台完成浏览与查询。',
    'cap.ai.title': 'AI 助手',
    'cap.ai.desc': '结合当前 SSH 和数据上下文辅助排查；变更操作始终需要确认。',
    'cap.mobile.title': '桌面与移动端',
    'cap.mobile.desc': '桌面端与 Android、iOS 共用 Go 核心，让工具随运维现场同行。',
    'home.platformsTitle': '面向桌面与移动运维现场',
    'home.showcaseTitle': '场景速览',
    'home.showcaseHint': '自动轮播 · 悬停可暂停',
    'home.terminalTitle': '会话预演',
    'carousel.prev': '上一张',
    'carousel.next': '下一张',
    'slide.ssh.kicker': '终端',
    'slide.ssh.title': '多机会话，一眼切换',
    'slide.ssh.body': '标签页、批量执行与会话日志，让多机运维保持同一节奏。',
    'slide.files.kicker': '文件',
    'slide.files.title': 'SFTP 与远程编辑同屏',
    'slide.files.body': '上传下载带进度，远程文本可在线改完直接保存。',
    'slide.data.kicker': '数据',
    'slide.data.title': '库表 · 缓存 · 搜索 · 消息',
    'slide.data.body': '关系库与中间件入口收拢到同一工作台，减少工具来回跳。',
    'slide.ai.kicker': 'AI',
    'slide.ai.title': '带着上下文排查',
    'slide.ai.body': '结合当前会话与对象上下文辅助分析，变更仍需你确认。',
    'features.title': '功能概览',
    'features.ssh.title': 'SSH 与终端',
    'features.ssh.body': '多标签会话、密码与 SSH 密钥认证、终端编码、主题、会话日志和常用命令提示。',
    'features.sftp.title': '文件管理（SFTP）',
    'features.sftp.body': '目录浏览、收藏路径、上传下载、冲突处理和远程文本编辑，日常文件操作不用离开工作台。',
    'features.monitor.title': '系统监控',
    'features.monitor.body': '查看 CPU、内存、磁盘、网络与主机信息；每个会话独立开启实时采集。',
    'features.data.title': '数据服务',
    'features.data.body': '通过 JDBC 与原生客户端连接关系库、缓存、搜索和消息系统，支持对象浏览与查询。',
    'features.ai.title': 'AI 助手（Copilot）',
    'features.ai.body': '接入 DeepSeek、OpenAI、Ollama 或自定义服务商，基于当前上下文辅助只读查询和确认后的变更。',
    'features.mobile.title': '跨平台客户端',
    'features.mobile.body': '桌面端基于 Go、Wails 和 Svelte；Android 与 iOS 使用同一 Go 后端能力。',
    'features.devtools.title': '本地开发工具',
    'features.devtools.body': '内置 JSON、Base64、URL 编解码、哈希、AES、SM4、时间戳与 UUID 工具。',
    'maturity.full': '完整',
    'maturity.ready': '可用',
    'maturity.early': '雏形',
    'download.title': '下载 AHaSSHTools',
    'download.noteMac': '若 macOS 提示应用无法打开，请执行 <code>xattr -cr AHaSSHTools.app</code> 后重新打开。',
    'download.noteAndroid': 'Android 下载 APK 后可直接安装；大多数真机请选择 ARM64 包。',
    'download.noteIos': 'iOS 提供未签名 IPA，需使用 AltStore、Sideloadly 等工具配合 Apple ID 自签侧载。',
    'download.platform.windows': 'Windows',
    'download.platform.macosArm': 'macOS（Apple Silicon）',
    'download.platform.macosIntel': 'macOS（Intel）',
    'download.platform.androidArm64': 'Android（ARM64，主流真机）',
    'download.platform.androidArmv7': 'Android（ARMv7）',
    'download.platform.androidX64': 'Android（x86_64）',
    'download.platform.ios': 'iOS（未签名 IPA）',
  },
  en: {
    brand: 'AHaSSHTools',
    tagline: 'Cross-platform ops client for SSH, files, monitoring, data, and AI',
    'nav.home': 'Home',
    'nav.features': 'Features',
    'nav.download': 'Download',
    'nav.github': 'GitHub',
    'nav.theme': 'Theme',
    'nav.lang': 'Language',
    'theme.open': 'Theme',
    'theme.mode': 'Combination',
    'theme.mode.preset': 'Preset',
    'theme.mode.duo': 'Duo',
    'theme.mode.trio': 'Trio',
    'theme.presets': 'Presets',
    'theme.bg': 'Background',
    'theme.fg': 'Text',
    'theme.accent': 'Accent',
    'theme.cursor': 'Cursor',
    'home.ctaFeatures': 'Explore features',
    'home.ctaDownload': 'Download now',
    'home.capabilitiesTitle': 'One client for the everyday operations workflow',
    'cap.ssh.title': 'SSH and terminal',
    'cap.ssh.desc': 'Manage host assets with tabbed SSH, local shells, key authentication, and session logs.',
    'cap.sftp.title': 'SFTP file management',
    'cap.sftp.desc': 'Browse, transfer, and edit remote files while tracking every transfer.',
    'cap.monitor.title': 'System monitoring',
    'cap.monitor.desc': 'Inspect CPU, memory, disks, network, and host details per session, live when needed.',
    'cap.data.title': 'Data services',
    'cap.data.desc': 'Browse and query relational, cache, search, and messaging systems from one workspace.',
    'cap.ai.title': 'AI assistant',
    'cap.ai.desc': 'Investigate with SSH and data context; every change still requires confirmation.',
    'cap.mobile.title': 'Desktop and mobile',
    'cap.mobile.desc': 'A shared Go core brings the same operations tools to desktop, Android, and iOS.',
    'home.platformsTitle': 'Built for desktop and mobile operations',
    'home.showcaseTitle': 'Scene carousel',
    'home.showcaseHint': 'Autoplay · hover to pause',
    'home.terminalTitle': 'Session preview',
    'carousel.prev': 'Previous',
    'carousel.next': 'Next',
    'slide.ssh.kicker': 'Terminal',
    'slide.ssh.title': 'Switch hosts without losing flow',
    'slide.ssh.body': 'Tabs, batch execute, and session logs keep multi-host work in sync.',
    'slide.files.kicker': 'Files',
    'slide.files.title': 'SFTP and remote editing together',
    'slide.files.body': 'Transfers show progress, and remote text edits save in place.',
    'slide.data.kicker': 'Data',
    'slide.data.title': 'SQL · cache · search · messaging',
    'slide.data.body': 'Bring databases and middleware into one workspace instead of juggling apps.',
    'slide.ai.kicker': 'AI',
    'slide.ai.title': 'Investigate with live context',
    'slide.ai.body': 'Use session and object context for assistance; changes still need confirmation.',
    'features.title': 'Feature overview',
    'features.ssh.title': 'SSH and terminal',
    'features.ssh.body': 'Tabbed sessions, password and SSH-key authentication, terminal encodings, themes, session logs, and command suggestions.',
    'features.sftp.title': 'File management (SFTP)',
    'features.sftp.body': 'Browse directories, save favorite paths, transfer files, handle conflicts, and edit remote text without leaving the workspace.',
    'features.monitor.title': 'System monitoring',
    'features.monitor.body': 'View CPU, memory, disks, network, and host details, with live collection controlled per session.',
    'features.data.title': 'Data services',
    'features.data.body': 'Connect relational databases, caches, search, and messaging systems through JDBC and native clients for browsing and queries.',
    'features.ai.title': 'AI assistant (Copilot)',
    'features.ai.body': 'Use DeepSeek, OpenAI, Ollama, or custom providers for context-aware investigation and confirmed changes.',
    'features.mobile.title': 'Cross-platform clients',
    'features.mobile.body': 'The desktop client uses Go, Wails, and Svelte; Android and iOS share the same Go backend capabilities.',
    'features.devtools.title': 'Local developer tools',
    'features.devtools.body': 'Built-in JSON, Base64, URL encoding, hashing, AES, SM4, timestamp, and UUID tools.',
    'maturity.full': 'Full',
    'maturity.ready': 'Ready',
    'maturity.early': 'Early',
    'download.title': 'Download AHaSSHTools',
    'download.noteMac': 'If macOS cannot open the app, run <code>xattr -cr AHaSSHTools.app</code> and open it again.',
    'download.noteAndroid': 'Install the downloaded APK directly on Android; ARM64 is right for most physical devices.',
    'download.noteIos': 'The iOS IPA is unsigned. Sideload it with an Apple ID using tools such as AltStore or Sideloadly.',
    'download.platform.windows': 'Windows',
    'download.platform.macosArm': 'macOS (Apple Silicon)',
    'download.platform.macosIntel': 'macOS (Intel)',
    'download.platform.androidArm64': 'Android (ARM64, most devices)',
    'download.platform.androidArmv7': 'Android (ARMv7)',
    'download.platform.androidX64': 'Android (x86_64)',
    'download.platform.ios': 'iOS (unsigned IPA)',
  },
};

export function normalizeLang(lang) {
  return lang === 'en' ? 'en' : 'zh';
}

export function readStoredLang(storage = globalThis.localStorage) {
  try {
    return normalizeLang(storage?.getItem?.(LANG_STORAGE_KEY));
  } catch {
    return 'zh';
  }
}

export function setStoredLang(lang, storage = globalThis.localStorage) {
  const value = normalizeLang(lang);
  try {
    storage?.setItem?.(LANG_STORAGE_KEY, value);
  } catch { /* ignore unavailable storage */ }
  return value;
}

export function t(dict, lang, key) {
  const table = dict[normalizeLang(lang)] || dict.zh || {};
  return table[key] ?? key;
}

export function applyI18n(root, dict, lang) {
  const normalizedLang = normalizeLang(lang);
  root.querySelectorAll('[data-i18n]').forEach((node) => {
    const key = node.getAttribute('data-i18n');
    if (key) node.textContent = t(dict, normalizedLang, key);
  });
  root.querySelectorAll('[data-i18n-html]').forEach((node) => {
    const key = node.getAttribute('data-i18n-html');
    if (key) node.innerHTML = t(dict, normalizedLang, key);
  });
  root.documentElement?.setAttribute?.('lang', normalizedLang === 'zh' ? 'zh-CN' : 'en');
}
