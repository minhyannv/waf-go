package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waf-go/internal/config"
	"waf-go/internal/db"
	"waf-go/internal/proxy"
	"waf-go/internal/router"
	"waf-go/internal/service"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	database := db.InitDB(cfg)

	// 初始化Redis客户端
	redisClient := db.InitRedis(cfg)

	// 创建代理管理器
	proxyManager := proxy.NewProxyManager()

	// 创建服务
	services := service.NewServices(database, redisClient)

	// 初始化路由
	r := router.Init(services)

	// 加载所有启用的域名到代理管理器
	if err := services.GetDomainService().LoadAllDomains(); err != nil {
		log.Printf("加载域名配置失败: %v", err)
	}

	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler: r,
	}

	// 创建HTTPS服务器
	httpsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HTTPSPort),
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				// 根据SNI获取证书
				if tlsConfig := proxyManager.GetTLSConfig(info.ServerName); tlsConfig != nil {
					return &tlsConfig.Certificates[0], nil
				}
				return nil, fmt.Errorf("no certificate for domain: %s", info.ServerName)
			},
			MinVersion: tls.VersionTLS12,
		},
	}

	// 启动HTTP和HTTPS服务器
	go func() {
		log.Printf("HTTP服务器启动在端口 %d", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	go func() {
		log.Printf("HTTPS服务器启动在端口 %d", cfg.Server.HTTPSPort)
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务器关闭失败: %v", err)
	}

	// 关闭HTTPS服务器
	if err := httpsServer.Shutdown(ctx); err != nil {
		log.Printf("HTTPS服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}
