//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"CUMT-autologin/internal/config"

	"github.com/getlantern/systray"
	webview "github.com/webview/webview_go"
	"golang.org/x/sys/windows"
)

type viewConfig struct {
	SSID              string `json:"ssid"`
	StudentID         string `json:"student_id"`
	Password          string `json:"password"`
	Operator          string `json:"operator"`
	LoginMode         string `json:"login_mode"`
	AutoLoginInterval int    `json:"auto_login_interval"`
	AutoStart         bool   `json:"auto_start"`
	OpenSettingsOnRun bool   `json:"open_settings_on_run"`
	BuildInfo         string `json:"build_info"`
}

var (
	settingsMu sync.Mutex
	settingsW  webview.WebView
)

const settingsHTML = `
<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<title>CUMT Autologin 设置</title>
<style>
:root {
  --bg: #020617;
  --card: #020617;
  --card-border: #1e293b;
  --accent: #22c55e;
  --accent-soft: rgba(34,197,94,0.12);
  --danger: #ef4444;
  --danger-soft: rgba(239,68,68,0.12);
  --text: #e5e7eb;
  --sub: #9ca3af;
  --muted: #111827;
  --input: #020617;
  --radius-lg: 18px;
  --radius-sm: 999px;
  --shadow-soft: 0 18px 45px rgba(15,23,42,0.55);
  --gap: 14px;
}

*,
*::before,
*::after {
  box-sizing: border-box;
}

html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  background: var(--bg);
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  color: var(--text);
}

.app {
  height: 100%;
  padding: 14px 14px 16px 14px;
}

.card {
  height: 100%;
  border-radius: var(--radius-lg);
  border: 1px solid var(--card-border);
  background:
     radial-gradient(circle at top left, rgba(56,189,248,0.10), transparent 55%),
     radial-gradient(circle at bottom right, rgba(34,197,94,0.10), transparent 55%),
     var(--card);
  box-shadow: var(--shadow-soft);
  padding: 16px 18px 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.app-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo {
  width: 28px;
  height: 28px;
  border-radius: 9px;
  background: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 15px;
  color: var(--accent);
  box-shadow: 0 0 0 1px rgba(34,197,94,0.4);
}

.app-meta {
  display: flex;
  flex-direction: column;
}

.app-meta-main {
  font-size: 14px;
  font-weight: 600;
}

.app-meta-sub {
  font-size: 11px;
  color: var(--sub);
}

.window-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--sub);
}

.window-actions span {
  opacity: 0.75;
}

.chip-mini {
  border-radius: 999px;
  padding: 2px 7px;
  background: rgba(15,23,42,0.9);
  border: 1px solid rgba(148,163,184,0.6);
  font-size: 10px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.chip-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow: 0 0 0 2px rgba(34,197,94,0.35);
}

.main-grid {
  display: grid;
  grid-template-columns: 1.05fr 1.05fr;
  gap: 14px;
  flex: 1;
  min-height: 0;
}

.section {
  border-radius: 14px;
  background: rgba(15,23,42,0.92);
  border: 1px solid rgba(30,64,175,0.7);
  padding: 12px 12px 11px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
}

.section-sub {
  font-size: 11px;
  color: var(--sub);
}

.badge-soft {
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 10px;
  border: 1px solid rgba(148,163,184,0.6);
  color: var(--sub);
  background: rgba(15,23,42,0.9);
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: var(--sub);
}

.label-main {
  font-size: 11px;
  font-weight: 500;
  color: #cbd5f5;
}

.hint-tag {
  font-size: 10px;
  padding: 2px 7px;
  border-radius: 999px;
  border: 1px solid rgba(34,197,94,0.35);
  background: var(--accent-soft);
  color: #bbf7d0;
}

.input {
  border-radius: 9px;
  padding: 6px 10px;
  border: 1px solid #1e293b;
  background: rgba(15,23,42,0.96);
  color: var(--text);
  font-size: 12px;
  outline: none;
}

.input:focus {
  border-color: #38bdf8;
  box-shadow: 0 0 0 1px rgba(56,189,248,0.7);
}

.input::placeholder {
  color: #6b7280;
}

.pill-row {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.pill {
  border-radius: 999px;
  padding: 3px 9px 4px 9px;
  font-size: 11px;
  border: 1px solid rgba(148,163,184,0.6);
  color: var(--sub);
  background: rgba(15,23,42,0.96);
  cursor: pointer;
  user-select: none;
}

.pill.active {
  border-color: rgba(34,197,94,0.75);
  background: var(--accent-soft);
  color: #bbf7d0;
}

.pill.danger {
  border-color: rgba(239,68,68,0.7);
  background: var(--danger-soft);
  color: #fecaca;
}

.button-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.btn {
  border-radius: 999px;
  padding: 6px 16px;
  font-size: 12px;
  border: 1px solid #1f2937;
  background: rgba(15,23,42,0.96);
  color: var(--text);
  cursor: pointer;
}

.btn-primary {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  border-color: transparent;
  color: #022c22;
  font-weight: 600;
}

.btn-primary:hover {
  filter: brightness(1.05);
}

.btn-ghost {
  background: rgba(15,23,42,0.96);
}

.btn-ghost:hover {
  background: #020617;
}

.btn-danger {
  border-color: rgba(239,68,68,0.8);
  background: var(--danger-soft);
  color: #fecaca;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 11px;
}

.switch-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.switch-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 6px;
  border-radius: 9px;
  background: rgba(15,23,42,0.96);
  border: 1px solid #1f2937;
}

.switch-label {
  display: flex;
  flex-direction: column;
}

.switch-title {
  font-size: 12px;
}

.switch-desc {
  font-size: 10px;
  color: var(--sub);
}

.switch {
  width: 34px;
  height: 18px;
  border-radius: 999px;
  background: #111827;
  border: 1px solid #4b5563;
  position: relative;
  cursor: pointer;
  transition: all 0.15s ease;
}

.switch-knob {
  position: absolute;
  top: 1px;
  left: 1px;
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: #9ca3af;
  transition: all 0.15s ease;
}

.switch.on {
  background: #22c55e;
  border-color: #22c55e;
}

.switch.on .switch-knob {
  left: 17px;
  background: #ecfdf5;
}

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: var(--sub);
  padding-top: 3px;
  border-top: 1px dashed rgba(55,65,81,0.8);
}

.footer-left, .footer-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34,197,94,0.35);
}

.status-text-ok {
  color: #4ade80;
}
.status-text-warn {
  color: #f97316;
}
</style>
</head>
<body>
<div class="app">
  <div class="card">
    <header class="card-header">
      <div class="app-title">
        <div class="logo">Wi</div>
        <div class="app-meta">
          <div class="app-meta-main">CUMT Autologin</div>
          <div class="app-meta-sub">校园网自动登录 · 设置面板</div>
        </div>
      </div>
      <div class="window-actions">
        <div class="chip-mini">
          <span class="chip-dot"></span>
          <span id="statusText">未检测</span>
        </div>
      </div>
    </header>

    <main class="main-grid">
      <!-- 左侧：网络 / 账号 -->
      <section class="section">
        <div class="section-header">
          <div>
            <div class="section-title">网络与账号</div>
            <div class="section-sub">当前 WiFi & 登录方式</div>
          </div>
          <span class="badge-soft">仅支持 CUMT_Stu</span>
        </div>

        <div class="field-group">
          <div class="label-row">
            <span class="label-main">WiFi 名称 (SSID)</span>
            <span class="hint-tag">自动登录仅在匹配时生效</span>
          </div>
          <input id="ssidInput" class="input" placeholder="CUMT_Stu" />
        </div>

        <div class="field-group">
          <div class="label-row">
            <span class="label-main">学号</span>
          </div>
          <input id="sidInput" class="input" placeholder="0823xxxx" />
        </div>

        <div class="field-group">
          <div class="label-row">
            <span class="label-main">登录密码</span>
          </div>
          <input id="pwdInput" class="input" type="password" placeholder="••••••••" />
        </div>

        <div class="field-group">
          <div class="label-row">
            <span class="label-main">运营商后缀</span>
          </div>
          <div class="pill-row" id="operatorPills">
            <div class="pill" data-op="telecom">@telecom（电信）</div>
            <div class="pill" data-op="cmcc">@cmcc（移动）</div>
            <div class="pill" data-op="unicom">@unicom（联通）</div>
            <div class="pill" data-op="none">无运营商后缀</div>
          </div>
        </div>

        <div class="field-group">
          <div class="label-row">
            <span class="label-main">登录模式</span>
          </div>
          <div class="pill-row" id="loginModePills">
            <div class="pill" data-mode="operator_id">运营商账号（学号+后缀）</div>
            <div class="pill" data-mode="campus_only">校园网账号（纯学号）</div>
          </div>
        </div>
      </section>

      <!-- 右侧：行为 / 操作 -->
      <section class="section">
        <div class="section-header">
          <div>
            <div class="section-title">行为与操作</div>
            <div class="section-sub">后台行为 · 立即控制</div>
          </div>
          <span class="badge-soft">win 后台小助手</span>
        </div>

        <div class="switch-row">
          <div class="switch-item" id="autoStartItem">
            <div class="switch-label">
              <div class="switch-title">开机自启</div>
              <div class="switch-desc">登录当前 Windows 用户时自动启动 CUMT Autologin</div>
            </div>
            <div class="switch" id="autoStartSwitch">
              <div class="switch-knob"></div>
            </div>
          </div>

          <div class="switch-item" id="openOnRunItem">
            <div class="switch-label">
              <div class="switch-title">启动时自动打开设置面板</div>
              <div class="switch-desc">适合调试；关闭后仅在托盘中驻留</div>
            </div>
            <div class="switch on" id="openOnRunSwitch">
              <div class="switch-knob"></div>
            </div>
          </div>

          <div class="switch-item">
            <div class="switch-label">
              <div class="switch-title">自动重试间隔</div>
              <div class="switch-desc">检测断网后尝试登录的时间间隔（秒）</div>
            </div>
            <input id="intervalInput" class="input" style="width:64px;text-align:center;" />
          </div>
        </div>

        <div class="button-row">
          <button class="btn btn-primary" id="btnLoginNow">立即尝试登录</button>
          <button class="btn btn-ghost btn-sm" id="btnLogout">注销会话</button>
          <button class="btn btn-ghost btn-sm" id="btnOpenConfig">打开 config.yaml</button>
          <button class="btn btn-danger btn-sm" id="btnQuit">退出程序</button>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="footer-left">
        <span class="dot"></span>
        <span id="footerStatus">后台循环检测中…</span>
      </div>
      <div class="footer-right">
        <span>托盘：右键图标可快速操作</span>
        <span>构建号：<span id="buildInfo"></span></span>
      </div>
    </footer>
  </div>
</div>

<script>
  const state = {
    operator: "telecom",
    loginMode: "operator_id",
    autoStart: false,
    openOnRun: true,
  };

  function applyPills(groupId, activeValue, attr) {
    const pills = document.querySelectorAll('#' + groupId + ' .pill');
    pills.forEach(p => {
      const v = p.getAttribute(attr);
      p.classList.toggle('active', v === activeValue);
    });
  }

  function toggleSwitch(el, on) {
    if (on) el.classList.add('on');
    else el.classList.remove('on');
  }

  async function loadConfig() {
    if (!window.goGetConfig) return;
    try {
      const cfg = await window.goGetConfig();
      document.getElementById('ssidInput').value = cfg.ssid || "";
      document.getElementById('sidInput').value = cfg.student_id || "";
      document.getElementById('pwdInput').value = cfg.password || "";
      document.getElementById('intervalInput').value = cfg.auto_login_interval || 10;

      state.operator = cfg.operator || "telecom";
      state.loginMode = cfg.login_mode || "operator_id";
      state.autoStart = !!cfg.auto_start;
      state.openOnRun = cfg.open_settings_on_run !== false;

      applyPills('operatorPills', state.operator, 'data-op');
      applyPills('loginModePills', state.loginMode, 'data-mode');
      toggleSwitch(document.getElementById('autoStartSwitch'), state.autoStart);
      toggleSwitch(document.getElementById('openOnRunSwitch'), state.openOnRun);

      if (cfg.build_info) {
        document.getElementById('buildInfo').innerText = cfg.build_info;
      }
    } catch (e) {
      console.error(e);
    }
  }

  function collectConfig() {
    return {
      ssid: document.getElementById('ssidInput').value.trim(),
      student_id: document.getElementById('sidInput').value.trim(),
      password: document.getElementById('pwdInput').value,
      operator: state.operator,
      login_mode: state.loginMode,
      auto_login_interval: parseInt(document.getElementById('intervalInput').value || "10", 10),
      auto_start: state.autoStart,
      open_settings_on_run: state.openOnRun,
    };
  }

  async function saveConfig() {
    if (!window.goSaveConfig) return;
    try {
      await window.goSaveConfig(collectConfig());
    } catch (e) {
      console.error(e);
    }
  }

  function bindEvents() {
    document.getElementById('operatorPills').addEventListener('click', e => {
      const pill = e.target.closest('.pill');
      if (!pill) return;
      state.operator = pill.getAttribute('data-op');
      applyPills('operatorPills', state.operator, 'data-op');
      saveConfig();
    });

    document.getElementById('loginModePills').addEventListener('click', e => {
      const pill = e.target.closest('.pill');
      if (!pill) return;
      state.loginMode = pill.getAttribute('data-mode');
      applyPills('loginModePills', state.loginMode, 'data-mode');
      saveConfig();
    });

    document.getElementById('autoStartItem').addEventListener('click', () => {
      state.autoStart = !state.autoStart;
      toggleSwitch(document.getElementById('autoStartSwitch'), state.autoStart);
      saveConfig();
      if (window.goSetAutoStart) {
        window.goSetAutoStart(state.autoStart);
      }
    });

    document.getElementById('openOnRunItem').addEventListener('click', () => {
      state.openOnRun = !state.openOnRun;
      toggleSwitch(document.getElementById('openOnRunSwitch'), state.openOnRun);
      saveConfig();
    });

    document.getElementById('ssidInput').addEventListener('blur', saveConfig);
    document.getElementById('sidInput').addEventListener('blur', saveConfig);
    document.getElementById('pwdInput').addEventListener('blur', saveConfig);
    document.getElementById('intervalInput').addEventListener('blur', saveConfig);

    document.getElementById('btnLoginNow').addEventListener('click', () => {
      if (window.goLoginNow) window.goLoginNow();
    });
    document.getElementById('btnLogout').addEventListener('click', () => {
      if (window.goLogoutNow) window.goLogoutNow();
    });
    document.getElementById('btnOpenConfig').addEventListener('click', () => {
      if (window.goOpenConfigFile) window.goOpenConfigFile();
    });
    document.getElementById('btnQuit').addEventListener('click', () => {
      if (window.goQuitApp) window.goQuitApp();
    });
  }
  function applyStatusToUI(text) {
    const statusEl = document.getElementById('statusText');
    const footerEl = document.getElementById('footerStatus');
    if (statusEl) statusEl.textContent = text || '未知';
    if (footerEl) footerEl.textContent = text || '后台循环检测中…';
  }

  async function pollStatusLoop() {
    if (!window.goGetStatus) return;
    while (true) {
      try {
        const s = await window.goGetStatus();
        applyStatusToUI(s);
      } catch (e) {
        console.error('poll status error', e);
      }
      await new Promise(r => setTimeout(r, 1000)); // 每秒刷新一次
    }
  }
  window.addEventListener('DOMContentLoaded', () => {
    bindEvents();
    loadConfig();
    pollStatusLoop();
  });
</script>
</body>
</html>
`

