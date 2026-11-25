package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/singhZhang/blog-0001/service/internal/article"
	"github.com/singhZhang/blog-0001/service/internal/category"
)

// articleStore 声明接口层需要的最小存储能力。
// 接口定义在使用方而非实现方（Go 惯例），article.FileStore 天然满足。
type articleStore interface {
	List(ctx context.Context) ([]article.Article, error)
	GetBySlug(ctx context.Context, slug string) (article.Article, error)
	Create(ctx context.Context, input article.Article) (article.Article, error)
	Update(ctx context.Context, slug string, input article.Article) (article.Article, error)
	Delete(ctx context.Context, slug string) error
}

// categoryStore 声明接口层需要的最小分类存储能力，category.FileStore 天然满足。
type categoryStore interface {
	List(ctx context.Context) ([]category.Category, error)
	Create(ctx context.Context, input category.Category) (category.Category, error)
	Update(ctx context.Context, name string, input category.Category) (category.Category, error)
	Delete(ctx context.Context, name string) error
}

// envelope 是列表接口的响应外层。前端对裸数组和 {data: [...]}
// 两种形式都兼容，统一采用后者以便后续扩展分页等字段。
type envelope struct {
	Data any `json:"data"`
}

// handleListArticles 返回全部文章的元数据列表（不含正文），存储层已按发布日期倒序。
func (a *API) handleListArticles(w http.ResponseWriter, r *http.Request) {
	rows, err := a.articles.List(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: rows})
}

// handleGetArticle 返回单篇文章（含 Markdown 正文）。
// 前端约定：404 时展示“文章缺失”页面，而不是报错。
func (a *API) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	found, err := a.articles.GetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		if errors.Is(err, article.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "文章不存在或已下线。")
			return
		}
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, found)
}

// handleListCategories 返回全部分类，存储层已按名称排序。
func (a *API) handleListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := a.categories.List(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: rows})
}

// handleHealth 提供存活探针。
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleNotFound 兜底处理未匹配到路由的 GET 请求。
func (a *API) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "接口不存在。")
}

// ---- 管理接口 ----

// maxBodyBytes 限制管理接口请求体大小。Markdown 正文即使很长也远小于此值，
// 超限基本意味着调用方传错了内容，及时拒绝而不是把内存和磁盘耗在无效请求上。
const maxBodyBytes = 2 << 20 // 2 MiB

// decodeJSON 把请求体解析到 v。超限或格式非法时写出 400 并返回 false。
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "请求体必须是合法的 JSON 且不超过 2 MiB。")
		return false
	}
	return true
}

// errorDetail 取出哨兵包装错误中的具体原因，
// 如 "invalid article: 缺少 title" → "缺少 title"，供管理界面定位问题。
// 存储层以 fmt.Errorf("%w: %s", sentinel, detail) 包装，这里做对应的反转。
func errorDetail(err error, sentinel error) string {
	return strings.TrimPrefix(err.Error(), sentinel.Error()+": ")
}

// handleCreateArticle 新增一篇文章。ID 与阅读时长由存储层生成，
// 请求体中的对应字段不生效。
func (a *API) handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var input article.Article
	if !decodeJSON(w, r, &input) {
		return
	}
	created, err := a.articles.Create(r.Context(), input)
	if err != nil {
		a.writeArticleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// handleUpdateArticle 以 URL 中的 slug 定位并整体替换一篇文章。
func (a *API) handleUpdateArticle(w http.ResponseWriter, r *http.Request) {
	var input article.Article
	if !decodeJSON(w, r, &input) {
		return
	}
	updated, err := a.articles.Update(r.Context(), r.PathValue("slug"), input)
	if err != nil {
		a.writeArticleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleDeleteArticle 删除一篇文章。
func (a *API) handleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	if err := a.articles.Delete(r.Context(), r.PathValue("slug")); err != nil {
		a.writeArticleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeArticleError 把文章存储的已知错误映射为对应状态码，
// 其余按内部错误处理。校验细节透传给管理界面，方便定位问题。
func (a *API) writeArticleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, article.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "文章不存在或已下线。")
	case errors.Is(err, article.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_article", errorDetail(err, article.ErrInvalid))
	case errors.Is(err, article.ErrDuplicateSlug):
		writeError(w, http.StatusConflict, "duplicate_slug", errorDetail(err, article.ErrDuplicateSlug))
	default:
		a.serverError(w, r, err)
	}
}

// handleCreateCategory 新增一个分类。
func (a *API) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input category.Category
	if !decodeJSON(w, r, &input) {
		return
	}
	created, err := a.categories.Create(r.Context(), input)
	if err != nil {
		a.writeCategoryError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// handleUpdateCategory 以 URL 中的 name 定位并整体替换一个分类。
func (a *API) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	var input category.Category
	if !decodeJSON(w, r, &input) {
		return
	}
	updated, err := a.categories.Update(r.Context(), r.PathValue("name"), input)
	if err != nil {
		a.writeCategoryError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleDeleteCategory 删除一个分类。
func (a *API) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	if err := a.categories.Delete(r.Context(), r.PathValue("name")); err != nil {
		a.writeCategoryError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeCategoryError 把分类存储的已知错误映射为对应状态码，其余按内部错误处理。
func (a *API) writeCategoryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, category.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "分类不存在。")
	case errors.Is(err, category.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_category", errorDetail(err, category.ErrInvalid))
	case errors.Is(err, category.ErrDuplicateName):
		writeError(w, http.StatusConflict, "duplicate_name", errorDetail(err, category.ErrDuplicateName))
	default:
		a.serverError(w, r, err)
	}
}

// serverError 记录内部错误并返回统一的 500 响应。
func (a *API) serverError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Error("处理请求失败", "method", r.Method, "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, "internal_error", "服务内部错误，请稍后重试。")
}
