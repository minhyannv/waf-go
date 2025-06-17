package middleware

import (
	"net/http"
	"strings"

	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 移除 Bearer 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 解析令牌
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("tenant_id", claims.TenantID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 角色权限检查中间件
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限访问",
			})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		c.Abort()
	}
}

// GetUserContext 从Gin上下文获取用户信息
func GetUserContext(c *gin.Context) *service.UserContext {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	tenantID, _ := c.Get("tenant_id")
	tenantCode, exists := c.Get("tenant_code")

	var tenantCodeStr string
	if exists && tenantCode != nil {
		tenantCodeStr = tenantCode.(string)
	}

	return &service.UserContext{
		UserID:     userID.(uint),
		Username:   username.(string),
		Role:       role.(string),
		TenantID:   tenantID.(uint),
		TenantCode: tenantCodeStr,
	}
}
