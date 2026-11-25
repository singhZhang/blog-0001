import { ref, type Ref } from 'vue'
import { getArticles } from '../services/blog'
import type { Article } from '../types/blog'

const articles = ref<Article[]>([])
const loading = ref(true)
const error = ref('')
let request: Promise<void> | null = null

function load(force = false) {
  if (request && !force) return request
  loading.value = true
  error.value = ''
  request = getArticles()
    .then((rows) => { articles.value = rows })
    .catch((reason: unknown) => {
      request = null
      error.value = reason instanceof Error ? reason.message : '暂时无法加载文章，请稍后重试。'
    })
    .finally(() => { loading.value = false })
  return request
}

export function invalidateArticles() {
  request = null
}

/** Article list shared by every page, fetched once per session. */
export function useArticles() {
  return {
    articles: articles as Readonly<Ref<Article[]>>,
    loading: loading as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string>>,
    load,
    reload: () => load(true),
  }
}
