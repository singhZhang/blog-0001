<script setup lang="ts">
import { inject } from 'vue'
import ParallaxHero from '../components/ParallaxHero.vue'
import ArticleList from '../components/ArticleList.vue'
import NewsletterBand from '../components/NewsletterBand.vue'
import { useArticles } from '../composables/useArticles'
import { siteBaseUrl, usePageMeta } from '../composables/usePageMeta'
import { timeSceneKey } from '../composables/useTimeScene'
import { siteProfile } from '../data/site'

const { scene, sceneStyle } = inject(timeSceneKey)!
const { articles, loading, error, reload } = useArticles()

usePageMeta(() => ({
  path: '/',
  structuredData: {
    '@type': 'Blog',
    name: siteProfile.name,
    description: siteProfile.description,
    url: `${siteBaseUrl()}/`,
    blogPost: articles.value.map((article) => ({
      '@type': 'BlogPosting',
      headline: article.title,
      description: article.excerpt,
      datePublished: article.publishedAt,
      url: `${siteBaseUrl()}/articles/${article.slug}`,
    })),
  },
}))
</script>

<template>
  <ParallaxHero :scene="scene" :scene-style="sceneStyle" />
  <ArticleList :articles="articles" :loading="loading" :error="error" @retry="reload" />
  <NewsletterBand />
</template>
