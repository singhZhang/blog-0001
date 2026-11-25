<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import DOMPurify from 'dompurify'
import Button from '../../components/ui/Button.vue'
import { invalidateArticles } from '../../composables/useArticles'
import { invalidateCategories } from '../../composables/useCategories'
import { renderMarkdown } from '../../lib/markdown.mjs'
import { ApiError } from '../../services/api'
import { createArticle, deleteArticle, updateArticle } from '../../services/admin'
import { getArticle, getCategories } from '../../services/blog'
import type { Category } from '../../types/blog'

const route = useRoute()
const router = useRouter()
const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const originalSlug = computed(() => route.name === 'admin-article-edit' ? String(route.params.slug ?? '') : '')
const editing = computed(() => originalSlug.value !== '')

function today() {
  const date = new Date()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

const form = reactive({
  title: '',
  slug: '',
  publishedAt: today(),
  category: '',
  tags: [] as string[],
  content: '',
})
const tagDraft = ref('')
const contentBox = ref<HTMLTextAreaElement | null>(null)
const previewing = ref(false)
const previewHtml = computed(() => DOMPurify.sanitize(renderMarkdown(form.content).html))

const markdownGroups: { label: string, title: string, wrap: (text: string) => string, sample: string }[][] = [
  [
    { label: 'H1', title: '一级标题', sample: '标题', wrap: (text) => `# ${text}` },
    { label: 'H2', title: '二级标题', sample: '标题', wrap: (text) => `## ${text}` },
    { label: 'H3', title: '三级标题', sample: '标题', wrap: (text) => `### ${text}` },
  ],
  [
    { label: 'B', title: '粗体', sample: '粗体', wrap: (text) => `**${text}**` },
    { label: 'I', title: '斜体', sample: '斜体', wrap: (text) => `*${text}*` },
    { label: 'S', title: '删除线', sample: '删除线', wrap: (text) => `~~${text}~~` },
  ],
  [
    { label: '链接', title: '链接', sample: '链接文字', wrap: (text) => `[${text}](https://)` },
    { label: '图片', title: '图片', sample: '图片说明', wrap: (text) => `![${text}](https://)` },
  ],
  [
    { label: '代码块', title: '代码块', sample: 'code', wrap: (text) => `\`\`\`\n${text}\n\`\`\`` },
  ],
  [
    { label: '引用', title: '引用', sample: '引用', wrap: (text) => text.split('\n').map((line) => `> ${line}`).join('\n') },
    { label: '列表', title: '无序列表', sample: '列表项', wrap: (text) => text.split('\n').map((line) => `- ${line}`).join('\n') },
    { label: '编号', title: '有序列表', sample: '列表项', wrap: (text) => text.split('\n').map((line, index) => `${index + 1}. ${line}`).join('\n') },
  ],
  [
    {
      label: '表格',
      title: '表格',
      sample: '',
      wrap: () => '| 列一 | 列二 |\n| --- | --- |\n|  |  |',
    },
    { label: '分隔线', title: '分隔线', sample: '', wrap: () => '---' },
  ],
]

function insertMarkdown(tool: (typeof markdownGroups)[number][number]) {
  const box = contentBox.value
  const start = box?.selectionStart ?? form.content.length
  const end = box?.selectionEnd ?? start
  const selected = form.content.slice(start, end)
  const inserted = `${tool.wrap(selected || tool.sample)}\n`
  form.content = form.content.slice(0, start) + inserted + form.content.slice(end)
  const cursor = start + inserted.length
  requestAnimationFrame(() => {
    box?.focus()
    box?.setSelectionRange(cursor, cursor)
  })
}

function reset() {
  form.title = ''
  form.slug = ''
  form.publishedAt = today()
  form.category = categories.value[0]?.name ?? ''
  form.tags = []
  tagDraft.value = ''
  form.content = ''
}

function addTag() {
  const parts = tagDraft.value.split(/[,，]/).map((tag) => tag.trim()).filter(Boolean)
  for (const tag of parts) {
    if (!form.tags.includes(tag)) form.tags.push(tag)
  }
  tagDraft.value = ''
}

function removeTag(tag: string) {
  form.tags = form.tags.filter((item) => item !== tag)
}

function onTagKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    addTag()
    return
  }
  if (event.key === 'Backspace' && tagDraft.value === '' && form.tags.length > 0) {
    form.tags.pop()
  }
}

async function loadCategories() {
  categories.value = await getCategories()
  if (!form.category || !categories.value.some((category) => category.name === form.category)) {
    form.category = categories.value[0]?.name ?? ''
  }
}

async function loadArticle(slug: string) {
  loading.value = true
  error.value = ''
  try {
    await loadCategories()
    const article = await getArticle(slug)
    if (!article) {
      error.value = '文章不存在或已下线。'
      return
    }
    form.title = article.title
    form.slug = article.slug
    form.publishedAt = article.publishedAt
    form.category = article.category
    form.tags = [...(article.tags ?? [])]
    form.content = article.content ?? ''
  } catch (reason) {
    await redirectIfUnauthorized(reason)
    error.value = reason instanceof Error ? reason.message : '文章加载失败。'
  } finally {
    loading.value = false
  }
}

async function redirectIfUnauthorized(reason: unknown) {
  if (reason instanceof ApiError && (reason.status === 401 || reason.status === 503)) {
    await router.push({ name: 'admin-login' })
  }
}

