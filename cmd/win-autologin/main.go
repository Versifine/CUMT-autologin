//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"CUMT-autologin/internal/config"
	"CUMT-autologin/internal/netcheck"
	"CUMT-autologin/internal/portal"

	"github.com/getlantern/systray"

	_ "embed"
)

var (
	buildInfo             = "dev"
	globalCfg             *config.Config
	loginUseCarrierSuffix = true
	statusMenu            *systray.MenuItem
	currentStatusText     = "启动中..."
	statusMu              sync.RWMutex
)

func setStatus(text string) {
	statusMu.Lock()
	currentStatusText = text
	statusMu.Unlock()

	if statusMenu != nil {
		statusMenu.SetTitle("状态: " + text)
	}
}

func getStatus() string {
	statusMu.RLock()
	defer statusMu.RUnlock()
	return currentStatusText
}

func runDaemon(cfg *config.Config, stopCh <-chan struct{}) {
	forceLoginInterval := 2 * time.Minute
	var lastLoginTime time.Time

	for {
		interval := cfg.AutoLoginInterval
		if interval <= 0 {
			interval = 10
		}
		checkInterval := time.Duration(interval) * time.Second

		select {
		case <-stopCh:
			setStatus("已停止")
			return
		case <-time.After(checkInterval):
			now := time.Now()

			if cfg.WifiSSID != "" {
				ssid, err := currentSSID()
				if err != nil {
					fmt.Println("[WARN] get current ssid error:", err)
					setStatus("读取 WiFi 状态出错")
					continue
				}
				if ssid == "" {
					fmt.Println("[INFO] no wifi connected, skip")
					setStatus("未连接 WiFi")
					continue
				}
				if ssid != cfg.WifiSSID {
					fmt.Printf("[INFO] current ssid=%q, target=%q, skip\n", ssid, cfg.WifiSSID)
					setStatus("已连接 " + ssid + " (非目标)")
					continue
				}
			} else {
				setStatus("未限制 WiFi")
			}

			needForceLogin := false
			if lastLoginTime.IsZero() || now.Sub(lastLoginTime) >= forceLoginInterval {
				needForceLogin = true
			}

			online := netcheck.IsOnline(cfg.CheckURL)
			fmt.Println("[DEBUG] IsOnline =", online, "needForceLogin =", needForceLogin)

			if online && !needForceLogin {
				setStatus("在线")
				fmt.Println("[INFO] already online and no need to force login, skip")
				continue
			}

			if !online {
				setStatus("未认证 / 尝试登录中...")
			} else {
				setStatus("定期重登录中...")
			}

			fmt.Println("[INFO] try login...")
			body, err := portal.Login(&cfg.Portal)
			if err != nil {
				fmt.Println("[ERROR] login error:", err)
				setStatus("登录失败（请求错误）")
				continue
			}

			lastLoginTime = now
			if portal.IsLoginSuccess(body, &cfg.Portal) {
				fmt.Println("[INFO] login response looks success")
				setStatus("在线")
			} else {
				fmt.Println("[WARN] login response not matched success keywords, save for debug")
				_ = os.WriteFile("last_login_response.html", []byte(body), 0644)
				setStatus("登录失败（网关响应异常）")
			}
		}
	}
}

func updateLoginAccountFields() {
	if globalCfg == nil {
		return
	}
	acc := globalCfg.Account

	var userAccount string
	if loginUseCarrierSuffix {
		userAccount = acc.StudentID + config.CarrierSuffix(acc.Carrier)
	} else {
		userAccount = acc.StudentID
	}

	if globalCfg.Portal.Form == nil {
		globalCfg.Portal.Form = make(map[string]string)
	}
	globalCfg.Portal.Form["user_account"] = userAccount
	globalCfg.Portal.Form["user_password"] = acc.Password

	fmt.Println("[INFO] login mode changed, user_account =", userAccount)
}

