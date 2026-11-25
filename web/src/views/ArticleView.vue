<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import DOMPurify from 'dompurify'
import { ArrowLeft, ArrowRight, RotateCw } from '@lucide/vue'
import Icon from '../components/ui/Icon.vue'
import PageMasthead from '../components/PageMasthead.vue'
import NewsletterBand from '../components/NewsletterBand.vue'
import { useArticles } from '../composables/useArticles'
import { siteBaseUrl, usePageMeta } from '../composables/usePageMeta'
import { formatChineseDate, getAdjacentArticles, getRelatedArticles, toTopicSlug } from '../lib/blog-content.mjs'
import { renderMarkdown } from '../lib/markdown.mjs'
import { getArticle } from '../services/blog'
import { siteProfile } from '../data/site'
import type { Article } from '../types/blog'

const route = useRoute()
const { articles, load } = useArticles()
const slug = computed(() => String(route.params.slug ?? ''))
const article = ref<Article | null>(null)
const status = ref<'loading' | 'ready' | 'missing' | 'error'>('loading')
const errorMessage = ref('')
const body = ref<HTMLElement | null>(null)
const progress = ref(0)
const activeHeading = ref('')
let controller: AbortController | null = null
let frame = 0

const rendered = computed(() => renderMarkdown(article.value?.content ?? ''))
const safeHtml = computed(() => DOMPurify.sanitize(rendered.value.html))
const toc = computed(() => rendered.value.headings.filter((heading) => heading.depth <= 3))
const adjacent = computed(() => getAdjacentArticles(articles.value, slug.value))
const related = computed(() => article.value ? getRelatedArticles(articles.value, article.value) : [])
const topicLink = computed(() => article.value ? `/topics/${toTopicSlug(article.value.category)}` : '/topics')

const mastheadTitle = computed(() => {
  if (article.value) return article.value.title
  if (status.value === 'missing') return '没有找到这篇文章'
  if (status.value === 'error') return '文章暂时无法打开'
  return '正在打开文章…'
})

async function fetchArticle() {
  controller?.abort()
  controller = new AbortController()
  status.value = 'loading'
  errorMessage.value = ''
  article.value = null
  try {
    const result = await getArticle(slug.value, controller.signal)
    article.value = result
    status.value = result ? 'ready' : 'missing'
  } catch (reason) {
    if (reason instanceof DOMException && reason.name === 'AbortError') return
    errorMessage.value = reason instanceof Error ? reason.message : '暂时无法加载文章，请稍后重试。'
    status.value = 'error'
  }
}

function updateReadingState() {
  frame = 0
  const element = body.value
  if (!element) return
  const rect = element.getBoundingClientRect()
  const readingLine = window.innerHeight / 2
  progress.value = Math.min(1, Math.max(0, (readingLine - rect.top) / rect.height))

  let current = ''
  for (const heading of element.querySelectorAll<HTMLElement>('h2[id], h3[id]')) {
    if (heading.getBoundingClientRect().top > window.innerHeight / 3) break
    current = heading.id
  }
  activeHeading.value = current
}

const onScroll = () => {
  if (!frame) frame = window.requestAnimationFrame(updateReadingState)
}

watch(slug, fetchArticle, { immediate: true })

watch(safeHtml, async () => {
  await nextTick()
  updateReadingState()
  // Hash targets only exist once the async body has rendered.
  if (route.hash) document.getElementById(route.hash.slice(1))?.scrollIntoView()
})

onMounted(() => {
  load()
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', onScroll)
})

onBeforeUnmount(() => {
  controller?.abort()
  window.cancelAnimationFrame(frame)
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', onScroll)
})

usePageMeta(() => {
  const path = `/articles/${slug.value}`
  if (!article.value) return { title: mastheadTitle.value, path, noindex: status.value !== 'loading' }
  return {
    title: article.value.title,
    description: article.value.excerpt,
    path,
    type: 'article',
    structuredData: {
      '@type': 'BlogPosting',
      headline: article.value.title,
      description: article.value.excerpt,
      datePublished: article.value.publishedAt,
      keywords: article.value.tags.join(', '),
      articleSection: article.value.category,
      url: `${siteBaseUrl()}${path}`,
      author: { '@type': 'Person', name: siteProfile.name },
    },
  }
})
</script>

