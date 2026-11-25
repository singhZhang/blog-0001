/** @typedef {import('../types/blog').Article} Article */

/** @param {string} name */
export function toTopicSlug(name) {
  return String(name)
    .trim()
    .toLocaleLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, '-')
    .replace(/^-+|-+$/g, '')
}

/** @param {string} isoDate `YYYY-MM-DD` */
export function formatChineseDate(isoDate) {
  const [year, month, day] = isoDate.split('-').map(Number)
  return `${year}年${month}月${day}日`
}

/** @param {Article[]} articles */
export function sortByDateDesc(articles) {
  return [...articles].sort((a, b) => b.publishedAt.localeCompare(a.publishedAt))
}

/**
 * @param {Article[]} articles
 * @returns {{ year: string, count: number, months: { month: string, articles: Article[] }[] }[]}
 */
export function groupArchive(articles) {
  /** @type {Map<string, Map<string, Article[]>>} */
  const years = new Map()
  for (const article of sortByDateDesc(articles)) {
    const [year, month] = article.publishedAt.split('-')
    if (!years.has(year)) years.set(year, new Map())
    const months = /** @type {Map<string, Article[]>} */ (years.get(year))
    if (!months.has(month)) months.set(month, [])
    months.get(month)?.push(article)
  }

  return [...years].map(([year, months]) => ({
    year,
    count: [...months.values()].reduce((total, rows) => total + rows.length, 0),
    months: [...months].map(([month, rows]) => ({ month, articles: rows })),
  }))
}

/**
 * @param {Article[]} articles
 * @returns {{ name: string, slug: string, count: number, latest: Article, articles: Article[] }[]}
 */
export function groupTopics(articles) {
  /** @type {Map<string, Article[]>} */
  const topics = new Map()
  for (const article of sortByDateDesc(articles)) {
    if (!topics.has(article.category)) topics.set(article.category, [])
    topics.get(article.category)?.push(article)
  }

  return [...topics]
    .map(([name, rows]) => ({ name, slug: toTopicSlug(name), count: rows.length, latest: rows[0], articles: rows }))
    .sort((a, b) => b.count - a.count || b.latest.publishedAt.localeCompare(a.latest.publishedAt))
}

/**
 * Categories are the source of truth for themes. Articles fill in counts.
 * A theme with no articles stays visible; an article category missing from the
 * category list is still shown so nothing disappears.
 * @param {{ name: string, description: string }[]} categories
 * @param {Article[]} articles
 */
export function buildTopics(categories, articles) {
  const grouped = new Map(groupTopics(articles).map((topic) => [topic.name, topic]))
  const known = new Set()
  const rows = []

  for (const category of categories) {
    known.add(category.name)
    const existing = grouped.get(category.name)
    rows.push(existing
      ? { ...existing, description: category.description }
      : {
          name: category.name,
          slug: toTopicSlug(category.name),
          count: 0,
          latest: null,
          articles: [],
          description: category.description,
        })
  }

  for (const topic of grouped.values()) {
    if (!known.has(topic.name)) {
      rows.push({ ...topic, description: `关于 ${topic.name} 的记录与思考。` })
    }
  }

  return rows.sort((a, b) => b.count - a.count || a.name.localeCompare(b.name, 'zh'))
}

/**
 * @param {Article[]} articles
 * @param {string} slug
 * @returns {{ newer: Article | null, older: Article | null }}
 */
export function getAdjacentArticles(articles, slug) {
  const sorted = sortByDateDesc(articles)
  const index = sorted.findIndex((article) => article.slug === slug)
  if (index === -1) return { newer: null, older: null }
  return { newer: sorted[index - 1] ?? null, older: sorted[index + 1] ?? null }
}

/**
 * Same category weighs more than a single shared tag; unrelated articles are never suggested.
 * @param {Article[]} articles
 * @param {Article} current
 * @param {number} [limit]
 */
export function getRelatedArticles(articles, current, limit = 3) {
  const tags = new Set(current.tags)
  return sortByDateDesc(articles)
    .filter((article) => article.slug !== current.slug)
    .map((article) => ({
      article,
      score: (article.category === current.category ? 2 : 0)
        + article.tags.filter((tag) => tags.has(tag)).length,
    }))
    .filter(({ score }) => score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
    .map(({ article }) => article)
}
