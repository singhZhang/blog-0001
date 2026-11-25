<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useAdminSession } from '../../composables/useAdminSession'
import { usePageMeta } from '../../composables/usePageMeta'
import '../../styles/admin.css'

const route = useRoute()
const router = useRouter()
const { authed, clear } = useAdminSession()

usePageMeta(() => ({
  title: '后台',
  description: '管理文章与主题。',
  path: route.path,
  noindex: true,
}))

function logout() {
  clear()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <div class="admin-app">
    <header class="admin-top">
      <RouterLink class="admin-brand" to="/">零玖肆玖肆</RouterLink>
      <nav v-if="authed" class="admin-nav" aria-label="后台导航">
        <RouterLink :to="{ name: 'admin-articles' }">文章</RouterLink>
        <RouterLink :to="{ name: 'admin-topics' }">主题</RouterLink>
      </nav>
      <div class="admin-top-actions">
        <RouterLink to="/">返回站点</RouterLink>
        <button v-if="authed" type="button" @click="logout">退出</button>
      </div>
    </header>
    <RouterView />
  </div>
</template>
