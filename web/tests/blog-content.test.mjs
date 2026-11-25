import test from 'node:test'
import assert from 'node:assert/strict'
import {
  formatChineseDate,
  getAdjacentArticles,
  getRelatedArticles,
  buildTopics,
  groupArchive,
  groupTopics,
  toTopicSlug,
} from '../src/lib/blog-content.mjs'

const article = (slug, publishedAt, category, tags = []) => ({
  id: slug, slug, title: slug, excerpt: '', publishedAt, category, readMinutes: 5, tags,
})

const articles = [
  article('api', '2026-09-08', 'Architecture', ['API']),
  article('index', '2026-10-18', 'Data', ['MySQL', 'Performance']),
  article('perf', '2026-09-28', 'Frontend', ['Performance']),
  article('pipeline', '2026-10-12', 'Data', ['Pipeline']),
  article('notes', '2025-12-30', '生活 随笔'),
]

test('builds URL-safe topic slugs for latin and CJK names', () => {
  assert.equal(toTopicSlug('AI Systems'), 'ai-systems')
  assert.equal(toTopicSlug(' Data '), 'data')
  assert.equal(toTopicSlug('生活 随笔'), '生活-随笔')
})

test('formats ISO dates in Chinese without leading zeros', () => {
  assert.equal(formatChineseDate('2026-09-08'), '2026年9月8日')
})

test('groups the archive by year and month, newest first', () => {
  const archive = groupArchive(articles)
  assert.deepEqual(archive.map(({ year, count }) => [year, count]), [['2026', 4], ['2025', 1]])
  assert.deepEqual(archive[0].months.map(({ month, articles: rows }) => [month, rows.map((row) => row.slug)]), [
    ['10', ['index', 'pipeline']],
    ['09', ['perf', 'api']],
  ])
})

test('groups topics by size, then by most recent article', () => {
  const topics = groupTopics(articles)
  assert.deepEqual(topics.map(({ name, count }) => [name, count]), [
    ['Data', 2],
    ['Frontend', 1],
    ['Architecture', 1],
    ['生活 随笔', 1],
  ])
  assert.equal(topics[0].latest.slug, 'index')
  assert.equal(topics[0].slug, 'data')
})

test('keeps empty themes and uses category descriptions', () => {
  const topics = buildTopics([
    { name: 'Data', description: '数据' },
    { name: 'Empty', description: '还没有文章的主题' },
  ], articles)
  const data = topics.find((topic) => topic.name === 'Data')
  const empty = topics.find((topic) => topic.name === 'Empty')
  assert.equal(data?.description, '数据')
  assert.equal(data?.count, 2)
  assert.equal(empty?.count, 0)
  assert.equal(empty?.latest, null)
  assert.equal(topics.some((topic) => topic.name === 'Frontend'), true)
})

test('finds the newer and older neighbours of an article', () => {
  assert.deepEqual(
    Object.fromEntries(Object.entries(getAdjacentArticles(articles, 'pipeline')).map(([key, value]) => [key, value?.slug])),
    { newer: 'index', older: 'perf' },
  )
  assert.equal(getAdjacentArticles(articles, 'index').newer, null)
  assert.equal(getAdjacentArticles(articles, 'notes').older, null)
  assert.deepEqual(getAdjacentArticles(articles, 'missing'), { newer: null, older: null })
})

test('ranks related articles by category and shared tags, skipping unrelated ones', () => {
  const current = articles.find(({ slug }) => slug === 'index')
  assert.deepEqual(getRelatedArticles(articles, current).map(({ slug }) => slug), ['pipeline', 'perf'])
  assert.equal(getRelatedArticles(articles, current, 1).length, 1)
})
