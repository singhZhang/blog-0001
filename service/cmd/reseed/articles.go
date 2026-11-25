package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/singhZhang/blog-0001/service/internal/article"
)

//go:embed posts/*.md
var postFS embed.FS

func articleList() ([]article.Article, error) {
	entries, err := fs.ReadDir(postFS, "posts")
	if err != nil {
		return nil, err
	}
	items := make([]article.Article, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		raw, err := postFS.ReadFile("posts/" + entry.Name())
		if err != nil {
			return nil, err
		}
		item, err := parsePost(string(raw), entry.Name())
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func parsePost(raw, name string) (article.Article, error) {
	text := strings.TrimPrefix(strings.TrimSpace(raw), "---")
	head, body, ok := strings.Cut(text, "\n---")
	if !ok {
		return article.Article{}, fmt.Errorf("%s 缺少正文分隔", name)
	}
	item := article.Article{Content: strings.TrimSpace(body)}
	for _, line := range strings.Split(head, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "slug":
			item.Slug = value
		case "title":
			item.Title = value
		case "publishedAt":
			item.PublishedAt = value
		case "category":
			item.Category = value
		case "tags":
			for _, tag := range strings.Split(value, ",") {
				if tag = strings.TrimSpace(tag); tag != "" {
					item.Tags = append(item.Tags, tag)
				}
			}
		}
	}
	if item.Tags == nil {
		item.Tags = []string{}
	}
	return item, nil
}
