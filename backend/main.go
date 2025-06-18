package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"waf-go/internal/config"
	"waf-go/internal/db"
	"waf-go/internal/logger"
	"waf-go/internal/router"
	"waf-go/internal/service"
	// "github.com/gin-gonic/gin"
)

// @title WAF Management API
// @version 1.0
// @description Web Application Firewall Management System API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger.InitLogger()
	logger.Init("debug")

	// 初始化数据库连接
	database := db.InitDB(cfg)

	// 初始化Redis客户端
	redisClient := db.InitRedis(cfg)

	// 初始化服务
	services := service.NewServices(database, redisClient)

	// 初始化路由
	r := router.Init(services)

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Server.HTTPPort),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("HTTP服务器启动在端口 %d", cfg.Server.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}