func currentViewConfig() viewConfig {
	vc := viewConfig{
		Operator:          "telecom",
		LoginMode:         "operator_id",
		AutoLoginInterval: 10,
		AutoStart:         config.IsAutoStartEnabled(),
		OpenSettingsOnRun: true,
		BuildInfo:         buildInfo,
	}

	if globalCfg == nil {
		return vc
	}

	vc.SSID = globalCfg.WifiSSID
	vc.StudentID = globalCfg.Account.StudentID
	vc.Password = globalCfg.Account.Password
	if globalCfg.Account.Carrier != "" {
		vc.Operator = globalCfg.Account.Carrier
	}
	if globalCfg.LoginMode != "" {
		vc.LoginMode = globalCfg.LoginMode
	}
	if globalCfg.AutoLoginInterval > 0 {
		vc.AutoLoginInterval = globalCfg.AutoLoginInterval
	}
	vc.AutoStart = globalCfg.AutoStart
	vc.OpenSettingsOnRun = globalCfg.OpenSettingsOnRun

	return vc
}

func applyViewConfig(vc viewConfig) error {
	if globalCfg == nil {
		return fmt.Errorf("config not loaded")
	}

	if vc.AutoLoginInterval <= 0 {
		vc.AutoLoginInterval = 10
	}
	if vc.Operator == "" {
		vc.Operator = "telecom"
	}
	if vc.LoginMode == "" {
		vc.LoginMode = "operator_id"
	}

	globalCfg.WifiSSID = vc.SSID
	globalCfg.Account.StudentID = vc.StudentID
	globalCfg.Account.Password = vc.Password
	globalCfg.Account.Carrier = vc.Operator
	globalCfg.AutoLoginInterval = vc.AutoLoginInterval
	globalCfg.AutoStart = vc.AutoStart
	globalCfg.OpenSettingsOnRun = vc.OpenSettingsOnRun
	globalCfg.LoginMode = vc.LoginMode

	if vc.LoginMode == "campus_only" {
		loginUseCarrierSuffix = false
	} else {
		loginUseCarrierSuffix = true
		globalCfg.LoginMode = "operator_id"
	}

	updateLoginAccountFields()

	if err := globalCfg.Save(); err != nil {
		return err
	}
	if err := config.SetAutoStart(globalCfg.AutoStart); err != nil {
		log.Println("SetAutoStart failed:", err)
	}

	return nil
}

