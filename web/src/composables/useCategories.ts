import { ref, type Ref } from 'vue'
import { getCategories } from '../services/blog'
import type { Category } from '../types/blog'

const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref('')
let request: Promise<void> | null = null

function load(force = false) {
  if (request && !force) return request
  loading.value = true
  error.value = ''
  request = getCategories()
    .then((rows) => { categories.value = rows })
    .catch((reason: unknown) => {
      request = null
      error.value = reason instanceof Error ? reason.message : '暂时无法加载主题，请稍后重试。'
    })
    .finally(() => { loading.value = false })
  return request
}

export function invalidateCategories() {
  request = null
}

export function useCategories() {
  return {
    categories: categories as Readonly<Ref<Category[]>>,
    loading: loading as Readonly<Ref<boolean>>,
    error: error as Readonly<Ref<string>>,
    load,
    reload: () => load(true),
  }
}
