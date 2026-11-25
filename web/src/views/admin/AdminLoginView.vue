<script setup lang="ts">
import { ref } from 'vue'
import Button from '../../components/ui/Button.vue'
import Input from '../../components/ui/Input.vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminSession } from '../../composables/useAdminSession'
import { ApiError } from '../../services/api'
import { verifyAdminSession } from '../../services/admin'

const route = useRoute()
const router = useRouter()
const { setToken, clear } = useAdminSession()
const token = ref('')
const error = ref('')
const submitting = ref(false)

function nextPath() {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/admin/articles'
  if (!redirect.startsWith('/admin') || redirect.startsWith('/admin/login')) return '/admin/articles'
  return redirect
}

async function submit() {
  error.value = ''
  submitting.value = true
  setToken(token.value)
  try {
    await verifyAdminSession()
    await router.push(nextPath())
  } catch (reason) {
    clear()
    error.value = reason instanceof ApiError ? reason.message : '无法连接管理接口。'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="admin-main">
    <form class="admin-sheet admin-login" @submit.prevent="submit">
      <p class="admin-kicker">ADMIN / 登录</p>
      <h1>后台</h1>
      <p class="admin-lead">用管理令牌进入，之后可以写文章、改主题。</p>
      <label class="admin-field">
        <span>管理令牌</span>
        <Input v-model="token" type="password" autocomplete="current-password" required />
      </label>
      <p class="admin-note">本地默认令牌是 blog-admin，对应服务端的 ADMIN_TOKEN。</p>
      <p v-if="error" class="admin-error" role="alert">{{ error }}</p>
      <div class="admin-actions">
        <Button type="submit" :disabled="submitting">
          {{ submitting ? '正在校验…' : '进入后台' }}
        </Button>
      </div>
    </form>
  </main>
</template>
