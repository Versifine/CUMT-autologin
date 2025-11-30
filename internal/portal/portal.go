package portal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"CUMT-autologin/internal/config"
)

var ErrEmptyURL = errors.New("portal: login_url is empty")

func buildQuery(base string, params map[string]string) string {
	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func doGet(fullURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func Login(cfg *config.PortalConfig) (string, error) {
	loginURL := cfg.LoginURL
	if loginURL == "" {
		return "", ErrEmptyURL
	}

	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = http.MethodGet
	}

	// 把 form 拼成 query / body
	values := url.Values{}
	for k, v := range cfg.Form {
		values.Set(k, v)
	}
	encoded := values.Encode()

	var body io.Reader
	if method == http.MethodPost {
		body = strings.NewReader(encoded)
	} else {
		// GET：把参数拼到 URL 上
		if strings.Contains(loginURL, "?") {
			loginURL = loginURL + "&" + encoded
		} else {
			loginURL = loginURL + "?" + encoded
		}
		body = nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, loginURL, body)
	if err != nil {
		return "", err
	}

	// 默认 header
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (campus-netlogin-win)")

	// 覆盖/追加用户配置的 header
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func Logout(p *config.PortalConfig) (string, error) {
	if len(p.LogoutForm) == 0 {
		// 没配置注销参数，就直接返回空
		return "", nil
	}
	fullURL := buildQuery(p.LoginURL, p.LogoutForm)
	return doGet(fullURL, p.Headers)
}

func IsLoginSuccess(body string, cfg *config.PortalConfig) bool {
	if len(cfg.SuccessKeywords) == 0 {
		// 如果没配置关键字，就不做判断
		return true
	}
	for _, kw := range cfg.SuccessKeywords {
		if kw == "" {
			continue
		}
		if strings.Contains(body, kw) {
			return true
		}
	}
	return false
}
