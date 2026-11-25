// 命令 server 启动博客服务的 HTTP 接口。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/singhZhang/blog-0001/service/internal/api"
	"github.com/singhZhang/blog-0001/service/internal/article"
	"github.com/singhZhang/blog-0001/service/internal/category"
	"github.com/singhZhang/blog-0001/service/internal/config"
	"github.com/singhZhang/blog-0001/service/internal/db"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("服务异常退出", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	conn, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	articles := article.NewMySQLStore(conn)
	categories := category.NewMySQLStore(conn)
	if err := seedContent(ctx, conn, articles, categories, cfg.ContentDir, logger); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.New(articles, categories, cfg.AdminToken, logger).Handler(cfg.AllowedOrigins),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 先同步监听端口，让绑定失败在启动日志之前暴露。
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("监听端口 %s: %w", cfg.Addr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ln)
	}()

	logger.Info("博客服务已启动",
		"addr", ln.Addr().String(),
		"articles", articles.Count(),
		"categories", categories.Count(),
		"mysql", cfg.MySQLDatabase,
	)

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("HTTP 服务退出: %w", err)
	case <-ctx.Done():
		// 收到退出信号，等待存量请求处理完再退出。
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("优雅停机失败: %w", err)
	}

	logger.Info("服务已退出")
	return nil
}
