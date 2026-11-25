import { computed, ref } from 'vue'

const STORAGE_KEY = 'blog-admin-token'

function readToken() {
  try {
    return sessionStorage.getItem(STORAGE_KEY) || ''
  } catch {
    return ''
  }
}

const token = ref(readToken())

export function adminToken() {
  return token.value
}

export function useAdminSession() {
  return {
    token,
    authed: computed(() => token.value !== ''),
    setToken(value: string) {
      token.value = value.trim()
      sessionStorage.setItem(STORAGE_KEY, token.value)
    },
    clear() {
      token.value = ''
      try {
        sessionStorage.removeItem(STORAGE_KEY)
      } catch {
        // 隐私模式可能拒绝访问存储，内存中的令牌已经清空。
      }
    },
  }
}
