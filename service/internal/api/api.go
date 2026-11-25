// Package api 实现博客服务的 HTTP 接口。
package api

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
)

// API 持有接口层的依赖：文章存储、分类存储与日志。
type API struct {
	articles   articleStore
	categories categoryStore
	adminToken string
	logger     *slog.Logger
}

// New 创建 API。store 只需满足本包声明的最小存储能力。
// adminToken 为空时，全部 /api/admin 路由返回 503。
func New(articles articleStore, categories categoryStore, adminToken string, logger *slog.Logger) *API {
	return &API{
		articles:   articles,
		categories: categories,
		adminToken: strings.TrimSpace(adminToken),
		logger:     logger,
	}
}

// Handler 组装路由与中间件，返回可直接交给 http.Server 的处理器。
// 中间件从外到内依次是：panic 恢复、请求日志、CORS。
func (a *API) Handler(allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/articles", a.handleListArticles)
	mux.HandleFunc("GET /api/articles/{slug}", a.handleGetArticle)
	mux.HandleFunc("GET /api/categories", a.handleListCategories)
	mux.HandleFunc("GET /api/health", a.handleHealth)

	mux.HandleFunc("GET /api/admin/session", a.requireAdmin(a.handleAdminSession))
	mux.HandleFunc("POST /api/admin/articles", a.requireAdmin(a.handleCreateArticle))
	mux.HandleFunc("PUT /api/admin/articles/{slug}", a.requireAdmin(a.handleUpdateArticle))
	mux.HandleFunc("DELETE /api/admin/articles/{slug}", a.requireAdmin(a.handleDeleteArticle))
	mux.HandleFunc("POST /api/admin/categories", a.requireAdmin(a.handleCreateCategory))
	mux.HandleFunc("PUT /api/admin/categories/{name}", a.requireAdmin(a.handleUpdateCategory))
	mux.HandleFunc("DELETE /api/admin/categories/{name}", a.requireAdmin(a.handleDeleteCategory))

	// 兜底路由限定 GET 方法：未知 GET 路径走 JSON 404，
	// 其他方法的请求由 ServeMux 自动返回 405（带 Allow 头）。
	mux.HandleFunc("GET /", a.handleNotFound)

	var h http.Handler = mux
	h = withCORS(h, allowedOrigins)
	h = withRequestLog(a.logger, h)
	h = withRecover(a.logger, h)
	return h
}

// handleAdminSession 供后台登录页校验令牌。令牌正确时返回 200。
func (a *API) handleAdminSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// requireAdmin 校验 Authorization: Bearer <token>。
// 未配置令牌时管理接口整体关闭，避免空口令把写接口暴露出去。
func (a *API) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.adminToken == "" {
			writeError(w, http.StatusServiceUnavailable, "admin_disabled", "管理接口未启用。")
			return
		}
		const scheme = "Bearer "
		header := r.Header.Get("Authorization")
		if len(header) <= len(scheme) || header[:len(scheme)] != scheme ||
			subtle.ConstantTimeCompare([]byte(header[len(scheme):]), []byte(a.adminToken)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="blog-admin"`)
			writeError(w, http.StatusUnauthorized, "unauthorized", "管理令牌无效。")
			return
		}
		next(w, r)
	}
}
