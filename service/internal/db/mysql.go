// Package db 打开 MySQL 并确保库表存在。
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/singhZhang/blog-0001/service/internal/config"
)

// identPattern 限制库名为普通标识符，避免拼进 CREATE DATABASE 时被注入。
var identPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const metaTable = `
CREATE TABLE IF NOT EXISTS app_meta (
  name VARCHAR(64) NOT NULL,
  value VARCHAR(255) NOT NULL,
  PRIMARY KEY (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const categoriesTable = `
CREATE TABLE IF NOT EXISTS categories (
  name VARCHAR(191) NOT NULL,
  description TEXT NOT NULL,
  PRIMARY KEY (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const articlesTable = `
CREATE TABLE IF NOT EXISTS articles (
  id VARCHAR(64) NOT NULL,
  slug VARCHAR(191) NOT NULL,
  title VARCHAR(500) NOT NULL,
  excerpt TEXT NOT NULL,
  published_at CHAR(10) NOT NULL,
  category VARCHAR(191) NOT NULL,
  read_minutes INT NOT NULL,
  tags JSON NOT NULL,
  featured_code TEXT NULL,
  content MEDIUMTEXT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_articles_slug (slug),
  KEY idx_articles_published (published_at, slug),
  KEY idx_articles_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

// Open 连接 MySQL。库已存在时直接使用；只有库不存在时才创建，并补齐文章与主题表。
func Open(cfg config.Config) (*sql.DB, error) {
	if !identPattern.MatchString(cfg.MySQLDatabase) {
		return nil, fmt.Errorf("MYSQL_DATABASE %q 只允许字母、数字和下划线", cfg.MySQLDatabase)
	}

	server := mysql.Config{
		User:                 cfg.MySQLUser,
		Passwd:               cfg.MySQLPassword,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.MySQLHost, cfg.MySQLPort),
		DBName:               cfg.MySQLDatabase,
		Params:               map[string]string{"charset": "utf8mb4"},
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	conn, err := openServer(server)
	if err != nil {
		if !isUnknownDatabase(err) {
			return nil, fmt.Errorf("连接数据库 %s: %w", cfg.MySQLDatabase, err)
		}
		if err := createDatabase(server, cfg.MySQLDatabase); err != nil {
			return nil, err
		}
		conn, err = openServer(server)
		if err != nil {
			return nil, fmt.Errorf("连接数据库 %s: %w", cfg.MySQLDatabase, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := migrate(ctx, conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func openServer(cfg mysql.Config) (*sql.DB, error) {
	conn, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func createDatabase(server mysql.Config, name string) error {
	server.DBName = ""
	bootstrap, err := openServer(server)
	if err != nil {
		return fmt.Errorf("连接 MySQL %s: %w", server.Addr, err)
	}
	defer bootstrap.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_, err = bootstrap.ExecContext(ctx,
		"CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil && !isDatabaseExists(err) {
		return fmt.Errorf("创建数据库 %s: %w", name, err)
	}
	return nil
}

func isUnknownDatabase(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1049
}

func isDatabaseExists(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1007
}

func migrate(ctx context.Context, conn *sql.DB) error {
	for _, stmt := range []struct {
		name string
		sql  string
	}{
		{"元数据表", metaTable},
		{"分类表", categoriesTable},
		{"文章表", articlesTable},
	} {
		if _, err := conn.ExecContext(ctx, stmt.sql); err != nil {
			return fmt.Errorf("创建%s: %w", stmt.name, err)
		}
	}
	return nil
}

// ContentSeeded 报告是否已经做过一次初始内容导入。
// 导入只发生一次，避免管理员删光文章后重启又把 Markdown 灌回去。
func ContentSeeded(ctx context.Context, conn *sql.DB) (bool, error) {
	var n int
	err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM app_meta WHERE name = ? AND value = ?`, "content_seeded", "1").Scan(&n)
	return n > 0, err
}

// MarkContentSeeded 记下初始导入已经完成。
func MarkContentSeeded(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `
		INSERT INTO app_meta (name, value) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value)`, "content_seeded", "1")
	return err
}
