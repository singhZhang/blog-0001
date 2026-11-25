package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// MySQLStore 把主题存在 MySQL 中。改名时同步更新文章上的分类字段。
type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore 使用已经完成建表的连接。
func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

// Count 返回主题总数，用于启动日志。
func (s *MySQLStore) Count() int {
	n, err := s.CountContext(context.Background())
	if err != nil {
		return 0
	}
	return n
}

// CountContext 返回主题总数。
func (s *MySQLStore) CountContext(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// List 返回按名称排序的主题。
func (s *MySQLStore) List(ctx context.Context) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, description FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Name, &c.Description); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// Create 校验并写入一个新主题。
func (s *MySQLStore) Create(ctx context.Context, input Category) (Category, error) {
	c := trimCategory(input)
	if err := c.validate(); err != nil {
		return Category{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO categories (name, description) VALUES (?, ?)`, c.Name, c.Description)
	if isDuplicate(err) {
		return Category{}, fmt.Errorf("%w: name %q 已存在", ErrDuplicateName, c.Name)
	}
	if err != nil {
		return Category{}, err
	}
	return c, nil
}

// Ensure 在主题不存在时写入。已存在时保留原来的描述。
func (s *MySQLStore) Ensure(ctx context.Context, input Category) error {
	c := trimCategory(input)
	if c.Description == "" {
		c.Description = fmt.Sprintf("关于 %s 的记录与思考。", c.Name)
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO categories (name, description) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE name = name`, c.Name, c.Description)
	return err
}

// Update 以 name 定位并整体替换一个主题。改名时同步文章分类。
func (s *MySQLStore) Update(ctx context.Context, name string, input Category) (Category, error) {
	c := trimCategory(input)
	if err := c.validate(); err != nil {
		return Category{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Category{}, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE name = ?`, name).Scan(&exists); err != nil {
		return Category{}, err
	}
	if exists == 0 {
		return Category{}, ErrNotFound
	}
	if c.Name != name {
		var dup int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE name = ?`, c.Name).Scan(&dup); err != nil {
			return Category{}, err
		}
		if dup > 0 {
			return Category{}, fmt.Errorf("%w: name %q 已存在", ErrDuplicateName, c.Name)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE categories SET name = ?, description = ? WHERE name = ?`, c.Name, c.Description, name); err != nil {
		if isDuplicate(err) {
			return Category{}, fmt.Errorf("%w: name %q 已存在", ErrDuplicateName, c.Name)
		}
		return Category{}, err
	}
	if c.Name != name {
		if _, err := tx.ExecContext(ctx, `UPDATE articles SET category = ? WHERE category = ?`, c.Name, name); err != nil {
			return Category{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Category{}, err
	}
	return c, nil
}

// Delete 删除一个主题。仍有文章使用时拒绝删除。
func (s *MySQLStore) Delete(ctx context.Context, name string) error {
	var articles int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles WHERE category = ?`, name).Scan(&articles); err != nil {
		return err
	}
	if articles > 0 {
		return fmt.Errorf("%w: 仍有 %d 篇文章使用该主题，请先调整文章分类", ErrInvalid, articles)
	}

	result, err := s.db.ExecContext(ctx, `DELETE FROM categories WHERE name = ?`, name)
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

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
