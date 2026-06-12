package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const tvAppFamilyPageHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TV 安装包下载</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f3efe6;
      --panel: rgba(255, 252, 246, 0.9);
      --panel-strong: #fffaf1;
      --text: #211b14;
      --muted: #6e6253;
      --line: rgba(71, 54, 32, 0.12);
      --brand: #b14a18;
      --brand-dark: #8d3510;
      --brand-soft: rgba(177, 74, 24, 0.12);
      --success: #1f6b45;
      --warning: #8a5a09;
      --danger: #9a2d24;
      --shadow: 0 24px 60px rgba(58, 39, 17, 0.12);
      --radius-lg: 28px;
      --radius-md: 18px;
      --radius-sm: 12px;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      min-height: 100vh;
      font-family: "PingFang SC", "Hiragino Sans GB", "Noto Sans SC", sans-serif;
      background:
        radial-gradient(circle at top left, rgba(226, 171, 94, 0.26), transparent 32%),
        radial-gradient(circle at top right, rgba(177, 74, 24, 0.16), transparent 28%),
        linear-gradient(180deg, #f8f3ea 0%, var(--bg) 52%, #efe7dc 100%);
      color: var(--text);
    }

    a {
      color: inherit;
    }

    .shell {
      width: min(960px, calc(100vw - 32px));
      margin: 0 auto;
      padding: 32px 0 64px;
    }

    .hero,
    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: var(--radius-lg);
      box-shadow: var(--shadow);
      backdrop-filter: blur(16px);
    }

    .hero {
      padding: 28px;
      display: grid;
      gap: 20px;
    }

    .hero-top {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: flex-start;
    }

    .eyebrow {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: var(--warning);
      font-weight: 700;
    }

    h1 {
      margin: 10px 0 10px;
      font-size: clamp(28px, 5vw, 42px);
      line-height: 1.05;
    }

    .hero p,
    .hint,
    .empty,
    .status-note,
    .field-hint,
    .meta,
    .file-name {
      color: var(--muted);
    }

    .hero p {
      margin: 0;
      max-width: 640px;
      font-size: 15px;
      line-height: 1.7;
    }

    .status-chip {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 9px 14px;
      border-radius: 999px;
      background: var(--brand-soft);
      color: var(--brand-dark);
      font-size: 13px;
      font-weight: 700;
      white-space: nowrap;
    }

    .status-chip[data-state="ready"] {
      background: rgba(31, 107, 69, 0.12);
      color: var(--success);
    }

    .status-chip[data-state="error"] {
      background: rgba(154, 45, 36, 0.12);
      color: var(--danger);
    }

    .hero-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
    }

    .button,
    .download-link {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      min-height: 44px;
      padding: 0 18px;
      border-radius: 999px;
      border: 1px solid transparent;
      text-decoration: none;
      font-size: 14px;
      font-weight: 700;
      cursor: pointer;
      transition: transform 160ms ease, box-shadow 160ms ease, background-color 160ms ease, border-color 160ms ease;
    }

    .button:hover,
    .download-link:hover {
      transform: translateY(-1px);
    }

    .button-primary,
    .download-link {
      background: var(--brand);
      color: #fffaf3;
      box-shadow: 0 12px 24px rgba(177, 74, 24, 0.24);
    }

    .button-primary:hover,
    .download-link:hover {
      background: var(--brand-dark);
    }

    .button-secondary {
      background: rgba(255, 250, 243, 0.9);
      border-color: var(--line);
      color: var(--text);
    }

    .button-ghost {
      background: transparent;
      border-color: var(--line);
      color: var(--muted);
    }

    .button:disabled {
      cursor: wait;
      opacity: 0.7;
      transform: none;
    }

    .layout {
      margin-top: 20px;
      display: grid;
      grid-template-columns: minmax(0, 1fr);
      gap: 20px;
    }

    .panel {
      padding: 24px;
    }

    .panel-title {
      margin: 0 0 8px;
      font-size: 22px;
    }

    .panel-header {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: flex-start;
      margin-bottom: 18px;
    }

    .panel-header p {
      margin: 0;
      font-size: 14px;
      line-height: 1.7;
    }

    .field {
      display: grid;
      gap: 8px;
      margin-bottom: 16px;
    }

    .field:last-child {
      margin-bottom: 0;
    }

    label {
      font-size: 14px;
      font-weight: 700;
      color: var(--text);
    }

    input {
      width: 100%;
      min-height: 48px;
      padding: 0 14px;
      border-radius: var(--radius-sm);
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.92);
      color: var(--text);
      font: inherit;
    }

    input:focus {
      outline: 2px solid rgba(177, 74, 24, 0.22);
      outline-offset: 1px;
      border-color: rgba(177, 74, 24, 0.45);
    }

    .field-hint,
    .hint {
      font-size: 13px;
      line-height: 1.6;
    }

    .status-note {
      min-height: 22px;
      margin-top: 12px;
      font-size: 14px;
    }

    .status-note[data-tone="error"] {
      color: var(--danger);
    }

    .status-note[data-tone="success"] {
      color: var(--success);
    }

    .releases {
      display: grid;
      gap: 18px;
    }

    .release-card {
      background: var(--panel-strong);
      border: 1px solid var(--line);
      border-radius: 24px;
      padding: 20px;
      display: grid;
      gap: 18px;
    }

    .release-top {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: flex-start;
    }

    .release-title {
      margin: 0;
      font-size: 22px;
    }

    .release-subtitle {
      margin: 6px 0 0;
      font-size: 14px;
      color: var(--muted);
    }

    .badge-row,
    .download-list,
    .missing-list {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }

    .badge,
    .missing-pill {
      display: inline-flex;
      align-items: center;
      min-height: 32px;
      padding: 0 12px;
      border-radius: 999px;
      font-size: 13px;
      font-weight: 700;
    }

    .badge {
      background: rgba(55, 88, 130, 0.1);
      color: #36506f;
    }

    .badge-recommended {
      background: rgba(177, 74, 24, 0.16);
      color: var(--brand-dark);
    }

    .badge-complete {
      background: rgba(31, 107, 69, 0.12);
      color: var(--success);
    }

    .badge-missing {
      background: rgba(138, 90, 9, 0.12);
      color: var(--warning);
    }

    .release-notes {
      margin: 0;
      line-height: 1.75;
      white-space: pre-wrap;
    }

    .meta-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
    }

    .meta-box {
      border-radius: var(--radius-md);
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.74);
      padding: 14px;
      display: grid;
      gap: 6px;
    }

    .meta-label {
      font-size: 12px;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      color: var(--muted);
      font-weight: 700;
    }

    .meta-value {
      font-size: 16px;
      font-weight: 700;
      color: var(--text);
    }

    .download-item {
      min-width: min(100%, 260px);
      flex: 1 1 280px;
      border-radius: 20px;
      border: 1px solid rgba(177, 74, 24, 0.18);
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(251, 245, 236, 0.98));
      padding: 16px;
      display: grid;
      gap: 12px;
    }

    .download-item h4 {
      margin: 0;
      font-size: 16px;
    }

    .file-name {
      font-size: 12px;
      word-break: break-all;
      line-height: 1.5;
    }

    .empty {
      text-align: center;
      padding: 42px 18px;
      border: 1px dashed var(--line);
      border-radius: 24px;
      background: rgba(255, 255, 255, 0.55);
      line-height: 1.7;
    }

    .visually-hidden {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      border: 0;
    }

    @media (max-width: 720px) {
      .shell {
        width: min(100vw - 20px, 960px);
        padding-top: 20px;
        padding-bottom: 40px;
      }

      .hero,
      .panel {
        padding: 20px;
        border-radius: 24px;
      }

      .hero-top,
      .panel-header,
      .release-top {
        flex-direction: column;
      }

      .button,
      .download-link {
        width: 100%;
      }

      .hero-actions {
        flex-direction: column;
        align-items: stretch;
      }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="hero-top">
        <div>
          <span class="eyebrow">TV App Distribution</span>
          <h1>TV 安装包下载</h1>
          <p>登录后可查看最近三版 TV 安装包，按设备 ABI 直接下载。推荐版本会自动指向当前家庭可见范围里最高的版本号。</p>
        </div>
        <span class="status-chip" id="session-chip" data-state="idle">未登录</span>
      </div>
      <div class="hero-actions">
        <button type="button" class="button button-secondary" id="refresh-button">刷新列表</button>
        <button type="button" class="button button-ghost" id="logout-button">退出登录</button>
      </div>
      <div class="hint" id="top-hint">请使用现有账号密码登录后再下载。该页面不会展示管理功能。</div>
    </section>

    <div class="layout">
      <section class="panel" id="login-panel">
        <div class="panel-header">
          <div>
            <h2 class="panel-title">登录后查看下载列表</h2>
            <p>复用现有账号密码登录体系。支持用户名或邮箱登录；登录成功后页面会自动拉取最近三版安装包。</p>
          </div>
        </div>
        <form id="login-form">
          <div class="field">
            <label for="identity">用户名或邮箱</label>
            <input id="identity" name="identity" type="text" autocomplete="username" placeholder="请输入用户名或邮箱" required>
          </div>
          <div class="field">
            <label for="password">密码</label>
            <input id="password" name="password" type="password" autocomplete="current-password" placeholder="请输入密码" required>
          </div>
          <button type="submit" class="button button-primary" id="login-button">登录并查看下载</button>
          <div class="field-hint">下载页只接受已登录家庭成员账号访问，不提供匿名链接或分享口令。</div>
          <div class="status-note" id="login-status" aria-live="polite"></div>
        </form>
      </section>

      <section class="panel" id="releases-panel">
        <div class="panel-header">
          <div>
            <h2 class="panel-title">最近三版安装包</h2>
            <p>点击 ABI 按钮后会立即开始下载，不额外弹确认框。若某个版本缺少 ABI，会直接在卡片中标出。</p>
          </div>
        </div>
        <div id="releases-empty" class="empty">登录后将显示最近三版 TV 安装包。</div>
        <div id="releases-list" class="releases visually-hidden"></div>
      </section>
    </div>
  </main>

  <script>
    (function () {
      const ACCESS_TOKEN_KEY = 'tv_app_access_token';
      const REFRESH_TOKEN_KEY = 'tv_app_refresh_token';

      const loginPanel = document.getElementById('login-panel');
      const loginForm = document.getElementById('login-form');
      const identityInput = document.getElementById('identity');
      const passwordInput = document.getElementById('password');
      const loginButton = document.getElementById('login-button');
      const loginStatus = document.getElementById('login-status');
      const refreshButton = document.getElementById('refresh-button');
      const logoutButton = document.getElementById('logout-button');
      const releasesEmpty = document.getElementById('releases-empty');
      const releasesList = document.getElementById('releases-list');
      const sessionChip = document.getElementById('session-chip');
      const topHint = document.getElementById('top-hint');

      function getAccessToken() {
        return window.localStorage.getItem(ACCESS_TOKEN_KEY) || '';
      }

      function getRefreshToken() {
        return window.localStorage.getItem(REFRESH_TOKEN_KEY) || '';
      }

      function decodeBase64URL(segment) {
        const normalized = segment.replaceAll('-', '+').replaceAll('_', '/');
        const padding = normalized.length % 4 === 0 ? '' : '='.repeat(4 - (normalized.length % 4));
        return window.atob(normalized + padding);
      }

      function parseTokenPayload(token) {
        const parts = String(token || '').split('.');
        if (parts.length < 2) {
          return null;
        }
        try {
          return JSON.parse(decodeBase64URL(parts[1]));
        } catch (error) {
          return null;
        }
      }

      function tokenExpiresSoon(token) {
        const payload = parseTokenPayload(token);
        if (!payload || typeof payload.exp !== 'number') {
          return false;
        }
        return payload.exp * 1000 <= Date.now() + 30 * 1000;
      }

      function setTokens(accessToken, refreshToken) {
        if (accessToken) {
          window.localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
        }
        if (refreshToken) {
          window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
        }
      }

      function clearTokens() {
        window.localStorage.removeItem(ACCESS_TOKEN_KEY);
        window.localStorage.removeItem(REFRESH_TOKEN_KEY);
      }

      function setLoginBusy(busy) {
        loginButton.disabled = busy;
        loginButton.textContent = busy ? '登录中...' : '登录并查看下载';
      }

      function setRefreshBusy(busy) {
        refreshButton.disabled = busy;
        refreshButton.textContent = busy ? '刷新中...' : '刷新列表';
      }

      function setSessionState(state, text) {
        sessionChip.dataset.state = state;
        sessionChip.textContent = text;
      }

      function setLoginStatus(message, tone) {
        loginStatus.textContent = message || '';
        loginStatus.dataset.tone = tone || '';
      }

      function setTopHint(message) {
        topHint.textContent = message;
      }

      function isUnauthorized(payload) {
        return payload && payload.code === 401;
      }

      async function requestJSON(url, options) {
        const response = await fetch(url, options);
        let payload = null;
        try {
          payload = await response.json();
        } catch (error) {
          throw new Error('服务器返回了无法识别的响应');
        }
        if (!response.ok) {
          throw new Error(payload && payload.msg ? payload.msg : '请求失败');
        }
        if (payload.code !== 0) {
          const err = new Error(payload.msg || '请求失败');
          err.payload = payload;
          throw err;
        }
        return payload.data;
      }

      async function login(identity, password) {
        const payload = { password: password };
        if (identity.indexOf('@') >= 0) {
          payload.email = identity;
        } else {
          payload.username = identity;
        }
        return requestJSON('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
      }

      async function refreshAccessToken() {
        const refreshToken = getRefreshToken();
        if (!refreshToken) {
          throw new Error('登录已过期，请重新登录');
        }
        const data = await requestJSON('/api/v1/auth/refresh', {
          method: 'POST',
          headers: { Authorization: 'Bearer ' + refreshToken }
        });
        setTokens(data.access_token, data.refresh_token);
        return data.access_token;
      }

      async function ensureFreshAccessToken() {
        const accessToken = getAccessToken();
        if (accessToken && !tokenExpiresSoon(accessToken)) {
          return accessToken;
        }
        if (!getRefreshToken()) {
          if (accessToken) {
            return accessToken;
          }
          throw new Error('登录已过期，请重新登录');
        }
        return refreshAccessToken();
      }

      async function fetchReleasesWithRetry() {
        let accessToken = getAccessToken();
        if (!accessToken) {
          throw new Error('请先登录');
        }
        try {
          return await requestJSON('/api/v1/tv-app/releases', {
            headers: { Authorization: 'Bearer ' + accessToken }
          });
        } catch (error) {
          if (!(error && error.payload && isUnauthorized(error.payload))) {
            throw error;
          }
        }
        accessToken = await refreshAccessToken();
        return requestJSON('/api/v1/tv-app/releases', {
          headers: { Authorization: 'Bearer ' + accessToken }
        });
      }

      function escapeHTML(value) {
        return String(value || '')
          .replaceAll('&', '&amp;')
          .replaceAll('<', '&lt;')
          .replaceAll('>', '&gt;')
          .replaceAll('"', '&quot;');
      }

      function formatFileSize(size) {
        const value = Number(size || 0);
        if (!Number.isFinite(value) || value <= 0) {
          return '未知大小';
        }
        const units = ['B', 'KB', 'MB', 'GB'];
        let next = value;
        let unitIndex = 0;
        while (next >= 1024 && unitIndex < units.length - 1) {
          next /= 1024;
          unitIndex += 1;
        }
        const digits = next >= 100 || unitIndex === 0 ? 0 : 1;
        return next.toFixed(digits) + ' ' + units[unitIndex];
      }

      function formatDateTime(raw) {
        if (!raw) {
          return '未发布';
        }
        const date = new Date(raw);
        if (Number.isNaN(date.getTime())) {
          return raw;
        }
        return new Intl.DateTimeFormat('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        }).format(date);
      }

      function buildDownloadURL(releaseID, abi, accessToken) {
        return '/api/v1/tv-app/releases/' + encodeURIComponent(releaseID) + '/download/' + encodeURIComponent(abi) + '?access_token=' + encodeURIComponent(accessToken);
      }

      function buildBadges(item) {
        const parts = [];
        if (item.latest_recommended) {
          parts.push('<span class="badge badge-recommended">推荐安装</span>');
        }
        if (Array.isArray(item.missing_abis) && item.missing_abis.length === 0) {
          parts.push('<span class="badge badge-complete">ABI 已齐全</span>');
        } else if (Array.isArray(item.missing_abis) && item.missing_abis.length > 0) {
          parts.push('<span class="badge badge-missing">缺少 ABI</span>');
        }
        if (Array.isArray(item.uploaded_abis) && item.uploaded_abis.length > 0) {
          parts.push('<span class="badge">已上传：' + escapeHTML(item.uploaded_abis.join(' / ')) + '</span>');
        }
        return parts.join('');
      }

      function buildMissing(item) {
        if (!Array.isArray(item.missing_abis) || item.missing_abis.length === 0) {
          return '';
        }
        return '<div><div class="meta-label">仍缺少的 ABI</div><div class="missing-list">' + item.missing_abis.map(function (abi) {
          return '<span class="missing-pill">' + escapeHTML(abi) + '</span>';
        }).join('') + '</div></div>';
      }

      function buildDownloads(item) {
        if (!Array.isArray(item.abi_items) || item.abi_items.length === 0) {
          return '<div class="empty">当前版本还没有可下载的 APK 文件。</div>';
        }
        return '<div class="download-list">' + item.abi_items.map(function (abi) {
          const href = '/api/v1/tv-app/releases/' + encodeURIComponent(item.id) + '/download/' + encodeURIComponent(abi.abi);
          return [
            '<article class="download-item">',
            '<div>',
            '<h4>' + escapeHTML(abi.abi) + '</h4>',
            '<div class="file-name">' + escapeHTML(abi.file_name || '') + '</div>',
            '</div>',
            '<div class="meta">大小：' + escapeHTML(formatFileSize(abi.file_size)) + '</div>',
            '<div class="meta">最近更新：' + escapeHTML(formatDateTime(abi.updated_at)) + '</div>',
            '<a class="download-link" href="' + href + '" data-download-release-id="' + escapeHTML(item.id) + '" data-download-abi="' + escapeHTML(abi.abi) + '">下载 ' + escapeHTML(abi.abi) + '</a>',
            '</article>'
          ].join('');
        }).join('') + '</div>';
      }

      function renderReleases(items) {
        if (!Array.isArray(items) || items.length === 0) {
          releasesList.classList.add('visually-hidden');
          releasesList.innerHTML = '';
          releasesEmpty.classList.remove('visually-hidden');
          releasesEmpty.textContent = '当前暂无可下载 TV 安装包。请等待管理员发布后再试。';
          return;
        }
        releasesEmpty.classList.add('visually-hidden');
        releasesList.classList.remove('visually-hidden');
        releasesList.innerHTML = items.map(function (item) {
          return [
            '<article class="release-card">',
            '<div class="release-top">',
            '<div>',
            '<h3 class="release-title">v' + escapeHTML(item.version_name) + '</h3>',
            '<div class="release-subtitle">versionCode ' + escapeHTML(item.version_code) + '</div>',
            '</div>',
            '<div class="badge-row">' + buildBadges(item) + '</div>',
            '</div>',
            '<p class="release-notes">' + escapeHTML(item.release_notes || '暂无版本说明') + '</p>',
            '<div class="meta-grid">',
            '<div class="meta-box"><div class="meta-label">发布时间</div><div class="meta-value">' + escapeHTML(formatDateTime(item.published_at)) + '</div></div>',
            '<div class="meta-box"><div class="meta-label">已上传 ABI</div><div class="meta-value">' + escapeHTML((item.uploaded_abis || []).join(' / ') || '暂无') + '</div></div>',
            '<div class="meta-box"><div class="meta-label">包名</div><div class="meta-value">' + escapeHTML(item.package_name || '-') + '</div></div>',
            '</div>',
            buildMissing(item),
            buildDownloads(item),
            '</article>'
          ].join('');
        }).join('');
      }

      function applyLoggedInUI() {
        loginPanel.classList.add('visually-hidden');
        setSessionState('ready', '已登录');
        setTopHint('当前展示最近三版 TV 安装包。点击 ABI 按钮会直接开始下载。');
      }

      function applyLoggedOutUI() {
        loginPanel.classList.remove('visually-hidden');
        setSessionState('idle', '未登录');
        setTopHint('请使用现有账号密码登录后再下载。该页面不会展示管理功能。');
        releasesList.classList.add('visually-hidden');
        releasesList.innerHTML = '';
        releasesEmpty.classList.remove('visually-hidden');
        releasesEmpty.textContent = '登录后将显示最近三版 TV 安装包。';
      }

      async function loadReleases() {
        setRefreshBusy(true);
        setSessionState('idle', '读取中');
        try {
          const data = await fetchReleasesWithRetry();
          applyLoggedInUI();
          setLoginStatus('', '');
          renderReleases(data.items || []);
          setSessionState('ready', '已登录');
        } catch (error) {
          const message = error && error.message ? error.message : '读取 TV 安装包列表失败';
          if (message.indexOf('登录') >= 0 || message.indexOf('token') >= 0 || message.indexOf('authorization') >= 0) {
            clearTokens();
            applyLoggedOutUI();
            setLoginStatus(message, 'error');
            setSessionState('error', '需要登录');
          } else {
            setSessionState('error', '读取失败');
            releasesEmpty.classList.remove('visually-hidden');
            releasesEmpty.textContent = message;
          }
        } finally {
          setRefreshBusy(false);
        }
      }

      loginForm.addEventListener('submit', async function (event) {
        event.preventDefault();
        const identity = identityInput.value.trim();
        const password = passwordInput.value;
        if (!identity || !password) {
          setLoginStatus('请输入用户名或邮箱，以及密码。', 'error');
          return;
        }
        setLoginBusy(true);
        setLoginStatus('', '');
        try {
          const data = await login(identity, password);
          setTokens(data.access_token, data.refresh_token);
          passwordInput.value = '';
          setLoginStatus('登录成功，正在载入安装包列表。', 'success');
          await loadReleases();
        } catch (error) {
          clearTokens();
          setLoginStatus(error && error.message ? error.message : '登录失败，请稍后再试。', 'error');
          setSessionState('error', '登录失败');
        } finally {
          setLoginBusy(false);
        }
      });

      refreshButton.addEventListener('click', function () {
        if (!getAccessToken()) {
          setLoginStatus('请先登录后再刷新。', 'error');
          return;
        }
        loadReleases();
      });

      releasesList.addEventListener('click', async function (event) {
        const link = event.target.closest('[data-download-release-id][data-download-abi]');
        if (!link) {
          return;
        }
        event.preventDefault();
        try {
          const accessToken = await ensureFreshAccessToken();
          window.location.assign(buildDownloadURL(link.dataset.downloadReleaseId, link.dataset.downloadAbi, accessToken));
        } catch (error) {
          clearTokens();
          applyLoggedOutUI();
          setLoginStatus(error && error.message ? error.message : '登录已过期，请重新登录。', 'error');
          setSessionState('error', '需要登录');
        }
      });

      logoutButton.addEventListener('click', function () {
        clearTokens();
        setLoginStatus('已退出本页面登录状态。', 'success');
        applyLoggedOutUI();
      });

      if (getAccessToken()) {
        loadReleases();
      } else {
        applyLoggedOutUI();
      }
    })();
  </script>
</body>
</html>
`

func mountTVAppFamilyPage(r *gin.Engine) {
	serve := func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(tvAppFamilyPageHTML))
	}
	r.GET("/tv-app", serve)
	r.HEAD("/tv-app", serve)
	r.GET("/tv-app/", serve)
	r.HEAD("/tv-app/", serve)
}
