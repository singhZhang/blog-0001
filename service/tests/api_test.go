package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/singhZhang/blog-0001/service/internal/api"
	"github.com/singhZhang/blog-0001/service/internal/article"
	"github.com/singhZhang/blog-0001/service/internal/category"
)

// discardLogger 返回一个丢弃输出的日志器，避免测试刷屏。
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// stubStore 是 api 包 articleStore 接口的测试替身：articles 是正常返回的数据，
// listErr/getErr 模拟存储故障，panicOnList 模拟业务代码崩溃。
type stubStore struct {
	articles    []article.Article
	listErr     error
	getErr      error
	panicOnList bool
	created     article.Article
}

func (s *stubStore) List(ctx context.Context) ([]article.Article, error) {
	if s.panicOnList {
		panic("模拟业务代码 panic")
	}
	if s.listErr != nil {
		return nil, s.listErr
	}
	rows := make([]article.Article, len(s.articles))
	copy(rows, s.articles)
	return rows, nil
}

func (s *stubStore) GetBySlug(ctx context.Context, slug string) (article.Article, error) {
	if s.getErr != nil {
		return article.Article{}, s.getErr
	}
	for _, a := range s.articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return article.Article{}, article.ErrNotFound
}

func (s *stubStore) Create(ctx context.Context, input article.Article) (article.Article, error) {
	input.ID = "generated"
	if input.Tags == nil {
		input.Tags = []string{}
	}
	s.created = input
	return input, nil
}

func (s *stubStore) Update(ctx context.Context, slug string, input article.Article) (article.Article, error) {
	return article.Article{}, errors.New("not implemented")
}

func (s *stubStore) Delete(ctx context.Context, slug string) error {
	return errors.New("not implemented")
}

// stubCategories 是分类存储的测试替身。
type stubCategories struct {
	rows []category.Category
	err  error
}

func (s *stubCategories) List(ctx context.Context) ([]category.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.rows == nil {
		return []category.Category{}, nil
	}
	return s.rows, nil
}

func (s *stubCategories) Create(ctx context.Context, input category.Category) (category.Category, error) {
	return input, nil
}

func (s *stubCategories) Update(ctx context.Context, name string, input category.Category) (category.Category, error) {
	return input, nil
}

func (s *stubCategories) Delete(ctx context.Context, name string) error {
	return nil
}

// apiTestArticles 返回两篇不含正文的文章，模拟 FileStore.List 的返回。
func apiTestArticles() []article.Article {
	second := validArticle()
	second.ID = "002"
	second.Slug = "go-generics-notes"
	second.Title = "Go 泛型学习笔记"
	return []article.Article{validArticle(), second}
}

// newTestHandler 用替身存储构建完整的 HTTP 处理链（含全部中间件）。
func newTestHandler(store *stubStore, allowedOrigins []string) http.Handler {
	return api.New(store, &stubCategories{}, "", discardLogger()).Handler(allowedOrigins)
}

// doRequest 发起一次请求并返回响应与响应体，headers 为额外请求头。
func doRequest(t *testing.T, h http.Handler, method, target string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("读取响应体: %v", err)
	}
	return res, string(body)
}

// errorCode 解析统一错误结构 {"error":{"code","message"}} 并返回其中的 code。
func errorCode(t *testing.T, body string) string {
	t.Helper()
	var parsed struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("解析错误响应 %q: %v", body, err)
	}
	return parsed.Error.Code
}

