<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ArrowRight, RotateCw } from '@lucide/vue'
import Icon from '../components/ui/Icon.vue'
import PageMasthead from '../components/PageMasthead.vue'
import { useArticles } from '../composables/useArticles'
import { useCategories } from '../composables/useCategories'
import { usePageMeta } from '../composables/usePageMeta'
import { buildTopics, formatChineseDate } from '../lib/blog-content.mjs'

const route = useRoute()
const { articles, loading: articlesLoading, error: articlesError, load, reload: reloadArticles } = useArticles()
const { categories, loading: categoriesLoading, error: categoriesError, load: loadCategories, reload: reloadCategories } = useCategories()
const loading = computed(() => articlesLoading.value || categoriesLoading.value)
const error = computed(() => articlesError.value || categoriesError.value)
const slug = computed(() => String(route.params.topic ?? ''))
const topics = computed(() => buildTopics(categories.value, articles.value))
const topic = computed(() => topics.value.find((item) => item.slug === slug.value) ?? null)
const otherTopics = computed(() => topics.value.filter((item) => item.slug !== slug.value))
const missing = computed(() => !loading.value && !error.value && !topic.value)
const description = computed(() => topic.value?.description)

const reload = () => {
  reloadArticles()
  reloadCategories()
}

onMounted(() => {
  load()
  loadCategories()
})

usePageMeta(() => ({
  title: topic.value ? `${topic.value.name} 主题` : '主题',
  description: description.value,
  path: `/topics/${slug.value}`,
  noindex: missing.value,
}))
</script>

<template>
  <PageMasthead
    :index="topic ? `TOPIC · ${topic.count} 篇` : 'TOPIC'"
    :title="topic?.name ?? (missing ? '没有这个主题' : '主题')"
    :description="description"
    :crumbs="[{ label: '首页', to: '/' }, { label: '主题', to: '/topics' }]"
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

    <div v-else-if="!topic" class="page-state">
      <p>这个主题还不存在，看看其他主题吧。</p>
      <div class="page-state-actions"><RouterLink to="/topics">全部主题</RouterLink></div>
    </div>

    <template v-else>
      <p v-if="topic.articles.length === 0" class="page-state">这个主题还没有文章。</p>
      <ol v-else class="topic-articles">
        <li v-for="article in topic.articles" :key="article.id">
          <RouterLink class="topic-article" :to="`/articles/${article.slug}`">
            <time :datetime="article.publishedAt">{{ formatChineseDate(article.publishedAt) }}</time>
            <div>
              <h2>{{ article.title }}</h2>
              <p>{{ article.excerpt }}</p>
              <span class="topic-article-tags">{{ article.tags.join(' / ') }}</span>
            </div>
            <span class="topic-article-read">{{ article.readMinutes }} 分钟</span>
            <Icon class="topic-article-arrow" :icon="ArrowRight" :size="22" />
          </RouterLink>
        </li>
      </ol>

      <nav v-if="otherTopics.length" class="other-topics" aria-label="其他主题">
        <span>其他主题</span>
        <RouterLink v-for="item in otherTopics" :key="item.slug" :to="`/topics/${item.slug}`">
          {{ item.name }} <small>{{ item.count }}</small>
        </RouterLink>
      </nav>
    </template>
  </div>
</template>

<style scoped>
.topic-articles {
  margin: 0;
  padding: 0;
  list-style: none;
}

.topic-article {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr) 72px 28px;
  align-items: center;
  gap: 24px;
  padding: 26px 12px;
  border-bottom: 1px solid var(--line);
  color: var(--ink);
  text-decoration: none;
  transition: background 180ms ease;
}

.topic-articles li:first-child .topic-article { padding-top: 10px; }
.topic-article:hover { background: var(--paper-blue); }

.topic-article time,
.topic-article-read {
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 12px;
}

.topic-article h2 {
  margin: 0 0 8px;
  font-size: 21px;
  line-height: 1.35;
  letter-spacing: -0.035em;
}

.topic-article p {
  margin: 0 0 10px;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.7;
}

.topic-article-tags {
  color: #6a7fa0;
  font-family: var(--font-mono);
  font-size: 11px;
}

.topic-article-read { text-align: right; }
.topic-article-arrow { transition: transform 180ms ease; }
.topic-article:hover .topic-article-arrow { transform: translateX(5px); color: var(--blue); }

.other-topics {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 34px;
}

.other-topics > span {
  margin-right: 8px;
  color: var(--muted);
  font-size: 13px;
}

.other-topics a {
  padding: 7px 14px;
  border: 1px solid var(--line);
  color: var(--ink);
  font-size: 14px;
  text-decoration: none;
}

.other-topics a:hover { border-color: var(--blue); color: var(--blue); }
.other-topics small { color: var(--muted); font-family: var(--font-mono); }

@media (max-width: 760px) {
  .topic-article {
    grid-template-columns: 1fr auto;
    gap: 10px 16px;
    padding: 22px 0;
  }
  .topic-article time { grid-column: 1; grid-row: 1; }
  .topic-article-read { grid-column: 2; grid-row: 1; }
  .topic-article > div { grid-column: 1 / -1; }
  .topic-article-arrow { display: none; }
}
</style>