<template>
  <div class="reading-progress" :style="{ transform: `scaleX(${progress})` }" aria-hidden="true" />

  <PageMasthead
    :index="article ? `ARTICLE · ${article.category.toUpperCase()}` : 'ARTICLE'"
    :title="mastheadTitle"
    :description="article?.excerpt"
    :crumbs="[
      { label: '首页', to: '/' },
      article ? { label: article.category, to: topicLink } : { label: '文章', to: { name: 'home', hash: '#articles' } },
    ]"
  >
    <p v-if="article" class="article-meta">
      <time :datetime="article.publishedAt">{{ formatChineseDate(article.publishedAt) }}</time>
      <span aria-hidden="true">·</span>
      <span>{{ article.readMinutes }} 分钟阅读</span>
    </p>
  </PageMasthead>

  <div class="page-sheet article-sheet">
    <div v-if="status === 'loading'" class="article-skeleton" aria-live="polite">
      <span v-for="line in 6" :key="line" class="skeleton-line" />
      <p class="sr-only">正在加载文章…</p>
    </div>

    <div v-else-if="status === 'error'" class="page-state" role="alert">
      <p>{{ errorMessage }}</p>
      <button type="button" @click="fetchArticle">
        <Icon :icon="RotateCw" :size="18" />
        重新加载
      </button>
    </div>

    <div v-else-if="status === 'missing'" class="page-state">
      <p>这篇文章可能已经搬家，或者链接有误。</p>
      <div class="page-state-actions">
        <RouterLink to="/">回到首页</RouterLink>
        <RouterLink to="/archive">浏览归档</RouterLink>
      </div>
    </div>

    <template v-else-if="article">
      <div class="article-layout" :class="{ 'article-layout--no-toc': toc.length === 0 }">
        <article v-if="article.content" ref="body" class="prose" v-html="safeHtml" />
        <article v-else ref="body" class="prose">
          <p>{{ article.excerpt }}</p>
          <p class="prose-note">正文整理中，敬请期待。</p>
        </article>

        <aside v-if="toc.length" class="article-aside">
          <nav class="toc" aria-label="文章目录">
            <p class="toc-title">目录</p>
            <a
              v-for="heading in toc"
              :key="heading.id"
              :href="`#${heading.id}`"
              :class="{ 'toc-sub': heading.depth === 3, active: activeHeading === heading.id }"
            >
              {{ heading.text }}
            </a>
          </nav>
        </aside>
      </div>

      <footer class="article-footer">
        <ul class="article-tags" aria-label="文章标签">
          <li v-for="tag in article.tags" :key="tag"># {{ tag }}</li>
        </ul>

        <nav v-if="adjacent.newer || adjacent.older" class="adjacent" aria-label="相邻文章">
          <RouterLink v-if="adjacent.newer" class="adjacent-link" :to="`/articles/${adjacent.newer.slug}`">
            <span><Icon :icon="ArrowLeft" :size="16" /> 较新一篇</span>
            <strong>{{ adjacent.newer.title }}</strong>
          </RouterLink>
          <span v-else />
          <RouterLink v-if="adjacent.older" class="adjacent-link adjacent-link--older" :to="`/articles/${adjacent.older.slug}`">
            <span>较早一篇 <Icon :icon="ArrowRight" :size="16" /></span>
            <strong>{{ adjacent.older.title }}</strong>
          </RouterLink>
        </nav>
      </footer>

      <section v-if="related.length" class="related" aria-labelledby="related-heading">
        <h2 id="related-heading">相关阅读</h2>
        <div class="related-grid">
          <RouterLink v-for="item in related" :key="item.id" class="related-card" :to="`/articles/${item.slug}`">
            <span>{{ item.category }} · {{ formatChineseDate(item.publishedAt) }}</span>
            <strong>{{ item.title }}</strong>
            <p>{{ item.excerpt }}</p>
          </RouterLink>
        </div>
      </section>
    </template>
  </div>

  <NewsletterBand />
</template>

<style scoped>
.reading-progress {
  position: fixed;
  z-index: 60;
  top: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background: var(--blue);
  transform-origin: left;
  pointer-events: none;
}

.article-meta {
  display: flex;
  gap: 10px;
  margin: 20px 0 0;
  font-family: var(--font-mono);
  font-size: 13px;
  opacity: 0.85;
}

.article-sheet { padding-top: 44px; }

.article-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(190px, 230px);
  column-gap: clamp(36px, 3vw, 56px);
  padding-right: clamp(56px, 6vw, 120px);
}

.article-layout--no-toc {
  grid-template-columns: minmax(0, 720px);
  padding-right: 0;
}

.article-aside { position: relative; }

.toc {
  position: sticky;
  top: 28px;
  display: grid;
  gap: 2px;
  padding-left: 18px;
  border-left: 1px solid var(--line);
}

.toc-title {
  margin: 0 0 10px;
  color: var(--muted);
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.14em;
}

.toc a {
  position: relative;
  padding: 6px 0;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.45;
  text-decoration: none;
  transition: color 160ms ease;
}

.toc a.toc-sub { padding-left: 14px; font-size: 13px; }
.toc a:hover { color: var(--ink); }

.toc a.active {
  color: var(--blue);
  font-weight: 700;
}

.toc a.active::before {
  position: absolute;
  top: 6px;
  bottom: 6px;
  left: -19px;
  width: 2px;
  background: var(--blue);
  content: '';
}

