package tests

import (
	"strings"
	"testing"

	"github.com/singhZhang/blog-0001/service/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// 设为空串既屏蔽本机可能存在的环境变量，也覆盖 envOr 的空值分支，保证用例可重复。
	t.Setenv("PORT", "")
	t.Setenv("BLOG_CONTENT_DIR", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("ADMIN_TOKEN", "")
	t.Setenv("MYSQL_HOST", "")
	t.Setenv("MYSQL_PORT", "")
	t.Setenv("MYSQL_USER", "")
	t.Setenv("MYSQL_PASSWORD", "")
	t.Setenv("MYSQL_DATABASE", "")

	cfg := config.Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, 期望 :8080", cfg.Addr)
	}
	if cfg.ContentDir != "content/articles" {
		t.Fatalf("ContentDir = %q, 期望 content/articles", cfg.ContentDir)
	}
	want := "http://localhost:5173,http://127.0.0.1:5173"
	if got := strings.Join(cfg.AllowedOrigins, ","); got != want {
		t.Fatalf("AllowedOrigins = %q, 期望 %q", got, want)
	}
	if cfg.MySQLHost != "47.94.248.206" || cfg.MySQLPort != "3306" || cfg.MySQLUser != "root" || cfg.MySQLPassword != "" || cfg.MySQLDatabase != "blog" {
		t.Fatalf("MySQL 默认配置 = %s:%s %s/%s@%s", cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLUser, cfg.MySQLPassword, cfg.MySQLDatabase)
	}
	if cfg.AdminToken != "blog-admin" {
		t.Fatalf("AdminToken = %q, 期望 blog-admin", cfg.AdminToken)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("PORT", " 9000 ")
	t.Setenv("BLOG_CONTENT_DIR", "/data/articles")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example ,,https://c.example")

	cfg := config.Load()
	if cfg.Addr != ":9000" {
		t.Fatalf("Addr = %q, 期望 :9000（应去除首尾空白）", cfg.Addr)
	}
	if cfg.ContentDir != "/data/articles" {
		t.Fatalf("ContentDir = %q, 期望 /data/articles", cfg.ContentDir)
	}
	want := "https://a.example,https://b.example,https://c.example"
	if got := strings.Join(cfg.AllowedOrigins, ","); got != want {
		t.Fatalf("AllowedOrigins = %q, 期望 %q（应去除空白项）", got, want)
	}
}

func TestLoadWildcardOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	cfg := config.Load()
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("AllowedOrigins = %v, 期望 [\"*\"]", cfg.AllowedOrigins)
	}
}
