package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	appconfig "CUMT-autologin/internal/config"
	"CUMT-autologin/internal/netcheck"
	"CUMT-autologin/internal/portal"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Status is returned to the frontend to describe connectivity.
type Status struct {
	Online    bool      `json:"online"`
	Message   string    `json:"message"`
	LastCheck time.Time `json:"last_check"`
}

// App bridges internal logic to the Wails frontend.
type App struct {
	ctx context.Context

	statusMu sync.RWMutex
	status   Status

	stopCh    chan struct{}
	wg        sync.WaitGroup
	loginMu   sync.Mutex
	lastLogin time.Time
}

func NewApp() *App {
	return &App{
		stopCh: make(chan struct{}),
		status: Status{
			Online:    false,
			Message:   "初始化中",
			LastCheck: time.Now(),
		},
	}
}

// Startup is invoked by Wails once the runtime is ready.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Ensure window is visible and centered even if last saved position was off-screen.
	runtime.WindowShow(a.ctx)
	runtime.WindowCenter(a.ctx)
	a.setupTray()
	a.startBackgroundLoop()
}

// Shutdown cleans up background goroutines.
func (a *App) Shutdown(_ context.Context) {
	a.stopBackgroundLoop()
}

// GetConfig reads config.yaml.
func (a *App) GetConfig() (*appconfig.Config, error) {
	return appconfig.Load(appconfig.DefaultConfigPath)
}

// SaveConfig persists the provided configuration.
func (a *App) SaveConfig(cfg *appconfig.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.AutoLoginInterval <= 0 {
		cfg.AutoLoginInterval = 10
	}
	if cfg.LoginMode == "" {
		cfg.LoginMode = "operator_id"
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}

// LoginNow triggers a login immediately.
func (a *App) LoginNow() (string, error) {
	cfg, err := appconfig.Load(appconfig.DefaultConfigPath)
	if err != nil {
		a.setStatus(false, "配置读取失败", time.Now())
		return "", err
	}
	msg, err := a.performLogin(cfg)
	if err != nil {
		return "", err
	}
	a.setStatus(true, msg, time.Now())
	return msg, nil
}

// LogoutNow calls the portal logout endpoint.
func (a *App) LogoutNow() (string, error) {
	cfg, err := appconfig.Load(appconfig.DefaultConfigPath)
	if err != nil {
		a.setStatus(false, "配置读取失败", time.Now())
		return "", err
	}
	pCfg := preparePortalConfig(cfg)
	body, err := portal.Logout(pCfg)
	if err != nil {
		a.setStatus(false, "注销失败", time.Now())
		return "", err
	}
	msg := "注销完成"
	if body != "" && !portal.IsLoginSuccess(body, pCfg) {
		msg = "注销可能失败（检查浏览器）"
	}
	a.setStatus(false, msg, time.Now())
	return msg, nil
}

// GetStatus returns the latest cached status.
func (a *App) GetStatus() Status {
	a.statusMu.RLock()
	defer a.statusMu.RUnlock()
	return a.status
}

func (a *App) setStatus(online bool, message string, ts time.Time) Status {
	st := Status{
		Online:    online,
		Message:   message,
		LastCheck: ts,
	}
	a.statusMu.Lock()
	a.status = st
	a.statusMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "status:update", st)
	}
	return st
}

func (a *App) startBackgroundLoop() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.checkAndLogin()
			case <-a.stopCh:
				return
			}
		}
	}()
}

func (a *App) stopBackgroundLoop() {
	select {
	case <-a.stopCh:
	default:
		close(a.stopCh)
	}
	a.wg.Wait()
}

func (a *App) checkAndLogin() {
	cfg, err := appconfig.Load(appconfig.DefaultConfigPath)
	now := time.Now()
	if err != nil {
		a.setStatus(false, fmt.Sprintf("配置读取失败: %v", err), now)
		return
	}

	online := netcheck.IsOnline(cfg.CheckURL)
	if online {
		a.setStatus(true, "在线", now)
		return
	}

	interval := cfg.AutoLoginInterval
	if interval <= 0 {
		interval = 10
	}
	if time.Since(a.lastLogin) < time.Duration(interval)*time.Second {
		a.setStatus(false, "离线，等待重试", now)
		return
	}

	msg, err := a.performLogin(cfg)
	if err != nil {
		a.setStatus(false, "登录失败", now)
		return
	}
	a.setStatus(true, msg, now)
}

func (a *App) performLogin(cfg *appconfig.Config) (string, error) {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()

	pCfg := preparePortalConfig(cfg)
	body, err := portal.Login(pCfg)
	a.lastLogin = time.Now()
	if err != nil {
		return "", err
	}
	if portal.IsLoginSuccess(body, pCfg) {
		return "登录成功", nil
	}

	_ = os.WriteFile("last_login_response.html", []byte(body), 0644)
	return "登录可能失败（网关响应异常）", fmt.Errorf("login response did not match success keywords")
}

func preparePortalConfig(cfg *appconfig.Config) *appconfig.PortalConfig {
	pCfg := cfg.Portal
	if pCfg.Form == nil {
		pCfg.Form = make(map[string]string)
	}
	account := cfg.Account.StudentID
	if strings.ToLower(cfg.LoginMode) != "campus_only" {
		account += appconfig.CarrierSuffix(cfg.Account.Carrier)
	}
	pCfg.Form["user_account"] = account
	pCfg.Form["user_password"] = cfg.Account.Password
	return &pCfg
}

func (a *App) setupTray() {
	if a.ctx == nil {
		return
	}
	appMenu := menu.NewMenuFromItems(
		menu.Text("打开设置面板", nil, func(_ *menu.CallbackData) {
			runtime.WindowShow(a.ctx)
			runtime.WindowCenter(a.ctx)
		}),
		menu.Text("立即登录", nil, func(_ *menu.CallbackData) {
			if _, err := a.LoginNow(); err != nil {
				runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Title:   "登录失败",
					Message: err.Error(),
				})
			}
		}),
		menu.Text("注销", nil, func(_ *menu.CallbackData) {
			if _, err := a.LogoutNow(); err != nil {
				runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Title:   "注销失败",
					Message: err.Error(),
				})
			}
		}),
		menu.Separator(),
		menu.Text("退出", nil, func(_ *menu.CallbackData) {
			a.stopBackgroundLoop()
			runtime.Quit(a.ctx)
		}),
	)
	runtime.MenuSetApplicationMenu(a.ctx, appMenu)
}
