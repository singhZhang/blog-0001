<script setup lang="ts">
import { RouterLink } from 'vue-router'

const year = new Date().getFullYear()
</script>

<template>
  <footer id="page-end" class="mx-auto flex min-h-[84px] w-[min(calc(100%-48px),1400px)] items-center justify-between gap-6 text-xs text-muted max-[600px]:w-[calc(100%-24px)] max-[600px]:flex-col max-[600px]:items-start max-[600px]:py-7">
    <p>© {{ year }} 零玖肆玖肆。记录日常，也沉淀思考。</p>
    <nav class="flex gap-7" aria-label="页脚导航">
      <RouterLink class="hover:text-blue" :to="{ name: 'home', hash: '#articles' }">文章</RouterLink>
      <RouterLink class="hover:text-blue" to="/topics">主题</RouterLink>
      <RouterLink class="hover:text-blue" to="/archive">归档</RouterLink>
      <RouterLink class="hover:text-blue" to="/about">关于</RouterLink>
      <RouterLink class="hover:text-blue" to="/admin">管理</RouterLink>
    </nav>
  </footer>
</template>