func openSettingsWindow() {
	settingsMu.Lock()
	if settingsW != nil {
		settingsMu.Unlock()
		return
	}
	settingsMu.Unlock()

	go func() {
		w := webview.New(false)
		settingsMu.Lock()
		settingsW = w
		settingsMu.Unlock()
		defer func() {
			w.Destroy()
			settingsMu.Lock()
			settingsW = nil
			settingsMu.Unlock()
		}()

		w.SetTitle("CUMT Autologin 设置")
		w.SetSize(620, 440, webview.HintNone)
		centerSettingsWindow(w, 620, 440)

		_ = w.Bind("goGetConfig", func() viewConfig {
			return currentViewConfig()
		})

		_ = w.Bind("goSaveConfig", func(vc viewConfig) error {
			return applyViewConfig(vc)
		})

		_ = w.Bind("goSetAutoStart", func(enabled bool) {
			if globalCfg != nil {
				globalCfg.AutoStart = enabled
			}
			if err := config.SetAutoStart(enabled); err != nil {
				log.Println("SetAutoStart failed:", err)
			}
		})
		_ = w.Bind("goGetStatus", func() string {
			return getStatus()
		})

		_ = w.Bind("goLoginNow", func() {
			go loginOnce()
		})

		_ = w.Bind("goLogoutNow", func() {
			go logoutOnce()
		})

		_ = w.Bind("goOpenConfigFile", func() {
			go openConfig()
		})

		_ = w.Bind("goQuitApp", func() {
			systray.Quit()
		})

		w.SetHtml(settingsHTML)
		w.Run()
	}()

}

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procMoveWindow       = user32.NewProc("MoveWindow")
)

const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

func getSystemMetrics(idx int32) int32 {
	r, _, _ := procGetSystemMetrics.Call(uintptr(idx))
	return int32(r)
}

func centerSettingsWindow(w webview.WebView, width, height int) {
	hwndPtr := w.Window()
	if hwndPtr == nil {
		return
	}
	hwnd := windows.Handle(uintptr(hwndPtr))

	screenW := getSystemMetrics(SM_CXSCREEN)
	screenH := getSystemMetrics(SM_CYSCREEN)
	if screenW == 0 || screenH == 0 {
		return
	}

	x := (screenW - int32(width)) / 2
	y := (screenH - int32(height)) / 3 // 稍微偏上看起来舒服点

	// 延迟一点再移，保证窗口已经创建好
	go func() {
		time.Sleep(80 * time.Millisecond)
		procMoveWindow.Call(
			uintptr(hwnd),
			uintptr(x),
			uintptr(y),
			uintptr(int32(width)),
			uintptr(int32(height)),
			uintptr(1), // bRepaint = TRUE
		)
	}()
}
