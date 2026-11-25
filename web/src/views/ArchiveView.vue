<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { RotateCw } from '@lucide/vue'
import Icon from '../components/ui/Icon.vue'
import PageMasthead from '../components/PageMasthead.vue'
import { useArticles } from '../composables/useArticles'
import { usePageMeta } from '../composables/usePageMeta'
import { formatChineseDate, groupArchive, sortByDateDesc, toTopicSlug } from '../lib/blog-content.mjs'

const { articles, loading, error, load, reload } = useArticles()
const years = computed(() => groupArchive(articles.value))
const topicCount = computed(() => new Set(articles.value.map((article) => article.category)).size)
const earliest = computed(() => sortByDateDesc(articles.value).at(-1))

onMounted(load)

usePageMeta(() => ({
  title: '归档',
  description: '按时间回看零玖肆玖肆的每一篇记录。',
  path: '/archive',
}))
</script>

<template>
  <PageMasthead
    index="ARCHIVE / 归档"
    title="归档"
    description="按时间回看每一篇记录。有些问题早已解决，有些想法还在生长。"
    :crumbs="[{ label: '首页', to: '/' }]"
  >
    <p v-if="articles.length" class="archive-stats">
      共 {{ articles.length }} 篇 · {{ topicCount }} 个主题
      <template v-if="earliest"> · 始于 {{ formatChineseDate(earliest.publishedAt) }}</template>
    </p>
  </PageMasthead>

  <div class="page-sheet">
    <div v-if="loading" class="page-state" aria-live="polite"><p>正在整理归档…</p></div>

    <div v-else-if="error" class="page-state" role="alert">
      <p>{{ error }}</p>
      <button type="button" @click="reload">
        <Icon :icon="RotateCw" :size="18" />
        重新加载
      </button>
    </div>

    <div v-else-if="years.length === 0" class="page-state"><p>还没有发布任何文章。</p></div>

    <template v-else>
      <section
        v-for="year in years"
        :key="year.year"
        class="archive-year"
        :aria-labelledby="`year-${year.year}`"
      >
        <header class="year-heading">
          <h2 :id="`year-${year.year}`">{{ year.year }}</h2>
          <span>{{ year.count }} 篇</span>
        </header>

        <div v-for="month in year.months" :key="month.month" class="archive-month">
          <h3>{{ Number(month.month) }} 月</h3>
          <ol class="timeline">
            <li v-for="article in month.articles" :key="article.id" class="timeline-item">
              <time :datetime="article.publishedAt">{{ article.publishedAt.slice(5) }}</time>
              <RouterLink class="timeline-title" :to="`/articles/${article.slug}`">{{ article.title }}</RouterLink>
              <RouterLink class="timeline-topic" :to="`/topics/${toTopicSlug(article.category)}`">
                {{ article.category }}
              </RouterLink>
            </li>
          </ol>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.archive-stats {
  margin: 22px 0 0;
  font-family: var(--font-mono);
  font-size: 13px;
  opacity: 0.85;
}

.archive-year + .archive-year { margin-top: 48px; }

.year-heading {
  display: flex;
  align-items: baseline;
  gap: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--line);
}

.year-heading h2 {
  margin: 0;
  font-family: var(--font-mono);
  font-size: clamp(34px, 3.4vw, 46px);
  letter-spacing: -0.04em;
}

.year-heading span {
  color: var(--muted);
  font-size: 13px;
}

.archive-month {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  gap: 24px;
  padding: 26px 0;
  border-bottom: 1px solid var(--line);
}

.archive-month h3 {
  margin: 0;
  padding-top: 4px;
  color: var(--muted);
  font-size: 15px;
  font-weight: 700;
}

.timeline {
  position: relative;
  display: grid;
  gap: 4px;
  margin: 0;
  padding: 0 0 0 26px;
  list-style: none;
}

.timeline::before {
  position: absolute;
  top: 10px;
  bottom: 10px;
  left: 4px;
  width: 1px;
  background: var(--line);
  content: '';
}

.timeline-item {
  position: relative;
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr) auto;
  align-items: baseline;
  gap: 18px;
  padding: 8px 0;
}

.timeline-item::before {
  position: absolute;
  top: 17px;
  left: -26px;
  width: 9px;
  height: 9px;
  border: 2px solid var(--paper);
  border-radius: 50%;
  background: #9fb0c6;
  content: '';
  transition: background 160ms ease;
}

.timeline-item:hover::before { background: var(--blue); }

.timeline-item time {
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 13px;
}

.timeline-title {
  color: var(--ink);
  font-size: 17px;
  font-weight: 700;
  line-height: 1.5;
  text-decoration: none;
}

.timeline-title:hover { color: var(--blue); }

.timeline-topic {
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 12px;
  text-decoration: none;
}

.timeline-topic:hover { color: var(--blue); }

@media (max-width: 760px) {
  .archive-month { grid-template-columns: 1fr; gap: 12px; }
  .timeline-item { grid-template-columns: 1fr auto; gap: 4px 12px; }
  .timeline-item time { grid-column: 1; }
  .timeline-topic { grid-column: 2; grid-row: 1; }
  .timeline-title { grid-column: 1 / -1; }
}
</style>
