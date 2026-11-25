/** @param {string} base */
export function buildArticlesUrl(base) {
  return `${String(base || '/api').replace(/\/+$/, '')}/articles`
}

/**
 * @param {string} base
 * @param {string} slug
 */
export function buildArticleUrl(base, slug) {
  return `${buildArticlesUrl(base)}/${encodeURIComponent(slug)}`
}

/** @param {unknown} payload */
export function extractArticles(payload) {
  if (Array.isArray(payload)) return payload
  if (payload && typeof payload === 'object' && Array.isArray(payload.data)) return payload.data
  throw new Error('文章数据格式无效')
}

/** @param {unknown} payload */
export function extractArticle(payload) {
  const article = payload && typeof payload === 'object' && 'data' in payload ? payload.data : payload
  if (article && typeof article === 'object' && !Array.isArray(article) && typeof article.slug === 'string') {
    return article
  }
  throw new Error('文章数据格式无效')
}
