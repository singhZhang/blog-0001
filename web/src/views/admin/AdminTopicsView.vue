<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Button from '../../components/ui/Button.vue'
import { useRouter } from 'vue-router'
import { invalidateArticles } from '../../composables/useArticles'
import { invalidateCategories } from '../../composables/useCategories'
import { ApiError } from '../../services/api'
import { createCategory, deleteCategory, updateCategory } from '../../services/admin'
import { getCategories } from '../../services/blog'
import type { Category } from '../../types/blog'

const router = useRouter()
const categories = ref<Category[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const editingName = ref('')
const form = reactive({ name: '', description: '' })

async function redirectIfUnauthorized(reason: unknown) {
  if (reason instanceof ApiError && (reason.status === 401 || reason.status === 503)) {
    await router.push({ name: 'admin-login' })
    return true
  }
  return false
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    categories.value = await getCategories()
  } catch (reason) {
    if (await redirectIfUnauthorized(reason)) return
    error.value = reason instanceof Error ? reason.message : '主题加载失败。'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  editingName.value = ''
  form.name = ''
  form.description = ''
}

function edit(category: Category) {
  editingName.value = category.name
  form.name = category.name
  form.description = category.description
}

async function save() {
  saving.value = true
  error.value = ''
  const input = { name: form.name.trim(), description: form.description.trim() }
  try {
    if (editingName.value) {
      await updateCategory(editingName.value, input)
      invalidateArticles()
    } else {
      await createCategory(input)
    }
    invalidateCategories()
    resetForm()
    await load()
  } catch (reason) {
    if (!(await redirectIfUnauthorized(reason))) {
      error.value = reason instanceof Error ? reason.message : '保存失败。'
    }
  } finally {
    saving.value = false
  }
}

async function remove(category: Category) {
  if (!window.confirm(`删除主题「${category.name}」？仍有文章使用时无法删除。`)) return
  error.value = ''
  try {
    await deleteCategory(category.name)
    invalidateCategories()
    if (editingName.value === category.name) resetForm()
    categories.value = categories.value.filter((item) => item.name !== category.name)
  } catch (reason) {
    if (!(await redirectIfUnauthorized(reason))) {
      error.value = reason instanceof Error ? reason.message : '删除失败。'
    }
  }
}

onMounted(load)
</script>

<template>
  <main class="admin-main">
    <section class="admin-sheet">
      <p class="admin-kicker">TOPICS / 主题</p>
      <h1>主题</h1>
      <p class="admin-lead">主题会出现在前台的主题页，也是写文章时的分类。改名会同步到已有文章。</p>

      <form class="admin-form" @submit.prevent="save">
        <div class="admin-grid">
          <label class="admin-field">
            <span>名称</span>
            <input v-model="form.name" required maxlength="191" placeholder="例如 Data">
          </label>
          <label class="admin-field">
            <span>介绍</span>
            <input v-model="form.description" required placeholder="这个主题在写什么">
          </label>
        </div>
        <div class="admin-actions">
          <Button type="submit" :disabled="saving">
            {{ saving ? '正在保存…' : (editingName ? '更新主题' : '新建主题') }}
          </Button>
          <Button v-if="editingName" variant="secondary" @click="resetForm">取消编辑</Button>
        </div>
      </form>

      <p v-if="error" class="admin-error" role="alert">{{ error }}</p>
      <p v-else-if="loading" class="admin-note">正在读取主题…</p>
      <p v-else-if="categories.length === 0" class="admin-note">还没有任何主题。</p>

      <table v-if="categories.length" class="admin-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>介绍</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="category in categories" :key="category.name">
            <td>{{ category.name }}</td>
            <td class="admin-muted">{{ category.description }}</td>
            <td>
              <div class="admin-row-actions">
                <Button variant="secondary" size="sm" @click="edit(category)">编辑</Button>
                <Button variant="danger" size="sm" @click="remove(category)">删除</Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>
