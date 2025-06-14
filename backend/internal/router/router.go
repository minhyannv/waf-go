package router

import (
	"waf-go/internal/handler"
	"waf-go/internal/middleware"
	"waf-go/internal/service"

	"github.com/gin-gonic/gin"
)

func Init(services *service.Services) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WAF 中间件 - 应用于所有非管理接口
	wafMiddleware := services.WAFEngine.ProcessRequest()

	// 创建处理器
	authHandler := handler.NewAuthHandler(services.Auth)
	ruleHandler := handler.NewRuleHandler(services.Rule)
	policyHandler := handler.NewPolicyHandler(services.Policy)
	logHandler := handler.NewLogHandler(services.Log)
	dashboardHandler := handler.NewDashboardHandler(services.Dashboard)
	whiteListHandler := handler.NewWhiteListHandler(services.WhiteList)
	blackListHandler := handler.NewBlackListHandler(services.BlackList)
	configHandler := handler.NewConfigHandler(services.Config)

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// 用户信息
			protected.GET("/auth/userinfo", authHandler.GetUserInfo)

			// 仪表盘
			protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
			protected.GET("/dashboard/realtime", dashboardHandler.GetRealtimeStats)

			// 规则管理
			rules := protected.Group("/rules")
			{
				rules.GET("", ruleHandler.GetRuleList)
				rules.POST("", ruleHandler.CreateRule)
				rules.GET("/:id", ruleHandler.GetRule)
				rules.PUT("/:id", ruleHandler.UpdateRule)
				rules.DELETE("/:id", ruleHandler.DeleteRule)
				rules.POST("/:id/toggle", ruleHandler.ToggleRule)
			}

			// 策略管理
			policies := protected.Group("/policies")
			{
				policies.GET("", policyHandler.GetPolicyList)
				policies.POST("", policyHandler.CreatePolicy)
				policies.GET("/:id", policyHandler.GetPolicy)
				policies.GET("/:id/rules", policyHandler.GetPolicyWithRules)
				policies.PUT("/:id", policyHandler.UpdatePolicy)
				policies.DELETE("/:id", policyHandler.DeletePolicy)
				policies.POST("/:id/toggle", policyHandler.TogglePolicy)
				policies.DELETE("/batch", policyHandler.BatchDeletePolicies)
				policies.GET("/rules/available", policyHandler.GetAvailableRules)
			}

			// 日志管理
			logs := protected.Group("/logs")
			{
				logs.GET("/attacks", logHandler.GetAttackLogList)
				logs.GET("/attacks/:id", logHandler.GetAttackLogDetail)
				logs.DELETE("/attacks/:id", logHandler.DeleteAttackLog)
				logs.DELETE("/attacks/batch", logHandler.BatchDeleteAttackLogs)
				logs.GET("/attacks/export", logHandler.ExportAttackLogs)
				logs.POST("/attacks/clean", logHandler.CleanOldLogs)
			}

			// 白名单管理
			whitelists := protected.Group("/whitelists")
			{
				whitelists.GET("", whiteListHandler.GetWhiteListList)
				whitelists.POST("", whiteListHandler.CreateWhiteList)
				whitelists.GET("/:id", whiteListHandler.GetWhiteListByID)
				whitelists.PUT("/:id", whiteListHandler.UpdateWhiteList)
				whitelists.DELETE("/:id", whiteListHandler.DeleteWhiteList)
				whitelists.DELETE("/batch", whiteListHandler.BatchDeleteWhiteList)
				whitelists.PATCH("/:id/toggle", whiteListHandler.ToggleWhiteListStatus)
			}

			// 黑名单管理
			blacklists := protected.Group("/blacklists")
			{
				blacklists.GET("", blackListHandler.GetBlackListList)
				blacklists.POST("", blackListHandler.CreateBlackList)
				blacklists.GET("/:id", blackListHandler.GetBlackListByID)
				blacklists.PUT("/:id", blackListHandler.UpdateBlackList)
				blacklists.DELETE("/:id", blackListHandler.DeleteBlackList)
				blacklists.DELETE("/batch", blackListHandler.BatchDeleteBlackList)
				blacklists.PATCH("/:id/toggle", blackListHandler.ToggleBlackListStatus)
			}

			// 系统配置
			config := protected.Group("/config")
			{
				config.GET("", configHandler.GetSystemConfig)
				config.PUT("", configHandler.UpdateSystemConfig)
				config.POST("/reset", configHandler.ResetSystemConfig)
				config.GET("/stats", configHandler.GetConfigStats)
			}
		}
	}

	// 受保护的应用路由（应用WAF）
	app := r.Group("/app")
	app.Use(wafMiddleware)
	{
		// 这里可以添加实际的应用路由
		// 例如：app.GET("/api/data", someHandler)

		// 测试路由
		app.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello from protected app!",
				"ip":      c.ClientIP(),
			})
		})

		// 触发WAF的测试路由
		app.GET("/admin", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Admin page",
			})
		})
	}

	return r
}
