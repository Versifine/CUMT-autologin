package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config.yaml"

type PortalConfig struct {
	LoginURL        string            `yaml:"login_url"`
	Method          string            `yaml:"method"`
	Form            map[string]string `yaml:"form"`
	LogoutForm      map[string]string `yaml:"logout_form"`
	Headers         map[string]string `yaml:"headers"`
	SuccessKeywords []string          `yaml:"success_keywords"`
}

type AccountConfig struct {
	StudentID string `yaml:"student_id"`
	Carrier   string `yaml:"carrier"` // telecom / unicom / cmcc
	Password  string `yaml:"password"`
}
type UIConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type Config struct {
	WifiSSID string        `yaml:"wifi_ssid"`
	CheckURL string        `yaml:"check_url"`
	Account  AccountConfig `yaml:"account"`
	Portal   PortalConfig  `yaml:"portal"`
	UI       UIConfig      `yaml:"ui"`

	AutoLoginInterval int    `yaml:"auto_login_interval" json:"auto_login_interval"`
	LoginMode         string `yaml:"login_mode" json:"login_mode"`
	AutoStart         bool   `yaml:"auto_start" json:"auto_start"`
	OpenSettingsOnRun bool   `yaml:"open_settings_on_run" json:"open_settings_on_run"`

	path string `yaml:"-"`
}

func CarrierSuffix(carrier string) string {
	switch strings.ToLower(carrier) {
	case "none", "":
		return ""
	case "telecom", "ct", "dx":
		return "@telecom"
	case "unicom", "cu", "lt":
		return "@unicom"
	case "cmcc", "mobile", "yd":
		return "@cmcc"
	default:
		return "@telecom"
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw := map[string]any{}
	_ = yaml.Unmarshal(data, &raw)

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	c.path = path

	if c.CheckURL == "" {
		c.CheckURL = "http://www.msftconnecttest.com/connecttest.txt"
	}
	if c.Portal.Method == "" {
		c.Portal.Method = "GET"
	}
	if c.Portal.Form == nil {
		c.Portal.Form = make(map[string]string)
	}
	if c.Portal.LogoutForm == nil {
		c.Portal.LogoutForm = make(map[string]string)
	}
	if c.UI.Width <= 0 {
		c.UI.Width = 400
	}
	if c.UI.Height <= 0 {
		c.UI.Height = 460
	}
	if c.AutoLoginInterval <= 0 {
		c.AutoLoginInterval = 10
	}
	if c.LoginMode == "" {
		c.LoginMode = "operator_id"
	}
	if _, ok := raw["open_settings_on_run"]; !ok {
		c.OpenSettingsOnRun = true
	}

	if c.Account.StudentID != "" {
		suffix := CarrierSuffix(c.Account.Carrier)
		userAccount := c.Account.StudentID + suffix
		c.Portal.Form["user_account"] = userAccount
		c.Portal.Form["user_password"] = c.Account.Password
	}

	return &c, nil
}

func (c *Config) Save() error {
	if c.path == "" {
		c.path = DefaultConfigPath
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, out, 0644)
}
