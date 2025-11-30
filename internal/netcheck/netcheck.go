package netcheck

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ---------- NCSI HTTP 检测 ----------
// 访问 http://www.msftconnecttest.com/connecttest.txt
// 期望响应体为 "Microsoft Connect Test"
func ncsiHTTP() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://www.msftconnecttest.com/connecttest.txt",
		nil,
	)
	if err != nil {
		return false
	}

	// 不用代理，避免本地代理干扰判断
	transport := &http.Transport{
		Proxy: nil,
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return false
	}
	body := strings.TrimSpace(string(bodyBytes))

	// NCSI 期望的精确内容
	return body == "Microsoft Connect Test"
}

// ---------- NCSI DNS 检测 ----------
// 解析 dns.msftncsi.com，期望 IP 为 131.107.255.255
func ncsiDNS() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, "dns.msftncsi.com")
	if err != nil {
		return false
	}

	expected := net.ParseIP("131.107.255.255")
	if expected == nil {
		return false
	}

	for _, ip := range ips {
		if ip.IP.Equal(expected) {
			return true
		}
	}
	return false
}

// 对外接口：checkURL 参数目前不使用，只为了兼容原来的签名。
func IsOnline(checkURL string) bool {
	httpOK := ncsiHTTP()
	dnsOK := ncsiDNS()

	// 两个都 OK 时认为“网络基本是通的”
	if httpOK && dnsOK {
		return true
	}
	return false
}
