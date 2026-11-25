import { adminToken } from '../composables/useAdminSession'
import type { Article, Category } from '../types/blog'
import { apiBase, readError } from './api'

export interface ArticleInput {
  slug: string
  title: string
  publishedAt: string
  category: string
  tags: string[]
  featuredCode?: string
  content: string
}

async function adminFetch(path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (init.body != null) headers.set('Content-Type', 'application/json')
  const token = adminToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const response = await fetch(`${apiBase()}${path}`, { ...init, headers })
  if (response.status === 204) return null
  if (!response.ok) throw await readError(response)
  if (response.status === 200 || response.status === 201) return response.json()
  return null
}

export function verifyAdminSession() {
  return adminFetch('/admin/session')
}

export function createArticle(input: ArticleInput) {
  return adminFetch('/admin/articles', { method: 'POST', body: JSON.stringify(input) }) as Promise<Article>
}

export function updateArticle(slug: string, input: ArticleInput) {
  return adminFetch(`/admin/articles/${encodeURIComponent(slug)}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }) as Promise<Article>
}

export function deleteArticle(slug: string) {
  return adminFetch(`/admin/articles/${encodeURIComponent(slug)}`, { method: 'DELETE' })
}

export function createCategory(input: Category) {
  return adminFetch('/admin/categories', { method: 'POST', body: JSON.stringify(input) }) as Promise<Category>
}

export function updateCategory(name: string, input: Category) {
  return adminFetch(`/admin/categories/${encodeURIComponent(name)}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }) as Promise<Category>
}

export function deleteCategory(name: string) {
  return adminFetch(`/admin/categories/${encodeURIComponent(name)}`, { method: 'DELETE' })
}