func main() {
	if !ensureSingleInstance() {
		return
	}
	defer releaseSingleInstance()

	cfg, err := config.Load(config.DefaultConfigPath)
	if err != nil {
		panic(err)
	}
	globalCfg = cfg

	if cfg.LoginMode == "campus_only" {
		loginUseCarrierSuffix = false
	} else {
		loginUseCarrierSuffix = true
		cfg.LoginMode = "operator_id"
	}
	updateLoginAccountFields()

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("CUMT Autologin")
	systray.SetTooltip("CUMT 校园网自动登录")

	statusMenu = systray.AddMenuItem("状态: 检测中...", "当前连接状态")
	statusMenu.Disable()
	systray.AddSeparator()

	mLoginNow := systray.AddMenuItem("立即尝试登录", "立即尝试登录一次校园网")

	mMode := systray.AddMenuItem("登录模式", "选择登录方式")
	mModeCarrier := mMode.AddSubMenuItemCheckbox("运营商账号 (@telecom)", "通过运营商账号登录", loginUseCarrierSuffix)
	mModeCampus := mMode.AddSubMenuItemCheckbox("校园网账号 (纯学号)", "通过校园网账号登录", !loginUseCarrierSuffix)

	mLogout := systray.AddMenuItem("注销当前会话", "调用网关注销接口")

	mOpenSettings := systray.AddMenuItem("设置...", "打开设置窗口")
	mOpenConfig := systray.AddMenuItem("打开配置文件 (YAML)", "用记事本打开 config.yaml")

	mQuit := systray.AddMenuItem("退出", "退出自动登录")

	stopCh := make(chan struct{})
	go runDaemon(globalCfg, stopCh)

	if err := config.SetAutoStart(globalCfg.AutoStart); err != nil {
		log.Println("SetAutoStart on startup failed:", err)
	}

	if globalCfg.OpenSettingsOnRun {
		openSettingsWindow()
	}

	go func() {
		for {
			select {
			case <-mLoginNow.ClickedCh:
				go loginOnce()

			case <-mLogout.ClickedCh:
				go logoutOnce()

			case <-mOpenSettings.ClickedCh:
				openSettingsWindow()

			case <-mOpenConfig.ClickedCh:
				openConfig()

			case <-mModeCarrier.ClickedCh:
				loginUseCarrierSuffix = true
				globalCfg.LoginMode = "operator_id"
				mModeCarrier.Check()
				mModeCampus.Uncheck()
				updateLoginAccountFields()

			case <-mModeCampus.ClickedCh:
				loginUseCarrierSuffix = false
				globalCfg.LoginMode = "campus_only"
				mModeCarrier.Uncheck()
				mModeCampus.Check()
				updateLoginAccountFields()

			case <-mQuit.ClickedCh:
				close(stopCh)
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	fmt.Println("[INFO] systray exiting")
}

func loginOnce() {
	updateLoginAccountFields()
	setStatus("手动登录中...")
	body, err := portal.Login(&globalCfg.Portal)
	if err != nil {
		fmt.Println("[ERROR] manual login error:", err)
		setStatus("手动登录失败")
		return
	}
	if portal.IsLoginSuccess(body, &globalCfg.Portal) {
		fmt.Println("[INFO] manual login success")
		setStatus("在线（手动登录成功）")
	} else {
		fmt.Println("[WARN] manual login response not matched success keywords")
		setStatus("手动登录失败（网关响应异常）")
	}
}

func logoutOnce() {
	fmt.Println("[INFO] manual logout triggered from tray")
	setStatus("手动注销中...")
	body, err := portal.Logout(&globalCfg.Portal)
	if err != nil {
		fmt.Println("[ERROR] logout error:", err)
		setStatus("注销失败（请求错误）")
		return
	}
	if body == "" {
		fmt.Println("[INFO] logout_form not configured, nothing to do")
		setStatus("未配置注销参数")
		return
	}
	if portal.IsLoginSuccess(body, &globalCfg.Portal) {
		fmt.Println("[INFO] logout response looks success")
		setStatus("已注销")
	} else {
		fmt.Println("[WARN] logout response may not be success")
		setStatus("注销可能失败（检查浏览器）")
	}
}

func openConfig() {
	cmd := exec.Command("notepad.exe", config.DefaultConfigPath)
	_ = cmd.Start()
}
