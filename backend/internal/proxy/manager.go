package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
	"waf-go/internal/models"

	"github.com/gin-gonic/gin"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	proxies   sync.Map
	domains   map[string]*models.Domain
	tlsConfig map[string]*tls.Config
	mu        sync.RWMutex
}

// NewProxyManager 创建代理管理器
func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		domains:   make(map[string]*models.Domain),
		tlsConfig: make(map[string]*tls.Config),
	}
}

// UpdateDomain 更新或添加域名配置
func (pm *ProxyManager) UpdateDomain(domain *models.Domain) error {
	if !domain.Enabled {
		pm.RemoveDomain(domain.Domain)
		return nil
	}

	// 解析后端URL
	target, err := url.Parse(domain.BackendURL)
	if err != nil {
		return fmt.Errorf("invalid backend URL: %v", err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		pm.modifyRequest(req, target)
	}

	// 自定义错误处理
	proxy.ErrorHandler = pm.errorHandler

	// 配置Transport
	proxy.Transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DisableKeepAlives:   false,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 如果是HTTPS，配置TLS
	if domain.Protocol == "https" {
		cert, err := tls.X509KeyPair([]byte(domain.SSLCertificate), []byte(domain.SSLPrivateKey))
		if err != nil {
			return fmt.Errorf("invalid SSL certificate: %v", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		pm.mu.Lock()
		pm.tlsConfig[domain.Domain] = tlsConfig
		pm.mu.Unlock()
	}

	// 更新代理配置
	pm.mu.Lock()
	pm.proxies.Store(domain.Domain, proxy)
	pm.domains[domain.Domain] = domain
	pm.mu.Unlock()

	log.Printf("Domain updated: %s -> %s", domain.Domain, domain.BackendURL)
	return nil
}

// RemoveDomain 移除域名配置
func (pm *ProxyManager) RemoveDomain(domain string) {
	pm.mu.Lock()
	pm.proxies.Delete(domain)
	delete(pm.domains, domain)
	delete(pm.tlsConfig, domain)
	pm.mu.Unlock()
	log.Printf("Domain removed: %s", domain)
}

// GetProxy 获取域名代理
func (pm *ProxyManager) GetProxy(domain string) *httputil.ReverseProxy {
	if proxy, ok := pm.proxies.Load(domain); ok {
		return proxy.(*httputil.ReverseProxy)
	}
	return nil
}

// GetTLSConfig 获取域名的TLS配置
func (pm *ProxyManager) GetTLSConfig(domain string) *tls.Config {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.tlsConfig[domain]
}

// GetDomainConfig 获取域名配置信息
func (pm *ProxyManager) GetDomainConfig(domain string) *models.Domain {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if domainConfig, exists := pm.domains[domain]; exists {
		return domainConfig
	}
	return nil
}

// modifyRequest 修改请求
func (pm *ProxyManager) modifyRequest(req *http.Request, target *url.URL) {
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host

	// 添加代理相关的头部
	if clientIP := req.Header.Get("X-Real-IP"); clientIP == "" {
		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			req.Header.Set("X-Real-IP", ip)
		}
	}

	// 添加或更新X-Forwarded-For
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			req.Header.Set("X-Forwarded-For", xff+", "+ip)
		}
	} else {
		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			req.Header.Set("X-Forwarded-For", ip)
		}
	}

	// 添加X-Forwarded-Proto
	if req.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	} else {
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	// 添加X-Forwarded-Host
	if req.Header.Get("X-Forwarded-Host") == "" {
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
}

// errorHandler 错误处理
func (pm *ProxyManager) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Proxy error: %v", err)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(fmt.Sprintf("Proxy error: %v", err)))
}

// ServeHTTP 处理HTTP请求
func (pm *ProxyManager) ServeHTTP(c *gin.Context) {
	domain := c.Request.Host
	pm.mu.RLock()
	domainConfig, exists := pm.domains[domain]
	pm.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Domain not found: %s", domain),
		})
		return
	}

	if !domainConfig.Enabled {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Domain is disabled",
		})
		return
	}

	proxy := pm.GetProxy(domain)
	if proxy == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Proxy not found",
		})
		return
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