async function save() {
  if (!form.category) {
    error.value = '请先创建一个主题，再写文章。'
    return
  }
  saving.value = true
  error.value = ''
  addTag()
  const input = {
    title: form.title.trim(),
    slug: form.slug.trim().toLowerCase(),
    publishedAt: editing.value ? form.publishedAt : today(),
    category: form.category,
    tags: [...form.tags],
    content: form.content.trim(),
  }
  try {
    const saved = editing.value
      ? await updateArticle(originalSlug.value, input)
      : await createArticle(input)
    invalidateArticles()
    if (saved.slug !== originalSlug.value) {
      await router.replace({ name: 'admin-article-edit', params: { slug: saved.slug } })
    }
  } catch (reason) {
    await redirectIfUnauthorized(reason)
    error.value = reason instanceof Error ? reason.message : '保存失败。'
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!editing.value || !window.confirm('删除这篇文章？此操作不可恢复。')) return
  saving.value = true
  error.value = ''
  try {
    await deleteArticle(originalSlug.value)
    invalidateArticles()
    invalidateCategories()
    await router.push({ name: 'admin-articles' })
  } catch (reason) {
    await redirectIfUnauthorized(reason)
    error.value = reason instanceof Error ? reason.message : '删除失败。'
  } finally {
    saving.value = false
  }
}

watch(originalSlug, (slug) => {
  if (slug) {
    loadArticle(slug)
    return
  }
  error.value = ''
  loadCategories().then(reset).catch((reason: unknown) => {
    error.value = reason instanceof Error ? reason.message : '主题加载失败。'
  })
}, { immediate: true })
</script>

<template>
  <main class="admin-main">
    <section class="admin-sheet">
      <div class="admin-toolbar">
        <div>
          <p class="admin-kicker">{{ editing ? 'EDIT / 编辑' : 'NEW / 新文章' }}</p>
          <h1>{{ editing ? '编辑文章' : '写一篇文章' }}</h1>
        </div>
        <RouterLink class="admin-button secondary" :to="{ name: 'admin-articles' }">返回列表</RouterLink>
      </div>

      <p v-if="loading" class="admin-note">正在打开文章…</p>
      <form v-else class="admin-form" @submit.prevent="save">
        <label class="admin-field">
          <span>标题</span>
          <input v-model="form.title" required maxlength="500">
        </label>
        <div class="admin-grid">
          <label class="admin-field">
            <span>SLUG</span>
            <input v-model="form.slug" required autocapitalize="off" spellcheck="false" placeholder="小写字母、数字和连字符">
          </label>
          <label class="admin-field">
            <span>主题</span>
            <select v-model="form.category" required>
              <option v-if="categories.length === 0" value="" disabled>还没有主题</option>
              <option
                v-if="form.category && !categories.some((category) => category.name === form.category)"
                :value="form.category"
              >{{ form.category }}</option>
              <option v-for="category in categories" :key="category.name" :value="category.name">{{ category.name }}</option>
            </select>
          </label>
        </div>
        <label v-if="editing" class="admin-field">
          <span>发布日期</span>
          <input v-model="form.publishedAt" type="date" required>
        </label>

        <div class="admin-field">
          <span>标签</span>
          <div class="tag-editor">
            <span v-for="tag in form.tags" :key="tag" class="tag-chip">
              {{ tag }}
              <button type="button" :aria-label="`移除标签 ${tag}`" @click="removeTag(tag)">×</button>
            </span>
            <input
              v-model="tagDraft"
              placeholder="输入后按回车"
              @keydown="onTagKeydown"
            >
          </div>
        </div>
        <div class="admin-field">
          <span>正文</span>
          <div class="markdown-editor">
            <div class="markdown-toolbar" role="toolbar" aria-label="Markdown 格式">
              <div class="markdown-tools">
                <template v-for="(group, index) in markdownGroups" :key="index">
                  <span v-if="index > 0" class="markdown-split" aria-hidden="true" />
                  <button
                    v-for="tool in group"
                    :key="tool.label"
                    type="button"
                    :title="tool.title"
                    :disabled="previewing"
                    @click="insertMarkdown(tool)"
                  >
                    {{ tool.label }}
                  </button>
                </template>
              </div>
              <div class="markdown-mode" role="tablist" aria-label="正文视图">
                <button type="button" :class="{ active: !previewing }" @click="previewing = false">编辑</button>
                <button type="button" :class="{ active: previewing }" @click="previewing = true">预览</button>
              </div>
            </div>
            <textarea v-show="!previewing" ref="contentBox" v-model="form.content" class="content" required />
            <div v-show="previewing" class="markdown-preview prose" aria-live="polite">
              <p v-if="!form.content.trim()" class="markdown-empty">还没有正文。</p>
              <div v-else v-html="previewHtml" />
            </div>
          </div>
        </div>

        <p v-if="categories.length === 0" class="admin-note">
          还没有主题。先到<RouterLink to="/admin/topics">主题</RouterLink>里创建一个，再回来写文章。
        </p>
        <p v-if="error" class="admin-error" role="alert">{{ error }}</p>
        <div class="admin-actions">
          <Button type="submit" :disabled="saving || categories.length === 0">
            {{ saving ? '正在保存…' : '保存' }}
          </Button>
          <Button v-if="editing" variant="danger" :disabled="saving" @click="remove">删除</Button>
        </div>
      </form>
    </section>
  </main>
</template>
