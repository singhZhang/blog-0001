<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from '../../components/ui/Button.vue'
import { RouterLink, useRouter } from 'vue-router'
import { invalidateArticles } from '../../composables/useArticles'
import { ApiError } from '../../services/api'
import { deleteArticle } from '../../services/admin'
import { getArticles } from '../../services/blog'
import type { Article } from '../../types/blog'

const router = useRouter()
const articles = ref<Article[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    articles.value = await getArticles()
  } catch (reason) {
    if (reason instanceof ApiError && reason.status === 401) {
      await router.push({ name: 'admin-login' })
      return
    }
    error.value = reason instanceof Error ? reason.message : '文章加载失败。'
  } finally {
    loading.value = false
  }
}

async function remove(article: Article) {
  if (!window.confirm(`删除《${article.title}》？此操作不可恢复。`)) return
  error.value = ''
  try {
    await deleteArticle(article.slug)
    invalidateArticles()
    articles.value = articles.value.filter((item) => item.slug !== article.slug)
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : '删除失败。'
  }
}

onMounted(load)
</script>

<template>
  <main class="admin-main">
    <section class="admin-sheet">
      <div class="admin-toolbar">
        <div>
          <p class="admin-kicker">ARTICLES / 文章</p>
          <h1>文章</h1>
          <p class="admin-lead">新建、修改或下线一篇记录。阅读时长由正文长度自动估算。</p>
        </div>
        <RouterLink class="inline-flex min-h-10 items-center border border-ink bg-ink px-4 text-sm text-white no-underline hover:border-blue hover:bg-blue" :to="{ name: 'admin-article-new' }">写一篇</RouterLink>
      </div>

      <p v-if="error" class="admin-error" role="alert">{{ error }}</p>
      <p v-else-if="loading" class="admin-note">正在读取文章…</p>
      <p v-else-if="articles.length === 0" class="admin-note">还没有文章。</p>

      <table v-else class="admin-table">
        <thead>
          <tr>
            <th>日期</th>
            <th>标题</th>
            <th>主题</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="article in articles" :key="article.id">
            <td class="admin-muted">{{ article.publishedAt }}</td>
            <td>
              <RouterLink :to="{ name: 'admin-article-edit', params: { slug: article.slug } }">{{ article.title }}</RouterLink>
            </td>
            <td class="admin-muted">{{ article.category }}</td>
            <td>
              <div class="admin-row-actions">
                <RouterLink :to="{ name: 'admin-article-edit', params: { slug: article.slug } }">编辑</RouterLink>
                <Button variant="danger" size="sm" @click="remove(article)">删除</Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>
