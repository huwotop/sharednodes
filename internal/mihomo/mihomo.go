package mihomo

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

type HC struct {
	*http.Client
	proxy constant.Proxy
}

var clientPool = sync.Pool{
	New: func() interface{} {
		return &http.Client{
			Timeout: 300 * time.Second,
		}
	},
}

var transportPool = sync.Pool{
	New: func() interface{} {
		return &http.Transport{
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		}
	},
}

func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

func Proxy(raw map[string]any) *HC {
	if raw == nil {
		return nil
	}
	proxy, err := adapter.ParseProxy(raw)
	if err != nil {
		if proxy != nil {
			proxy.Close()
		}
		return nil
	}

	client := clientPool.Get().(*http.Client)
	transport := transportPool.Get().(*http.Transport)

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		u16Port, err := parsePort(portStr)
		if err != nil {
			u16Port = 0
		}

		return proxy.DialContext(ctx, &constant.Metadata{
			Host:    host,
			DstPort: u16Port,
		})
	}

	client.Transport = transport
	return &HC{Client: client, proxy: proxy}
}

func (h *HC) Release() {
	if h.Client == nil {
		return
	}
	if h.proxy != nil {
		h.proxy.Close()
		h.proxy = nil
	}
	if transport, ok := h.Transport.(*http.Transport); ok {
		transport.DialContext = nil
		transport.TLSClientConfig = nil
		transport.Proxy = nil
		transport.CloseIdleConnections()
		transportPool.Put(transport)
	}
	h.Transport = nil
	h.Timeout = 300 * time.Second
	h.CheckRedirect = nil
	h.Jar = nil
	clientPool.Put(h.Client)
}
