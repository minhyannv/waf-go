package router

import (
	"fmt"
	"net/http"
	"strings"
	"waf-go/internal/handler"
	"waf-go/internal/middleware"
	"waf-go/internal/service"

	"github.com/gin-gonic/gin"
)

// WAF 中间件 - 应用于所有非管理接口
func wafMiddleware(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过API请求
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		result, err := services.GetWAFEngine().CheckRequest(c)
		if err != nil {
			c.JSON(500, gin.H{"error": "WAF check failed"})
			c.Abort()
			return
		}

		if result.Action == "block" {
			services.GetWAFEngine().LogAttack(c, result)
			c.JSON(result.StatusCode, gin.H{"message": result.Message})
			c.Abort()
			return
		}

		if result.Action == "log" {
			services.GetWAFEngine().LogAttack(c, result)
		}

		c.Next()
	}
}

func Init(services *service.Services) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Content-Type", "application/json; charset=utf-8")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建处理器
	authHandler := handler.NewAuthHandler(services.GetAuthService())
	ruleHandler := handler.NewRuleHandler(services.GetRuleService())
	policyHandler := handler.NewPolicyHandler(services.GetPolicyService())
	logHandler := handler.NewLogHandler(services.GetLogService())
	dashboardHandler := handler.NewDashboardHandler(services.GetDashboardService())
	whiteListHandler := handler.NewWhiteListHandler(services.GetWhiteListService())
	blackListHandler := handler.NewBlackListHandler(services.GetBlackListService())
	configHandler := handler.NewConfigHandler(services.GetConfigService())
	domainHandler := handler.NewDomainHandler(services.GetDomainService(), services.GetTenantSecurityService())

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/userinfo", middleware.JWTAuth(), authHandler.GetUserInfo)
		}

		// 需要认证的API
		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 规则管理
			rules := protected.Group("/rules")
			{
				rules.GET("", ruleHandler.GetRuleList)
				rules.POST("", ruleHandler.CreateRule)
				rules.GET("/:id", ruleHandler.GetRule)
				rules.PUT("/:id", ruleHandler.UpdateRule)
				rules.DELETE("/:id", ruleHandler.DeleteRule)
				rules.DELETE("/batch", ruleHandler.BatchDeleteRules)
				rules.PATCH("/:id/toggle", ruleHandler.ToggleRule)
				rules.POST("/:id/toggle", ruleHandler.ToggleRule)
			}

			// 策略管理
			policies := protected.Group("/policies")
			{
				policies.GET("", policyHandler.GetPolicyList)
				policies.POST("", policyHandler.CreatePolicy)
				policies.GET("/:id", policyHandler.GetPolicy)
				policies.GET("/:id/with-rules", policyHandler.GetPolicyWithRules)
				policies.PUT("/:id", policyHandler.UpdatePolicy)
				policies.DELETE("/:id", policyHandler.DeletePolicy)
				policies.DELETE("/batch", policyHandler.BatchDeletePolicies)
				policies.PATCH("/:id/toggle", policyHandler.TogglePolicy)
				policies.GET("/:id/rules", policyHandler.GetPolicyRules)
				policies.PUT("/:id/rules", policyHandler.UpdatePolicyRules)
			}

			// 日志管理
			logs := protected.Group("/logs")
			{
				// 攻击日志
				attacks := logs.Group("/attacks")
				{
					attacks.GET("", logHandler.GetAttackLogList)
					attacks.GET("/:id", logHandler.GetAttackLogDetail)
					attacks.DELETE("/:id", logHandler.DeleteAttackLog)
					attacks.DELETE("/batch", logHandler.BatchDeleteAttackLogs)
					attacks.POST("/clean", logHandler.CleanOldLogs)
					attacks.GET("/export", logHandler.ExportAttackLogs)
				}
			}

			// 仪表盘
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.GetDashboardStats)
				dashboard.GET("/overview", dashboardHandler.GetOverview)
				dashboard.GET("/attack_trend", dashboardHandler.GetAttackTrend)
				dashboard.GET("/top_rules", dashboardHandler.GetTopRules)
				dashboard.GET("/top_ips", dashboardHandler.GetTopIPs)
				dashboard.GET("/top_uris", dashboardHandler.GetTopURIs)
				dashboard.GET("/top_user_agents", dashboardHandler.GetTopUserAgents)
			}

			// 白名单管理
			whitelists := protected.Group("/whitelists")
			{
				whitelists.GET("", whiteListHandler.GetWhiteLists)
				whitelists.POST("", whiteListHandler.CreateWhiteList)
				whitelists.GET("/:id", whiteListHandler.GetWhiteListByID)
				whitelists.PUT("/:id", whiteListHandler.UpdateWhiteList)
				whitelists.DELETE("/:id", whiteListHandler.DeleteWhiteList)
				whitelists.PATCH("/:id/toggle", whiteListHandler.ToggleWhiteList)
				whitelists.POST("/:id/toggle", whiteListHandler.ToggleWhiteList)
			}

			// 黑名单管理
			blacklists := protected.Group("/blacklists")
			{
				blacklists.GET("", blackListHandler.GetBlackLists)
				blacklists.POST("", blackListHandler.CreateBlackList)
				blacklists.GET("/:id", blackListHandler.GetBlackListByID)
				blacklists.PUT("/:id", blackListHandler.UpdateBlackList)
				blacklists.DELETE("/:id", blackListHandler.DeleteBlackList)
				blacklists.PATCH("/:id/toggle", blackListHandler.ToggleBlackList)
				blacklists.POST("/:id/toggle", blackListHandler.ToggleBlackList)
			}

			// 系统配置
			config := protected.Group("/config")
			{
				config.GET("", configHandler.GetSystemConfig)
				config.PUT("", configHandler.UpdateSystemConfig)
				config.POST("/reset", configHandler.ResetSystemConfig)
				config.GET("/stats", configHandler.GetConfigStats)
			}

			// 域名管理
			domains := protected.Group("/domains")
			{
				domains.GET("", domainHandler.GetDomainList)
				domains.POST("", domainHandler.CreateDomain)
				domains.GET("/:id", domainHandler.GetDomain)
				domains.PUT("/:id", domainHandler.UpdateDomain)
				domains.DELETE("/:id", domainHandler.DeleteDomain)
				domains.POST("/:id/toggle", domainHandler.ToggleDomain)
				domains.GET("/:id/policies", domainHandler.GetDomainPolicies)
				domains.PUT("/:id/policies", domainHandler.UpdateDomainPolicies)
				domains.DELETE("/batch", domainHandler.BatchDeleteDomains)
			}
		}
	}

	// 受保护的应用路由（应用WAF）
	app := r.Group("/app")
	app.Use(wafMiddleware(services))
	{
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
				"message": "Admin panel accessed!",
				"ip":      c.ClientIP(),
			})
		})
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// 先检查域名是否存在
		proxy := services.GetDomainService().GetProxyManager().GetProxy(c.Request.Host)
		if proxy == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Domain not found: %s", c.Request.Host),
			})
			return
		}

		// 检查域名是否关联了策略（即是否接入WAF）
		hasPolicies := services.GetDomainService().HasDomainPolicies(c.Request.Host)
		if hasPolicies {
			// 域名关联了策略，需要经过WAF检查
			wafMiddleware(services)(c)
			if c.IsAborted() {
				return // WAF拦截了请求
			}
		}

		// 转发到后端服务
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	return r
}
