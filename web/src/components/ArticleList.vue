<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowRight, RotateCw } from '@lucide/vue'
import Icon from './ui/Icon.vue'
import { toTopicSlug } from '../lib/blog-content.mjs'
import type { Article } from '../types/blog'

const props = defineProps<{
  articles: Article[]
  loading: boolean
  error: string
}>()
defineEmits<{ retry: [] }>()

const activeCategory = ref('全部')
const categories = computed(() => ['全部', ...new Set(props.articles.map((article) => article.category))])
const visibleArticles = computed(() => activeCategory.value === '全部'
  ? props.articles
  : props.articles.filter((article) => article.category === activeCategory.value))

const formatDate = (value: string) => value.replaceAll('-', '—')
</script>

<template>
  <section id="articles" class="article-sheet" aria-labelledby="latest-heading">
    <div class="article-heading">
      <div>
        <p class="section-index">01 / WRITING</p>
        <h2 id="latest-heading">最新文章</h2>
      </div>
      <div class="topic-filter" aria-label="按主题筛选文章">
        <button
          v-for="category in categories"
          :key="category"
          type="button"
          :class="{ active: activeCategory === category }"
          @click="activeCategory = category"
        >
          {{ category }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="article-state" aria-live="polite">
      <span class="loading-line" />
      <span class="loading-line loading-line--short" />
      <p>正在整理最新记录…</p>
    </div>

    <div v-else-if="error" class="article-state article-state--error" role="alert">
      <p>{{ error }}</p>
      <button type="button" @click="$emit('retry')">
        <Icon :icon="RotateCw" :size="18" />
        重新加载
      </button>
    </div>

    <div v-else-if="visibleArticles.length === 0" class="article-state">
      <p>这个主题还没有文章，先看看其他记录吧。</p>
    </div>

    <div v-else class="article-list">
      <article v-for="article in visibleArticles" :id="article.id" :key="article.id" class="article-row">
        <time :datetime="article.publishedAt" class="article-date">{{ formatDate(article.publishedAt) }}</time>
        <p class="article-category">
          <RouterLink :to="`/topics/${toTopicSlug(article.category)}`">{{ article.category }}</RouterLink>
        </p>
        <div class="article-copy">
          <h3><RouterLink :to="`/articles/${article.slug}`">{{ article.title }}</RouterLink></h3>
          <p>{{ article.excerpt }}</p>
        </div>
        <pre v-if="article.featuredCode" class="code-preview" aria-label="文章代码示例"><code>{{ article.featuredCode }}</code></pre>
        <p v-else class="article-tags" aria-label="文章标签">{{ article.tags.join(' / ') }}</p>
        <p class="read-time">{{ article.readMinutes }} 分钟阅读</p>
        <RouterLink class="article-arrow" :to="`/articles/${article.slug}`" :aria-label="`阅读《${article.title}》`">
          <Icon :icon="ArrowRight" :size="24" />
        </RouterLink>
      </article>
    </div>

    <nav v-if="!loading && !error" class="article-more" aria-label="更多文章">
      <RouterLink to="/archive">
        浏览全部归档
        <Icon :icon="ArrowRight" :size="18" />
      </RouterLink>
      <RouterLink to="/topics">
        按主题浏览
        <Icon :icon="ArrowRight" :size="18" />
      </RouterLink>
    </nav>
  </section>
</template>

<style scoped>
.article-sheet {
  position: relative;
  z-index: 10;
  width: min(calc(100% - 48px), var(--content-wide));
  margin: -64px auto 0;
  padding: 24px 34px 12px;
  background: var(--paper);
}

.article-heading {
  min-height: 72px;
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 32px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--line);
}

.section-index {
  margin: 0 0 4px;
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
}

h2 {
  margin: 0;
  font-size: clamp(26px, 2.3vw, 34px);
  letter-spacing: -0.05em;
}

.topic-filter {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 6px 18px;
}

