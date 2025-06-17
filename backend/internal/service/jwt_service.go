package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID uint   `json:"tenant_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("waf-secret-key-change-in-production")

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username string, role string, tenantID uint) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	fmt.Printf("Parsing token: %s\n", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("Token parse error: %v\n", err)
		return nil, fmt.Errorf("token解析失败: %v", err)
	}

	if !token.Valid {
		fmt.Printf("Token is invalid\n")
		return nil, fmt.Errorf("token已失效")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		fmt.Printf("Failed to parse claims\n")
		return nil, fmt.Errorf("token格式错误")
	}

	// 验证必要字段
	if claims.UserID == 0 {
		fmt.Printf("Invalid user ID: %d\n", claims.UserID)
		return nil, fmt.Errorf("用户ID无效")
	}
	if claims.TenantID == 0 {
		fmt.Printf("Invalid tenant ID: %d\n", claims.TenantID)
		return nil, fmt.Errorf("租户ID无效")
	}
	if claims.Username == "" {
		fmt.Printf("Invalid username: %s\n", claims.Username)
		return nil, fmt.Errorf("用户名无效")
	}

	// 验证token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		fmt.Printf("Token expired at: %v\n", claims.ExpiresAt)
		return nil, fmt.Errorf("token已过期")
	}

	fmt.Printf("Token parsed successfully: user_id=%d, username=%s, tenant_id=%d, exp=%v\n",
		claims.UserID, claims.Username, claims.TenantID, claims.ExpiresAt)

	return claims, nil
}
