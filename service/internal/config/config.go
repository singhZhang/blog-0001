// Package config 从环境变量读取服务配置。
package config

import (
	"os"
	"strings"
)

// Config 汇总服务启动所需的全部配置。
type Config struct {
	// Addr 是 HTTP 监听地址，如 ":8080"。
	Addr string
	// ContentDir 是 Markdown 文章目录，仅在数据库为空时用于导入初始文章。
	ContentDir string
	// AllowedOrigins 是 CORS 允许的来源列表，"*" 表示全部放行。
	AllowedOrigins []string
	// AdminToken 是管理接口的 Bearer 令牌；为空表示禁用全部管理接口。
	AdminToken string
	// MySQLHost 是 MySQL 主机。
	MySQLHost string
	// MySQLPort 是 MySQL 端口。
	MySQLPort string
	// MySQLUser 是 MySQL 用户。
	MySQLUser string
	// MySQLPassword 是 MySQL 密码。
	MySQLPassword string
	// MySQLDatabase 是业务库名，不存在时由服务启动时创建。
	MySQLDatabase string
}

// Load 从环境变量读取配置，未设置时使用适合本地开发的默认值。
func Load() Config {
	return Config{
		Addr:           ":" + envOr("PORT", "8080"),
		ContentDir:     envOr("BLOG_CONTENT_DIR", "content/articles"),
		AllowedOrigins: splitList(envOr("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		AdminToken:     envOr("ADMIN_TOKEN", "blog-admin"),
		MySQLHost:      envOr("MYSQL_HOST", "47.94.248.206"),
		MySQLPort:      envOr("MYSQL_PORT", "3306"),
		MySQLUser:      envOr("MYSQL_USER", "root"),
		MySQLPassword:  envOr("MYSQL_PASSWORD", ""),
		MySQLDatabase:  envOr("MYSQL_DATABASE", "blog"),
	}
}

// envOr 返回环境变量的值（去除首尾空白），未设置或为空时返回 fallback。
func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return fallback
}

// splitList 按逗号拆分列表并去除空白项。
func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
