package article

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// MySQLStore 把文章存在 MySQL 中。列表接口不返回正文。
type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore 使用已经完成建表的连接。
func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

// Count 返回文章总数，用于启动日志。
func (s *MySQLStore) Count() int {
	n, err := s.CountContext(context.Background())
	if err != nil {
		return 0
	}
	return n
}

// CountContext 返回文章总数。
func (s *MySQLStore) CountContext(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// List 返回按发布日期倒序的文章副本，不含正文。
func (s *MySQLStore) List(ctx context.Context) ([]Article, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, title, excerpt, published_at, category, read_minutes, tags, featured_code, content
		FROM articles
		ORDER BY published_at DESC, slug ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]Article, 0)
	for rows.Next() {
		a, err := scanArticle(rows, true)
		if err != nil {
			return nil, err
		}
		if excerpt := ExcerptFromContent(a.Content); excerpt != "" {
			a.Excerpt = excerpt
		}
		a.Content = ""
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

// GetBySlug 返回文章全文；不存在时返回 ErrNotFound。
func (s *MySQLStore) GetBySlug(ctx context.Context, slug string) (Article, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, slug, title, excerpt, published_at, category, read_minutes, tags, featured_code, content
		FROM articles WHERE slug = ?`, slug)
	a, err := scanArticle(row, true)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, ErrNotFound
	}
	return a, err
}

// Create 校验并写入一篇新文章。ID、阅读时长和摘要由服务端生成。
func (s *MySQLStore) Create(ctx context.Context, input Article) (Article, error) {
	a := input
	a.ID = NewID()
	a.Tags = normalizeTags(a.Tags)
	a.ReadMinutes = EstimateReadMinutes(a.Content)
	a.Excerpt = ExcerptFromContent(a.Content)
	if err := a.Validate(); err != nil {
		return Article{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if err := s.requireCategory(ctx, a.Category); err != nil {
		return Article{}, err
	}
	if err := s.insert(ctx, s.db, a); err != nil {
		return Article{}, err
	}
	return a, nil
}

// Update 以 slug 定位并整体替换一篇文章。ID 保持不变。
func (s *MySQLStore) Update(ctx context.Context, slug string, input Article) (Article, error) {
	old, err := s.GetBySlug(ctx, slug)
	if err != nil {
		return Article{}, err
	}

	a := input
	a.ID = old.ID
	a.Tags = normalizeTags(a.Tags)
	a.ReadMinutes = EstimateReadMinutes(a.Content)
	a.Excerpt = ExcerptFromContent(a.Content)
	if err := a.Validate(); err != nil {
		return Article{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if err := s.requireCategory(ctx, a.Category); err != nil {
		return Article{}, err
	}
	if a.Slug != slug {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles WHERE slug = ?`, a.Slug).Scan(&n); err != nil {
			return Article{}, err
		}
		if n > 0 {
			return Article{}, fmt.Errorf("%w: slug %q 已存在", ErrDuplicateSlug, a.Slug)
		}
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE articles
		SET slug = ?, title = ?, excerpt = ?, published_at = ?, category = ?,
		    read_minutes = ?, tags = ?, featured_code = ?, content = ?
		WHERE slug = ?`,
		a.Slug, a.Title, a.Excerpt, a.PublishedAt, a.Category,
		a.ReadMinutes, mustJSON(a.Tags), nullIfEmpty(a.FeaturedCode), a.Content, slug,
	)
	if err != nil {
		if isDuplicate(err) {
			return Article{}, fmt.Errorf("%w: slug %q 已存在", ErrDuplicateSlug, a.Slug)
		}
		return Article{}, err
	}
	return a, nil
}

// Delete 删除一篇文章；不存在时返回 ErrNotFound。
func (s *MySQLStore) Delete(ctx context.Context, slug string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM articles WHERE slug = ?`, slug)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ImportAll 在一个事务里写入已有文章，保留文件里的 ID 与阅读时长。
// 任一篇失败则整批回滚，方便下次启动重试。
func (s *MySQLStore) ImportAll(ctx context.Context, articles []Article) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range articles {
		a.Tags = normalizeTags(a.Tags)
		if err := a.Validate(); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalid, err)
		}
		if err := s.insert(ctx, tx, a); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type sqlExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (s *MySQLStore) insert(ctx context.Context, exec sqlExec, a Article) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO articles (
			id, slug, title, excerpt, published_at, category, read_minutes, tags, featured_code, content
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Slug, a.Title, a.Excerpt, a.PublishedAt, a.Category,
		a.ReadMinutes, mustJSON(a.Tags), nullIfEmpty(a.FeaturedCode), a.Content,
	)
	if isDuplicate(err) {
		return fmt.Errorf("%w: slug %q 已存在", ErrDuplicateSlug, a.Slug)
	}
	return err
}

func (s *MySQLStore) requireCategory(ctx context.Context, name string) error {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%w: 分类 %q 不存在，请先创建主题", ErrInvalid, name)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanArticle(row scanner, withContent bool) (Article, error) {
	var (
		a        Article
		tagsRaw  []byte
		featured sql.NullString
	)
	dest := []any{&a.ID, &a.Slug, &a.Title, &a.Excerpt, &a.PublishedAt, &a.Category, &a.ReadMinutes, &tagsRaw, &featured}
	if withContent {
		dest = append(dest, &a.Content)
	}
	if err := row.Scan(dest...); err != nil {
		return Article{}, err
	}
	if len(tagsRaw) > 0 {
		if err := json.Unmarshal(tagsRaw, &a.Tags); err != nil {
			return Article{}, fmt.Errorf("解析标签: %w", err)
		}
	}
	a.Tags = normalizeTags(a.Tags)
	if featured.Valid {
		a.FeaturedCode = featured.String
	}
	return a, nil
}

func mustJSON(tags []string) []byte {
	raw, err := json.Marshal(normalizeTags(tags))
	if err != nil {
		return []byte("[]")
	}
	return raw
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
