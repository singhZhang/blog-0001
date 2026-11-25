<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowUpRight, RotateCw } from '@lucide/vue'
import Icon from '../components/ui/Icon.vue'
import PageMasthead from '../components/PageMasthead.vue'
import { useArticles } from '../composables/useArticles'
import { useCategories } from '../composables/useCategories'
import { usePageMeta } from '../composables/usePageMeta'
import { buildTopics } from '../lib/blog-content.mjs'

const { articles, loading: articlesLoading, error: articlesError, load, reload: reloadArticles } = useArticles()
const { categories, loading: categoriesLoading, error: categoriesError, load: loadCategories, reload: reloadCategories } = useCategories()
const loading = computed(() => articlesLoading.value || categoriesLoading.value)
const error = computed(() => articlesError.value || categoriesError.value)
const topics = computed(() => buildTopics(categories.value, articles.value))
const reload = () => {
  reloadArticles()
  reloadCategories()
}

onMounted(() => {
  load()
  loadCategories()
})

usePageMeta(() => ({
  title: '主题',
  description: '按主题浏览零玖肆玖肆的文章：数据、前端、AI 系统与架构。',
  path: '/topics',
}))
</script>

<template>
  <PageMasthead
    index="TOPICS / 主题"
    title="主题"
    description="同一类问题，常常值得反复思考。按主题把散落的记录重新串起来。"
    :crumbs="[{ label: '首页', to: '/' }]"
  />

  <div class="page-sheet">
    <div v-if="loading" class="page-state" aria-live="polite"><p>正在整理主题…</p></div>

    <div v-else-if="error" class="page-state" role="alert">
      <p>{{ error }}</p>
      <button type="button" @click="reload">
        <Icon :icon="RotateCw" :size="18" />
        重新加载
      </button>
    </div>

    <div v-else-if="topics.length === 0" class="page-state"><p>还没有任何主题。</p></div>

    <ul v-else class="topic-grid">
      <li v-for="(topic, position) in topics" :key="topic.slug">
        <RouterLink class="topic-card" :to="`/topics/${topic.slug}`">
          <span class="topic-index">{{ String(position + 1).padStart(2, '0') }}</span>
          <Icon class="topic-arrow" :icon="ArrowUpRight" :size="22" />
          <h2>{{ topic.name }}</h2>
          <p class="topic-description">{{ topic.description }}</p>
          <p class="topic-meta">
            <strong>{{ topic.count }} 篇</strong>
            <span v-if="topic.latest">最近：{{ topic.latest.title }}</span>
            <span v-else>还没有文章</span>
          </p>
        </RouterLink>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.topic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 18px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.topic-card {
  position: relative;
  height: 100%;
  min-height: 236px;
  display: grid;
  grid-template-rows: auto auto 1fr auto;
  gap: 10px;
  padding: 26px 26px 22px;
  border: 1px solid var(--line);
  color: var(--ink);
  text-decoration: none;
  transition: border-color 180ms ease, background 180ms ease, transform 180ms ease;
}

.topic-card:hover {
  border-color: var(--blue);
  background: var(--paper-blue);
  transform: translateY(-3px);
}

.topic-index {
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.1em;
}

.topic-arrow {
  position: absolute;
  top: 22px;
  right: 22px;
  color: var(--muted);
  transition: color 180ms ease, transform 180ms ease;
}

.topic-card:hover .topic-arrow {
  color: var(--blue);
  transform: translate(3px, -3px);
}

h2 {
  margin: 8px 0 0;
  font-size: 28px;
  letter-spacing: -0.05em;
}

.topic-description {
  margin: 0;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.7;
}

.topic-meta {
  display: grid;
  gap: 4px;
  margin: 12px 0 0;
  padding-top: 14px;
  border-top: 1px solid var(--line);
  font-size: 13px;
}

.topic-meta strong { color: var(--blue); font-family: var(--font-mono); }

.topic-meta span {
  overflow: hidden;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 600px) {
  .topic-grid { grid-template-columns: 1fr; }
  .topic-card { min-height: 0; }
}
</style>
