import test from 'node:test'
import assert from 'node:assert/strict'
import { buildArticleUrl, buildArticlesUrl, extractArticle, extractArticles } from '../src/lib/data-source.mjs'

test('builds a stable articles endpoint from relative and absolute bases', () => {
  assert.equal(buildArticlesUrl('/api/'), '/api/articles')
  assert.equal(buildArticlesUrl('https://example.test/api'), 'https://example.test/api/articles')
})

test('accepts an article array or a data envelope', () => {
  const rows = [{ id: 'article-1' }]
  assert.deepEqual(extractArticles(rows), rows)
  assert.deepEqual(extractArticles({ data: rows }), rows)
})

test('rejects malformed API payloads', () => {
  assert.throws(() => extractArticles({ items: [] }), /文章数据格式无效/)
})

test('builds an encoded single-article endpoint', () => {
  assert.equal(buildArticleUrl('/api/', 'reliable-data-pipeline'), '/api/articles/reliable-data-pipeline')
  assert.equal(buildArticleUrl('/api', '慢查询 索引'), '/api/articles/%E6%85%A2%E6%9F%A5%E8%AF%A2%20%E7%B4%A2%E5%BC%95')
})

test('accepts a single article or a data envelope', () => {
  const row = { id: 'article-1', slug: 'article-1' }
  assert.deepEqual(extractArticle(row), row)
  assert.deepEqual(extractArticle({ data: row }), row)
  assert.throws(() => extractArticle([row]), /文章数据格式无效/)
  assert.throws(() => extractArticle({ data: null }), /文章数据格式无效/)
})
