package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

func NewSOCKS5ProxyClient(httpsProxy, httpProxy string) (*http.Client, error) {
	proxyAddress := strings.TrimSpace(httpsProxy)
	if proxyAddress == "" {
		proxyAddress = strings.TrimSpace(httpProxy)
	}
	if proxyAddress == "" {
		return nil, errors.New("SOCKS5 proxy URL is required in HTTPS_PROXY or HTTP_PROXY")
	}

	proxyURL, err := url.Parse(proxyAddress)
	if err != nil || proxyURL.Scheme != "socks5" || proxyURL.Host == "" {
		return nil, errors.New("proxy URL must use socks5:// and include a host")
	}
	host, portText, err := net.SplitHostPort(proxyURL.Host)
	if err != nil || host == "" {
		return nil, errors.New("SOCKS5 proxy URL must include a host and port")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("SOCKS5 proxy port must be between 1 and 65535")
	}

	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{User: proxyURL.User.Username(), Password: password}
	}
	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("configure SOCKS5 proxy: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return dialer.Dial(network, addr)
		},
	}
	return &http.Client{Transport: transport, Timeout: 180 * time.Second}, nil
}
