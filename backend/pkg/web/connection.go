package web

import (
	"net"
	"net/http"
	"time"
)

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每主机最大空闲连接
	MaxConnsPerHost     int           // 每主机最大连接数
	IdleConnTimeout     time.Duration // 空闲连接超时
	ConnTimeout         time.Duration // 连接超时
	ReadTimeout         time.Duration // 读取超时
	WriteTimeout        time.Duration // 写入超时
	EnableKeepAlive     bool          // 启用 KeepAlive
	KeepAlivePeriod     time.Duration // KeepAlive 周期
}

// DefaultConnectionPoolConfig 默认连接池配置
var DefaultConnectionPoolConfig = ConnectionPoolConfig{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	MaxConnsPerHost:     100,
	IdleConnTimeout:     90 * time.Second,
	ConnTimeout:         30 * time.Second,
	ReadTimeout:         30 * time.Second,
	WriteTimeout:        30 * time.Second,
	EnableKeepAlive:     true,
	KeepAlivePeriod:     30 * time.Second,
}

// createTransport 创建优化的 HTTP Transport
func createTransport(config ConnectionPoolConfig) *http.Transport {
	return &http.Transport{
		// 连接池设置
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,

		// 连接设置
		DialContext: (&net.Dialer{
			Timeout:   config.ConnTimeout,
			KeepAlive: config.KeepAlivePeriod,
		}).DialContext,

		// TLS 设置
		ForceAttemptHTTP2: true,

		// 压缩设置
		DisableCompression: false,
	}
}

// ClientConfig HTTP 客户端配置
type ClientConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	PoolConfig    ConnectionPoolConfig
}

// DefaultClientConfig 默认客户端配置
var DefaultClientConfig = ClientConfig{
	Timeout:       30 * time.Second,
	MaxRetries:    3,
	RetryInterval: 1 * time.Second,
	PoolConfig:    DefaultConnectionPoolConfig,
}

// NewHTTPClient 创建优化的 HTTP 客户端
func NewHTTPClient(config ClientConfig) *http.Client {
	return &http.Client{
		Timeout:   config.Timeout,
		Transport: createTransport(config.PoolConfig),
	}
}

// ClientPool HTTP 客户端池
type ClientPool struct {
	clients chan *http.Client
	config  ClientConfig
}

// NewClientPool 创建客户端池
func NewClientPool(size int, config ClientConfig) *ClientPool {
	pool := &ClientPool{
		clients: make(chan *http.Client, size),
		config:  config,
	}

	// 预先创建客户端
	for i := 0; i < size; i++ {
		pool.clients <- NewHTTPClient(config)
	}

	return pool
}

// Acquire 获取客户端
func (p *ClientPool) Acquire() *http.Client {
	select {
	case client := <-p.clients:
		return client
	default:
		// 池为空，创建新客户端
		return NewHTTPClient(p.config)
	}
}

// Release 释放客户端
func (p *ClientPool) Release(client *http.Client) {
	select {
	case p.clients <- client:
	default:
		// 池已满，丢弃客户端
	}
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	activeConnections int64
	maxConnections    int64
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(maxConnections int64) *ConnectionManager {
	return &ConnectionManager{
		maxConnections: maxConnections,
	}
}

// TCPKeepAliveListener 带 KeepAlive 的 TCP Listener
type TCPKeepAliveListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
}

// Accept 接受连接并设置 KeepAlive
func (ln TCPKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}

	if ln.keepAlivePeriod > 0 {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(ln.keepAlivePeriod)
	}

	return tc, nil
}

// createKeepAliveListener 创建带 KeepAlive 的 Listener
func createKeepAliveListener(addr string, keepAlivePeriod time.Duration) (net.Listener, error) {
	tc, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &TCPKeepAliveListener{
		TCPListener:     tc.(*net.TCPListener),
		keepAlivePeriod: keepAlivePeriod,
	}, nil
}
