package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// withRecover 捕获下游处理器的 panic，记录堆栈后返回 500，避免单个请求拖垮整个进程。
func withRecover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				writeError(w, http.StatusInternalServerError, "internal_error", "服务内部错误，请稍后重试。")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder 包装 ResponseWriter 以捕获状态码，供访问日志使用。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// withRequestLog 为每个请求输出一行访问日志。
func withRequestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(start).Round(time.Millisecond).String(),
		)
	})
}

// withCORS 提供跨域支持：前端 Vite 开发服务器没有配置代理，需要跨端口直连。
// allowedOrigins 是允许的来源列表，"*" 表示全部放行；名单外的来源不设置允许头。
func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
		}
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, originAllowed := allowed[origin]
		// 响应是否放行取决于 Origin，告知缓存按 Origin 区分。
		w.Header().Add("Vary", "Origin")
		if origin != "" && (allowAll || originAllowed) {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// 预检请求在中间件层直接应答，不进入路由。
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