.topic-filter button {
  padding: 5px 0;
  border: 0;
  border-bottom: 1px solid transparent;
  background: transparent;
  color: var(--muted);
  font: inherit;
  font-size: 13px;
  cursor: pointer;
}

.topic-filter button:hover,
.topic-filter button.active {
  border-color: var(--blue);
  color: var(--blue);
}

.article-row {
  min-height: 112px;
  display: grid;
  grid-template-columns: 120px 90px minmax(260px, 1fr) minmax(180px, 230px) 106px 32px;
  align-items: center;
  gap: 20px;
  border-bottom: 1px solid var(--line);
  transition: background 180ms ease;
}

.article-row:hover {
  background: var(--paper-blue);
}

.article-date,
.article-category,
.read-time,
.article-tags {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
}

.article-date,
.article-category {
  font-family: var(--font-mono);
}

.article-category a {
  color: inherit;
  text-decoration: none;
}

.article-category a:hover { color: var(--blue); }

.article-more {
  display: flex;
  justify-content: flex-end;
  gap: 28px;
  padding: 22px 0 14px;
}

.article-more a {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
}

.article-more a:hover { color: var(--blue); }
.article-more :deep(svg) { transition: transform 180ms ease; }
.article-more a:hover :deep(svg) { transform: translateX(4px); }

.article-copy h3 {
  margin: 0 0 6px;
  font-size: 19px;
  line-height: 1.25;
  letter-spacing: -0.035em;
}

.article-copy h3 a {
  color: inherit;
  text-decoration: none;
}

.article-copy > p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.code-preview {
  margin: 0;
  padding: 13px 15px;
  overflow: hidden;
  background: var(--blue);
  color: #c6e7ff;
  font-family: var(--font-mono);
  font-size: 10px;
  line-height: 1.55;
  white-space: pre-wrap;
}

.article-tags {
  font-family: var(--font-mono);
  line-height: 1.55;
}

.read-time { text-align: right; }

.article-arrow {
  width: 32px;
  height: 44px;
  display: grid;
  place-items: center;
  color: var(--ink);
  text-decoration: none;
}

.article-arrow :deep(svg) { transition: transform 180ms ease; }
.article-row:hover .article-arrow :deep(svg) { transform: translateX(5px); }

.article-state {
  min-height: 260px;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 12px;
  color: var(--muted);
  text-align: center;
}

.article-state button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 1px solid var(--ink);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
}

.loading-line {
  width: 220px;
  height: 8px;
  background: var(--line);
  animation: pulse 1.2s ease-in-out infinite alternate;
}

.loading-line--short { width: 150px; }

@keyframes pulse { to { opacity: 0.35; } }

@media (max-width: 1180px) {
  .article-row {
    grid-template-columns: 104px 76px minmax(250px, 1fr) 90px 32px;
  }
  .code-preview,
  .article-tags { display: none; }
}

@media (max-width: 760px) {
  .article-sheet {
    width: calc(100% - 24px);
    margin-top: -52px;
    padding: 20px 20px 6px;
  }
  .article-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 16px;
  }
  .topic-filter { justify-content: flex-start; }
  .article-row {
    min-height: 0;
    grid-template-columns: 1fr auto;
    gap: 8px 16px;
    padding: 22px 0;
  }
  .article-date { grid-column: 1; grid-row: 1; }
  .article-category { grid-column: 2; grid-row: 1; text-align: right; }
  .article-copy { grid-column: 1 / -1; grid-row: 2; }
  .read-time { grid-column: 1; grid-row: 3; text-align: left; }
  .article-arrow { grid-column: 2; grid-row: 3; }
  .article-copy h3 { font-size: 20px; }
  .article-more { justify-content: flex-start; flex-wrap: wrap; gap: 14px 24px; }
}

@media (prefers-reduced-motion: reduce) {
  .loading-line { animation: none; }
}
</style>
