package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"CUMT-autologin/internal/config"
	"CUMT-autologin/internal/netcheck"
	"CUMT-autologin/internal/portal"

	"github.com/energye/systray"
)

const (
	defaultIntervalSec = 10
	forceLoginInterval = 2 * time.Minute
	logoutFlagPath     = "logout.flag" // touch this file to trigger a logout on next tick
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	initLogging()
	if !ensureSingleInstance() {
		return
	}
	defer releaseSingleInstance()
	systray.Run(onReady, onExit)
}

// 托盘初始化
func onReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip("CUMT 校园网自动登录")

	// 左键点击托盘图标：唤起 GUI
	systray.SetOnClick(func(menu systray.IMenu) {
		go summonWildsapp()
	})

	// 右键菜单：手动登录 / 退出
	mLoginNow := systray.AddMenuItem("立即尝试登录", "立刻跑一次登录逻辑")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出 CUMT-autologin")

	// 自动登录循环
	go runCoreLoop()

	// 处理菜单点击
	mLoginNow.Click(func() {
		go runOnce()
	})
	mQuit.Click(func() {
		systray.Quit()
	})
}

func onExit() {
	// need cleanup if required
}

func initLogging() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[core] get executable path error: %v", err)
		return
	}
	exeDir := filepath.Dir(exe)
	logPath := filepath.Join(exeDir, "core.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[core] open log file error: %v", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	log.Printf("[core] logging to %s", logPath)
}

func summonWildsapp() {
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)

	candidates := []string{
		"CUMTAutologinGUI.exe", // wails GUI build
		"CUMT-autologin.exe",   // legacy name
		"wildsapp.exe",         // fallback
	}
	var uiPath string
	for _, name := range candidates {
		p := filepath.Join(exeDir, name)
		if _, err := os.Stat(p); err == nil {
			uiPath = p
			break
		}
	}
	if uiPath == "" {
		log.Printf("[tray] no UI executable found in %s", exeDir)
		return
	}

	logFile := filepath.Join(exeDir, "gui.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[tray] open gui.log error: %v", err)
	}

	cmd := exec.Command(uiPath)
	cmd.Dir = exeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// Workaround for some Windows hook/AV environments where Go async preemption crashes GUI.
	cmd.Env = append(os.Environ(), "GODEBUG=asyncpreemptoff=1")
	if f != nil {
		cmd.Stdout = f
		cmd.Stderr = f
	}
	log.Printf("[tray] starting UI: %s", uiPath)
	if err := cmd.Start(); err != nil {
		log.Printf("[tray] start UI error: %v", err)
		return
	}
	log.Printf("[tray] UI started, pid=%d", cmd.Process.Pid)
	if f != nil {
		_ = f.Close()
	}
}

func runCoreLoop() {
	var lastLogin time.Time

	for {
		cfg, err := config.Load(config.DefaultConfigPath)
		if err != nil {
			log.Printf("[core] read config error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := config.SetAutoStart(cfg.AutoStart); err != nil {
			log.Printf("[core] SetAutoStart failed: %v", err)
		}

		interval := cfg.AutoLoginInterval
		if interval <= 0 {
			interval = defaultIntervalSec
		}

		now := time.Now()

		if handleLogoutFlag(cfg) {
			lastLogin = now
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		if cfg.WifiSSID != "" {
			ssid, err := currentSSID()
			if err != nil {
				log.Printf("[core] read wifi ssid failed: %v", err)
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
			if ssid == "" {
				log.Printf("[core] wifi not connected, skip")
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
			if ssid != cfg.WifiSSID {
				log.Printf("[core] wifi=%q (target %q), skip", ssid, cfg.WifiSSID)
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
		}

		online := netcheck.IsOnline(cfg.CheckURL)
		if online && now.Sub(lastLogin) < forceLoginInterval {
			log.Printf("[core] online, no need to login (last=%s)", lastLogin.Format(time.RFC3339))
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		if online {
			log.Printf("[core] online but forcing re-login (interval elapsed)")
		} else {
			log.Printf("[core] offline, try login")
		}

		if err := doLogin(cfg); err != nil {
			log.Printf("[core] login failed: %v", err)
		} else {
			lastLogin = time.Now()
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func runOnce() {
	cfg, err := config.Load(config.DefaultConfigPath)
	if err != nil {
		log.Printf("[core] read config error: %v", err)
		return
	}

	if cfg.WifiSSID != "" {
		ssid, err := currentSSID()
		if err != nil {
			log.Printf("[core] read wifi ssid failed: %v", err)
			return
		}
		if ssid == "" || ssid != cfg.WifiSSID {
			log.Printf("[core] runOnce: wifi not match, skip")
			return
		}
	}

	if netcheck.IsOnline(cfg.CheckURL) {
		log.Printf("[core] runOnce: already online")
		return
	}

	if err := doLogin(cfg); err != nil {
		log.Printf("[core] runOnce: login failed: %v", err)
	} else {
		log.Printf("[core] runOnce: login ok")
	}
}

func handleLogoutFlag(cfg *config.Config) bool {
	if _, err := os.Stat(logoutFlagPath); err != nil {
		return false
	}
	log.Printf("[core] logout flag detected, try logout")
	pCfg := preparePortalConfig(cfg)
	body, err := portal.Logout(pCfg)
	if err != nil {
		log.Printf("[core] logout error: %v", err)
	} else if body != "" && !portal.IsLoginSuccess(body, pCfg) {
		log.Printf("[core] logout response may not indicate success")
	} else {
		log.Printf("[core] logout finished")
	}
	_ = os.Remove(logoutFlagPath)
	return true
}

func doLogin(cfg *config.Config) error {
	pCfg := preparePortalConfig(cfg)
	body, err := portal.Login(pCfg)
	if err != nil {
		return err
	}
	if portal.IsLoginSuccess(body, pCfg) {
		log.Printf("[core] login success")
		return nil
	}
	_ = os.WriteFile("last_login_response.html", []byte(body), 0644)
	return fmt.Errorf("login response did not match success keywords")
}

func preparePortalConfig(cfg *config.Config) *config.PortalConfig {
	pCfg := cfg.Portal
	if pCfg.Form == nil {
		pCfg.Form = make(map[string]string)
	}
	account := cfg.Account.StudentID
	if cfg.LoginMode != "campus_only" {
		account += config.CarrierSuffix(cfg.Account.Carrier)
	}
	pCfg.Form["user_account"] = account
	pCfg.Form["user_password"] = cfg.Account.Password
	return &pCfg
}
