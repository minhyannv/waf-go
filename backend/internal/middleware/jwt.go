package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		fmt.Printf("Auth header: %s\n", authHeader)

		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "未提供认证信息")
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证格式错误，请使用Bearer token")
			c.Abort()
			return
		}

		// 解析token
		claims, err := service.ParseToken(parts[1])
		if err != nil {
			fmt.Printf("Token parse error: %v\n", err)
			utils.ErrorResponse(c, http.StatusUnauthorized, "token已过期或无效，请重新登录: "+err.Error())
			c.Abort()
			return
		}

		// 验证claims是否完整
		if claims == nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息为空")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("tenant_id", claims.TenantID)
		c.Set("claims", claims)

		// 打印调试信息
		fmt.Printf("JWT Auth - User ID: %d, Username: %s, Role: %s, Tenant ID: %d\n",
			claims.UserID, claims.Username, claims.Role, claims.TenantID)

		c.Next()
	}
}
