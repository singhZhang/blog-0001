<script setup lang="ts">
import { computed, provide, ref, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { motion } from 'motion-v'
import SiteHeader from './components/SiteHeader.vue'
import SiteFooter from './components/SiteFooter.vue'
import SearchDialog from './components/SearchDialog.vue'
import ScrollControls from './components/ScrollControls.vue'
import { timeSceneKey, useTimeScene } from './composables/useTimeScene'
import { useArticles } from './composables/useArticles'

const route = useRoute()
const searchOpen = ref(false)
const timeScene = useTimeScene()
const { scene, sceneStyle } = timeScene
const { articles, load } = useArticles()
const isAdmin = computed(() => route.path.startsWith('/admin'))

provide(timeSceneKey, timeScene)
watch(() => route.path, (path) => {
  if (!path.startsWith('/admin')) load()
}, { immediate: true })
</script>

<template>
  <motion.div
    v-if="isAdmin"
    :key="route.fullPath"
    :initial="{ opacity: 0, y: 8 }"
    :animate="{ opacity: 1, y: 0 }"
    :transition="{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }"
  >
    <RouterView />
  </motion.div>
  <div v-else id="top" class="site-shell" :data-scene="scene.key" :style="sceneStyle">
    <SiteHeader
      :is-night="scene.key === 'night'"
      :active-section="route.meta.section"
      @open-search="searchOpen = true"
    />
    <main>
      <motion.div
        :key="route.fullPath"
        :initial="{ opacity: 0, y: 10 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: 0.32, ease: [0.22, 1, 0.36, 1] }"
      >
        <RouterView />
      </motion.div>
    </main>
    <SiteFooter />
    <ScrollControls />
    <SearchDialog
      :open="searchOpen"
      :articles="articles"
      @close="searchOpen = false"
    />
  </div>
</template>
