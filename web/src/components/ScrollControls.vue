<script setup lang="ts">
import { ArrowDown, ArrowUp } from '@lucide/vue'
import Icon from './ui/Icon.vue'
</script>

<template>
  <nav class="fixed right-[max(18px,env(safe-area-inset-right))] bottom-[max(18px,env(safe-area-inset-bottom))] z-40 grid gap-2" aria-label="页面快速滚动">
    <a class="grid size-[46px] place-items-center border border-ink/25 bg-paper/85 text-ink shadow-[0_10px_28px_rgba(10,22,36,0.12)] backdrop-blur-md" href="#top" aria-label="一键到顶" title="一键到顶">
      <Icon :icon="ArrowUp" :size="20" />
    </a>
    <a class="grid size-[46px] place-items-center border border-ink/25 bg-paper/85 text-ink shadow-[0_10px_28px_rgba(10,22,36,0.12)] backdrop-blur-md" href="#page-end" aria-label="一键到底" title="一键到底">
      <Icon :icon="ArrowDown" :size="20" />
    </a>
  </nav>
</template>