func TestListArticles(t *testing.T) {
	h := newTestHandler(&stubStore{articles: apiTestArticles()}, nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/articles", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200, 响应: %s", res.StatusCode, body)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, 期望 application/json", ct)
	}

	var parsed struct {
		Data []article.Article `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("解析列表响应 %q: %v", body, err)
	}
	if len(parsed.Data) != 2 {
		t.Fatalf("文章数量 = %d, 期望 2", len(parsed.Data))
	}
	if parsed.Data[0].Slug != "understand-query-index-again" || parsed.Data[1].Slug != "go-generics-notes" {
		t.Fatalf("文章顺序不符: %+v", parsed.Data)
	}
}

func TestGetArticle(t *testing.T) {
	withBody := validArticle()
	withBody.Content = "这是完整正文。"
	h := newTestHandler(&stubStore{articles: []article.Article{withBody}}, nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/articles/understand-query-index-again", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200, 响应: %s", res.StatusCode, body)
	}

	var got article.Article
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("解析详情响应 %q: %v", body, err)
	}
	if got.Title != "重新理解查询索引" || got.Content != "这是完整正文。" {
		t.Fatalf("详情应包含标题与正文, 实际: %+v", got)
	}
}

func TestGetArticleNotFound(t *testing.T) {
	h := newTestHandler(&stubStore{articles: apiTestArticles()}, nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/articles/missing", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("状态码 = %d, 期望 404, 响应: %s", res.StatusCode, body)
	}
	if code := errorCode(t, body); code != "not_found" {
		t.Fatalf("错误码 = %q, 期望 not_found", code)
	}
}

func TestStoreErrorsReturn500(t *testing.T) {
	storeErr := errors.New("模拟存储故障")
	h := newTestHandler(&stubStore{listErr: storeErr, getErr: storeErr}, nil)

	for _, target := range []string{"/api/articles", "/api/articles/whatever"} {
		res, body := doRequest(t, h, http.MethodGet, target, nil)
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("%s 状态码 = %d, 期望 500, 响应: %s", target, res.StatusCode, body)
		}
		if code := errorCode(t, body); code != "internal_error" {
			t.Fatalf("%s 错误码 = %q, 期望 internal_error", target, code)
		}
	}
}

func TestPanicRecovery(t *testing.T) {
	h := newTestHandler(&stubStore{panicOnList: true}, nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/articles", nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("状态码 = %d, 期望 500, 响应: %s", res.StatusCode, body)
	}
	if code := errorCode(t, body); code != "internal_error" {
		t.Fatalf("错误码 = %q, 期望 internal_error", code)
	}
}

func TestHealth(t *testing.T) {
	h := newTestHandler(&stubStore{}, nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/health", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200, 响应: %s", res.StatusCode, body)
	}
	var got map[string]string
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("解析健康检查响应 %q: %v", body, err)
	}
	if got["status"] != "ok" {
		t.Fatalf("status = %q, 期望 ok", got["status"])
	}
}

func TestUnknownPath(t *testing.T) {
	h := newTestHandler(&stubStore{articles: apiTestArticles()}, nil)

	for _, target := range []string{"/api/nope", "/"} {
		res, body := doRequest(t, h, http.MethodGet, target, nil)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("%s 状态码 = %d, 期望 404, 响应: %s", target, res.StatusCode, body)
		}
		if code := errorCode(t, body); code != "not_found" {
			t.Fatalf("%s 错误码 = %q, 期望 not_found", target, code)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestHandler(&stubStore{articles: apiTestArticles()}, nil)

	res, body := doRequest(t, h, http.MethodPost, "/api/articles", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("状态码 = %d, 期望 405, 响应: %s", res.StatusCode, body)
	}
	if allow := res.Header.Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("Allow 头 = %q, 期望包含 GET", allow)
	}
}

func TestCORS(t *testing.T) {
	allowed := []string{"http://localhost:5173", "http://127.0.0.1:5173"}

	t.Run("允许的来源回显", func(t *testing.T) {
		h := newTestHandler(&stubStore{articles: apiTestArticles()}, allowed)
		res, _ := doRequest(t, h, http.MethodGet, "/api/health",
			map[string]string{"Origin": "http://localhost:5173"})
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Fatalf("Access-Control-Allow-Origin = %q, 期望回显来源", got)
		}
	})

	t.Run("未允许的来源不回显但保留 Vary", func(t *testing.T) {
		h := newTestHandler(&stubStore{articles: apiTestArticles()}, allowed)
		res, _ := doRequest(t, h, http.MethodGet, "/api/health",
			map[string]string{"Origin": "http://evil.example"})
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("未允许的来源不应设置 Access-Control-Allow-Origin, 实际: %q", got)
		}
		if vary := res.Header.Get("Vary"); !strings.Contains(vary, "Origin") {
			t.Fatalf("Vary 头 = %q, 期望包含 Origin", vary)
		}
	})

	t.Run("未带 Origin 不设置 CORS 头", func(t *testing.T) {
		h := newTestHandler(&stubStore{articles: apiTestArticles()}, allowed)
		res, _ := doRequest(t, h, http.MethodGet, "/api/health", nil)
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("未带 Origin 不应设置 Access-Control-Allow-Origin, 实际: %q", got)
		}
	})

	t.Run("预检请求直接返回 204", func(t *testing.T) {
		h := newTestHandler(&stubStore{articles: apiTestArticles()}, allowed)
		res, _ := doRequest(t, h, http.MethodOptions, "/api/articles", map[string]string{
			"Origin":                        "http://localhost:5173",
			"Access-Control-Request-Method": "GET",
		})
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("预检状态码 = %d, 期望 204", res.StatusCode)
		}
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Fatalf("Access-Control-Allow-Origin = %q, 期望回显来源", got)
		}
		if methods := res.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(methods, http.MethodGet) {
			t.Fatalf("Access-Control-Allow-Methods = %q, 期望包含 GET", methods)
		}
	})

	t.Run("通配符放行全部来源", func(t *testing.T) {
		h := newTestHandler(&stubStore{articles: apiTestArticles()}, []string{"*"})
		res, _ := doRequest(t, h, http.MethodGet, "/api/health",
			map[string]string{"Origin": "http://any.example"})
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("Access-Control-Allow-Origin = %q, 期望 *", got)
		}
	})
}

func TestRequestLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	h := api.New(&stubStore{articles: apiTestArticles()}, &stubCategories{}, "", logger).Handler(nil)

	res, _ := doRequest(t, h, http.MethodGet, "/api/articles", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200", res.StatusCode)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "/api/articles") || !strings.Contains(logOutput, "status=200") {
		t.Fatalf("请求日志应包含路径与状态码, 实际: %q", logOutput)
	}
}

func doJSON(t *testing.T, h http.Handler, method, target, payload string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("读取响应体: %v", err)
	}
	return res, string(body)
}

func TestListCategories(t *testing.T) {
	h := api.New(&stubStore{}, &stubCategories{rows: []category.Category{
		{Name: "Data", Description: "数据"},
	}}, "", discardLogger()).Handler(nil)

	res, body := doRequest(t, h, http.MethodGet, "/api/categories", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200, 响应: %s", res.StatusCode, body)
	}
	var parsed struct {
		Data []category.Category `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("解析主题响应 %q: %v", body, err)
	}
	if len(parsed.Data) != 1 || parsed.Data[0].Name != "Data" {
		t.Fatalf("主题列表不符: %+v", parsed.Data)
	}
}

