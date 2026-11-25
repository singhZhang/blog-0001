// Package category 提供博客分类的领域模型与基于 YAML 文件的存储实现。
package category

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Category 的字段与文章元数据中的分类信息对应，name 同时是管理接口中的定位标识。
type Category struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// ErrNotFound 表示请求的分类不存在，接口层据此返回 404。
var ErrNotFound = errors.New("category not found")

// ErrDuplicateName 表示新增或改名后的分类名与现有分类冲突，接口层据此返回 409。
var ErrDuplicateName = errors.New("duplicate category name")

// ErrInvalid 表示分类校验不通过，接口层据此返回 400。
var ErrInvalid = errors.New("invalid category")

// fileDocument 对应分类文件的整体结构，根键为 categories。
type fileDocument struct {
	Categories []Category `yaml:"categories"`
}

// FileStore 在启动时一次性读取 YAML 文件并常驻内存。
// 管理接口的写操作在持锁状态下先落盘再更新内存，读者不会看到中间状态。
type FileStore struct {
	mu         sync.RWMutex
	path       string
	categories []Category
	byName     map[string]Category
}

// NewFileStore 读取并校验分类文件。文件不存在时返回空 store 并保留路径，
// 管理员创建第一个分类时会自动生成文件；文件存在但内容不合法则启动失败，
// 让配置错误在部署阶段暴露而不是在线上以 500 的形式出现。
func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{path: path, byName: make(map[string]Category)}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return store, nil
		}
		return nil, fmt.Errorf("读取分类文件 %s: %w", path, err)
	}

	var doc fileDocument
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("解析分类文件 %s: %w", path, err)
	}
	for _, c := range doc.Categories {
		if err := c.validate(); err != nil {
			return nil, fmt.Errorf("分类文件 %s: %w", path, err)
		}
		if _, exists := store.byName[c.Name]; exists {
			return nil, fmt.Errorf("分类名 %q 在 %s 中重复", c.Name, path)
		}
		store.categories = append(store.categories, c)
		store.byName[c.Name] = c
	}

	sortCategories(store.categories)
	return store, nil
}

// sortCategories 就地按名称排序。
func sortCategories(categories []Category) {
	slices.SortFunc(categories, func(a, b Category) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// validate 检查分类字段的完整性，返回第一个不满足的规则。
func (c Category) validate() error {
	switch {
	case c.Name == "":
		return fmt.Errorf("缺少 name")
	case c.Description == "":
		return fmt.Errorf("缺少 description")
	default:
		return nil
	}
}

// trimCategory 去除输入字段的首尾空白，避免 " Go " 与 "Go" 成为两个分类。
func trimCategory(c Category) Category {
	c.Name = strings.TrimSpace(c.Name)
	c.Description = strings.TrimSpace(c.Description)
	return c
}

// Count 返回分类总数，用于启动日志。
func (s *FileStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.categories)
}

// List 返回按名称排序的分类副本。
func (s *FileStore) List(ctx context.Context) ([]Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]Category, len(s.categories))
	copy(rows, s.categories)
	return rows, nil
}

// Create 校验并落盘一个新分类，返回去除首尾空白后的分类。
func (s *FileStore) Create(ctx context.Context, input Category) (Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := trimCategory(input)
	if err := c.validate(); err != nil {
		return Category{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if _, exists := s.byName[c.Name]; exists {
		return Category{}, fmt.Errorf("%w: name %q 已存在", ErrDuplicateName, c.Name)
	}

	next := append(s.clone(), c)
	if err := s.persist(next); err != nil {
		return Category{}, err
	}
	s.commit(next)
	return c, nil
}

// Update 以 name 定位并整体替换一个分类；不存在时返回 ErrNotFound。
// 允许修改分类名，改名后与其他分类冲突时返回 ErrDuplicateName。
func (s *FileStore) Update(ctx context.Context, name string, input Category) (Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byName[name]; !ok {
		return Category{}, ErrNotFound
	}

	c := trimCategory(input)
	if err := c.validate(); err != nil {
		return Category{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if c.Name != name {
		if _, exists := s.byName[c.Name]; exists {
			return Category{}, fmt.Errorf("%w: name %q 已存在", ErrDuplicateName, c.Name)
		}
	}

	next := make([]Category, 0, len(s.categories))
	for _, old := range s.categories {
		if old.Name != name {
			next = append(next, old)
		}
	}
	next = append(next, c)
	if err := s.persist(next); err != nil {
		return Category{}, err
	}
	s.commit(next)
	return c, nil
}

// Delete 删除一个分类；不存在时返回 ErrNotFound。
func (s *FileStore) Delete(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byName[name]; !ok {
		return ErrNotFound
	}

	next := make([]Category, 0, len(s.categories))
	for _, old := range s.categories {
		if old.Name != name {
			next = append(next, old)
		}
	}
	if err := s.persist(next); err != nil {
		return err
	}
	s.commit(next)
	return nil
}

// clone 返回当前分类列表的副本，作为构造候选状态的基础。
func (s *FileStore) clone() []Category {
	rows := make([]Category, len(s.categories))
	copy(rows, s.categories)
	return rows
}

// commit 把已落盘的候选状态写入内存索引并保持排序不变量。
func (s *FileStore) commit(categories []Category) {
	sortCategories(categories)
	s.categories = categories
	s.byName = make(map[string]Category, len(categories))
	for _, c := range categories {
		s.byName[c.Name] = c
	}
}

// persist 以原子方式把分类列表写成 YAML 文件：先写同目录临时文件并落盘，
// 再 rename 到目标路径，避免读者看到写了一半的内容。
func (s *FileStore) persist(categories []Category) error {
	data, err := yaml.Marshal(fileDocument{Categories: categories})
	if err != nil {
		return fmt.Errorf("序列化分类文件: %w", err)
	}

	dir := filepath.Dir(s.path)
	// 临时文件名不以 .yaml 结尾，且 NewFileStore 只按精确路径读取，崩溃残留不会被加载。
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("在 %s 创建临时文件: %w", dir, err)
	}
	tmpName := tmp.Name()
	// 成功路径上文件已被 rename 走，这里的 Remove 会命中 not exist，忽略即可。
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("写入 %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("落盘 %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭 %s: %w", tmpName, err)
	}
	// CreateTemp 默认权限 0600，改回 0644，保证宿主机挂载场景可读。
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("修改权限 %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("重命名 %s → %s: %w", tmpName, s.path, err)
	}
	return nil
}
