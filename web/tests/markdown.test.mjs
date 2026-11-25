import test from 'node:test'
import assert from 'node:assert/strict'
import { renderMarkdown } from '../src/lib/markdown.mjs'

test('renders markdown and gives every heading a stable ASCII id', () => {
  const { html, headings } = renderMarkdown('## 先看`执行计划`\n\n正文\n\n### 覆盖 **索引**\n\n## [链接](https://example.test)标题')

  assert.deepEqual(headings, [
    { id: 'section-1', depth: 2, text: '先看执行计划' },
    { id: 'section-2', depth: 3, text: '覆盖 索引' },
    { id: 'section-3', depth: 2, text: '链接标题' },
  ])
  assert.match(html, /<h2 id="section-1">先看<code>执行计划<\/code><\/h2>/)
  assert.match(html, /<h3 id="section-2">覆盖 <strong>索引<\/strong><\/h3>/)
  assert.match(html, /<p>正文<\/p>/)
})

test('renders fenced code and GFM tables', () => {
  const { html } = renderMarkdown('```sql\nSELECT 1;\n```\n\n| a | b |\n| --- | --- |\n| 1 | 2 |')
  assert.match(html, /<pre><code class="language-sql">SELECT 1;/)
  assert.match(html, /<table>/)
})

test('treats missing content as an empty document', () => {
  assert.deepEqual(renderMarkdown(undefined), { html: '', headings: [] })
})
