<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowRight, Mail } from '@lucide/vue'
import Icon from '../components/ui/Icon.vue'
import PageMasthead from '../components/PageMasthead.vue'
import { useArticles } from '../composables/useArticles'
import { useCategories } from '../composables/useCategories'
import { usePageMeta } from '../composables/usePageMeta'
import { aboutProfile, siteProfile } from '../data/site'

const { articles, load } = useArticles()
const { categories, load: loadCategories } = useCategories()
const stats = computed(() => {
  const topics = new Set(categories.value.map((category) => category.name))
  for (const article of articles.value) topics.add(article.category)
  return [
    { label: '篇文章', value: articles.value.length },
    { label: '个主题', value: topics.size },
    { label: '个标签', value: new Set(articles.value.flatMap((article) => article.tags)).size },
  ]
})

onMounted(() => {
  load()
  loadCategories()
})

usePageMeta(() => ({
  title: '关于',
  description: `关于${siteProfile.name}：写些什么、为什么写，以及如何联系我。`,
  path: '/about',
  structuredData: {
    '@type': 'AboutPage',
    name: `关于${siteProfile.name}`,
    mainEntity: { '@type': 'Person', name: siteProfile.name, email: siteProfile.email },
  },
}))
</script>

<template>
  <PageMasthead
    index="ABOUT / 关于"
    title="记录日常，也沉淀思考。"
    :description="aboutProfile.intro[0]"
    :crumbs="[{ label: '首页', to: '/' }]"
  />

  <div class="page-sheet about-sheet">
    <section class="about-row" aria-labelledby="about-intro">
      <h2 id="about-intro">你好</h2>
      <div class="about-intro">
        <p v-for="paragraph in aboutProfile.intro.slice(1)" :key="paragraph">{{ paragraph }}</p>
        <dl v-if="articles.length" class="about-stats">
          <div v-for="stat in stats" :key="stat.label">
            <dt>{{ stat.label }}</dt>
            <dd>{{ stat.value }}</dd>
          </div>
        </dl>
      </div>
    </section>

    <section class="about-row" aria-labelledby="about-writing">
      <h2 id="about-writing">写些什么</h2>
      <div>
        <ul class="writing-grid">
          <li v-for="item in aboutProfile.writing" :key="item.title">
            <strong>{{ item.title }}</strong>
            <p>{{ item.description }}</p>
          </li>
        </ul>
        <RouterLink class="about-link" to="/topics">
          按主题浏览
          <Icon :icon="ArrowRight" :size="18" />
        </RouterLink>
      </div>
    </section>

    <section class="about-row" aria-labelledby="about-principles">
      <h2 id="about-principles">写作原则</h2>
      <ol class="principles">
        <li v-for="principle in aboutProfile.principles" :key="principle">{{ principle }}</li>
      </ol>
    </section>

    <section class="about-row" aria-labelledby="about-colophon">
      <h2 id="about-colophon">关于本站</h2>
      <div class="colophon">
        <p v-for="line in aboutProfile.colophon" :key="line">{{ line }}</p>
      </div>
    </section>

    <section class="about-row" aria-labelledby="about-contact">
      <h2 id="about-contact">联系</h2>
      <div>
        <p class="contact-copy">欢迎来信交流想法、指出错误，或者只是打个招呼。</p>
        <a class="contact-mail" :href="`mailto:${siteProfile.email}`">
          <Icon :icon="Mail" :size="20" />
          {{ siteProfile.email }}
        </a>
      </div>
    </section>
  </div>
</template>

<style scoped>
.about-sheet { padding-top: 12px; }

.about-row {
  display: grid;
  grid-template-columns: 200px minmax(0, 760px);
  gap: 40px;
  padding: 36px 0;
  border-bottom: 1px solid var(--line);
}

.about-row:last-child { border-bottom: 0; }

.about-row h2 {
  margin: 0;
  color: var(--muted);
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.about-intro p {
  margin: 0 0 1em;
  font-size: 19px;
  line-height: 1.85;
}

.about-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 40px;
  margin: 28px 0 0;
}

.about-stats div { display: flex; flex-direction: column-reverse; gap: 4px; }
.about-stats dt { color: var(--muted); font-size: 13px; }

.about-stats dd {
  margin: 0;
  color: var(--blue);
  font-family: var(--font-mono);
  font-size: 34px;
  font-weight: 700;
  line-height: 1;
}

.writing-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin: 0 0 20px;
  padding: 0;
  list-style: none;
}

.writing-grid li {
  padding: 20px;
  background: var(--paper-blue);
}

.writing-grid strong { font-size: 17px; }

.writing-grid p {
  margin: 8px 0 0;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.7;
}

.about-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
}

.about-link:hover { color: var(--blue); }

.principles {
  display: grid;
  gap: 14px;
  margin: 0;
  padding: 0;
  list-style: none;
  counter-reset: principle;
}

.principles li {
  display: grid;
  grid-template-columns: 40px 1fr;
  font-size: 17px;
  line-height: 1.7;
  counter-increment: principle;
}

.principles li::before {
  color: var(--blue);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 2.2;
  content: counter(principle, decimal-leading-zero);
}

.colophon p,
.contact-copy {
  margin: 0 0 0.8em;
  color: #34445a;
  font-size: 16px;
  line-height: 1.8;
}

.contact-mail {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 18px;
  border: 1px solid var(--ink);
  color: var(--ink);
  font-family: var(--font-mono);
  font-size: 15px;
  text-decoration: none;
  transition: background 160ms ease, color 160ms ease;
}

.contact-mail:hover {
  background: var(--ink);
  color: var(--paper);
}

@media (max-width: 900px) {
  .about-row { grid-template-columns: 1fr; gap: 16px; }
  .writing-grid { grid-template-columns: 1fr; }
}

@media (max-width: 600px) {
  .about-intro p { font-size: 17px; }
}
</style>