func TestAdminRoutes(t *testing.T) {
	store := &stubStore{}
	h := api.New(store, &stubCategories{}, "secret", discardLogger()).Handler(nil)

	t.Run("缺少令牌", func(t *testing.T) {
		res, body := doJSON(t, h, http.MethodPost, "/api/admin/articles", `{"title":"不应写入"}`, nil)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("状态码 = %d, 期望 401, 响应: %s", res.StatusCode, body)
		}
		if store.created.Title != "" {
			t.Fatal("未授权请求不应调用存储")
		}
	})

	t.Run("令牌正确可创建", func(t *testing.T) {
		res, body := doJSON(t, h, http.MethodPost, "/api/admin/articles", `{"title":"新文章"}`, map[string]string{
			"Authorization": "Bearer secret",
		})
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("状态码 = %d, 期望 201, 响应: %s", res.StatusCode, body)
		}
		if store.created.Title != "新文章" || store.created.ID != "generated" {
			t.Fatalf("写入内容不符: %+v", store.created)
		}
	})

	t.Run("未配置令牌时关闭管理接口", func(t *testing.T) {
		disabled := api.New(&stubStore{}, &stubCategories{}, "", discardLogger()).Handler(nil)
		res, body := doRequest(t, disabled, http.MethodGet, "/api/admin/session", nil)
		if res.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("状态码 = %d, 期望 503, 响应: %s", res.StatusCode, body)
		}
	})
}
