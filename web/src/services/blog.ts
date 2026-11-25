import { buildArticleUrl, buildArticlesUrl, extractArticle, extractArticles } from '../lib/data-source.mjs'
import type { Article, Category } from '../types/blog'
import { apiBase, readError } from './api'

export async function getArticles(signal?: AbortSignal): Promise<Article[]> {
  const response = await fetch(buildArticlesUrl(apiBase()), {
    headers: { Accept: 'application/json' },
    signal,
  })
  if (!response.ok) throw await readError(response)
  return extractArticles(await response.json()) as Article[]
}

/** Resolves to `null` when the article does not exist. */
export async function getArticle(slug: string, signal?: AbortSignal): Promise<Article | null> {
  const response = await fetch(buildArticleUrl(apiBase(), slug), {
    headers: { Accept: 'application/json' },
    signal,
  })
  if (response.status === 404) return null
  if (!response.ok) throw await readError(response)
  return extractArticle(await response.json()) as Article
}

export async function getCategories(signal?: AbortSignal): Promise<Category[]> {
  const response = await fetch(`${apiBase()}/categories`, {
    headers: { Accept: 'application/json' },
    signal,
  })
  if (!response.ok) throw await readError(response)
  return extractArticles(await response.json()) as Category[]
}
