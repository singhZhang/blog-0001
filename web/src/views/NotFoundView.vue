<script setup lang="ts">
import { useRoute, RouterLink } from 'vue-router'
import PageMasthead from '../components/PageMasthead.vue'
import { usePageMeta } from '../composables/usePageMeta'

const route = useRoute()

usePageMeta(() => ({ title: '页面不存在', path: route.path, noindex: true }))
</script>

<template>
  <PageMasthead
    index="404 / NOT FOUND"
    title="在山里走岔了路。"
    description="你要找的页面不存在，可能已经被移动，或者链接写错了。"
  />
  <div class="page-sheet">
    <div class="page-state">
      <p>不妨从这些地方重新出发：</p>
      <div class="page-state-actions">
        <RouterLink to="/">回到首页</RouterLink>
        <RouterLink to="/archive">浏览归档</RouterLink>
        <RouterLink to="/topics">按主题浏览</RouterLink>
      </div>
    </div>
  </div>
</template>
