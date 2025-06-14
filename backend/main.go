package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"waf-go/internal/config"
	"waf-go/internal/database"
	"waf-go/internal/logger"
	"waf-go/internal/router"
	"waf-go/internal/service"

	"github.com/gin-gonic/gin"
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
	// 初始化配置
	cfg := config.Init()

	// 初始化日志
	logger.Init(cfg.Log.Level)

	// 初始化数据库
	db := database.Init(cfg.Database)

	// 初始化 Redis
	rdb := database.InitRedis(cfg.Redis)

	// 初始化服务
	services := service.NewServices(db, rdb, cfg)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化路由
	r := router.Init(services)

	// 启动服务器
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}