.prose {
  min-width: 0;
  color: #1b2737;
  font-size: 17px;
  line-height: 1.9;
}

.prose > :deep(:first-child) { margin-top: 0; }

.prose :deep(h2),
.prose :deep(h3) {
  scroll-margin-top: 28px;
  color: var(--ink);
  letter-spacing: -0.035em;
}

.prose :deep(h2) {
  margin: 2.2em 0 0.7em;
  font-size: 26px;
  line-height: 1.3;
}

.prose :deep(h3) {
  margin: 1.8em 0 0.5em;
  font-size: 20px;
  line-height: 1.4;
}

.prose :deep(p),
.prose :deep(ul),
.prose :deep(ol) { margin: 0 0 1.15em; }

.prose :deep(li) { margin: 0.3em 0; }
.prose :deep(strong) { color: var(--ink); }

.prose :deep(a) {
  color: var(--blue);
  text-decoration-thickness: 1px;
  text-underline-offset: 3px;
}

.prose :deep(:not(pre) > code) {
  padding: 0.12em 0.4em;
  background: var(--paper-blue);
  color: #0f3e9c;
  font-family: var(--font-mono);
  font-size: 0.86em;
}

.prose :deep(pre) {
  margin: 1.5em 0;
  padding: 18px 20px;
  overflow-x: auto;
  background: #0d2350;
  color: #d6ebff;
  font-family: var(--font-mono);
  font-size: 13.5px;
  line-height: 1.7;
}

.prose :deep(blockquote) {
  margin: 1.6em 0;
  padding: 4px 0 4px 20px;
  border-left: 3px solid var(--blue);
  color: var(--muted);
  font-size: 18px;
}

.prose :deep(blockquote p) { margin: 0; }

.prose :deep(table) {
  width: 100%;
  margin: 1.5em 0;
  border-collapse: collapse;
  font-size: 15px;
}

.prose :deep(th),
.prose :deep(td) {
  padding: 10px 12px;
  border-bottom: 1px solid var(--line);
  text-align: left;
}

.prose :deep(th) { background: var(--paper-blue); }

.prose-note { color: var(--muted); }

.article-footer {
  max-width: 720px;
  margin-top: 44px;
  padding-top: 26px;
  border-top: 1px solid var(--line);
}

.article-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.article-tags li {
  padding: 5px 10px;
  background: var(--paper-blue);
  color: #31517f;
  font-family: var(--font-mono);
  font-size: 12px;
}

.adjacent {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 32px;
}

.adjacent-link {
  display: grid;
  gap: 8px;
  padding: 18px 20px;
  border: 1px solid var(--line);
  color: var(--ink);
  text-decoration: none;
  transition: border-color 160ms ease, background 160ms ease;
}

.adjacent-link:hover {
  border-color: var(--blue);
  background: var(--paper-blue);
}

.adjacent-link span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--muted);
  font-size: 12px;
}

.adjacent-link strong { font-size: 16px; line-height: 1.45; }
.adjacent-link--older { text-align: right; }
.adjacent-link--older span { justify-content: flex-end; }

.related {
  margin-top: 56px;
  padding-top: 28px;
  border-top: 1px solid var(--line);
}

.related h2 {
  margin: 0 0 20px;
  font-size: 24px;
  letter-spacing: -0.04em;
}

.related-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.related-card {
  display: grid;
  align-content: start;
  gap: 10px;
  padding: 20px;
  background: var(--paper-blue);
  color: var(--ink);
  text-decoration: none;
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.related-card:hover {
  box-shadow: 0 14px 34px rgba(16, 40, 80, 0.1);
  transform: translateY(-3px);
}

.related-card span { color: var(--muted); font-family: var(--font-mono); font-size: 11px; }
.related-card strong { font-size: 17px; line-height: 1.4; }

.related-card p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.article-skeleton {
  max-width: 720px;
  display: grid;
  gap: 16px;
  padding: 12px 0 40px;
}

.skeleton-line {
  height: 12px;
  background: var(--line);
  animation: pulse 1.2s ease-in-out infinite alternate;
}

.skeleton-line:nth-child(3n) { width: 64%; }

@keyframes pulse { to { opacity: 0.35; } }

@media (max-width: 1080px) {
  .article-layout {
    grid-template-columns: minmax(0, 1fr);
    padding-right: 0;
  }
  .article-aside { display: none; }
  .related-grid { grid-template-columns: 1fr; }
}

@media (max-width: 600px) {
  .prose { font-size: 16px; }
  .prose :deep(h2) { font-size: 22px; }
  .adjacent { grid-template-columns: 1fr; }
  .adjacent-link--older { text-align: left; }
  .adjacent-link--older span { justify-content: flex-start; }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-line { animation: none; }
}
</style>
