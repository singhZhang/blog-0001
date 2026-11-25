<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowUpRight, Search, X } from '@lucide/vue'
import { motion } from 'motion-v'
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import Icon from './ui/Icon.vue'
import type { Article } from '../types/blog'

const props = defineProps<{ open: boolean, articles: Article[] }>()
const emit = defineEmits<{ close: [] }>()
const query = ref('')
const input = ref<HTMLInputElement | null>(null)

const results = computed(() => {
  const keyword = query.value.trim().toLocaleLowerCase()
  if (!keyword) return props.articles.slice(0, 4)
  return props.articles.filter((article) => [article.title, article.excerpt, article.category, ...article.tags]
    .join(' ')
    .toLocaleLowerCase()
    .includes(keyword))
})

watch(() => props.open, (open) => {
  document.body.style.overflow = open ? 'hidden' : ''
  if (open) query.value = ''
})

function onOpenChange(open: boolean) {
  if (!open) emit('close')
}
</script>

<template>
  <DialogRoot :open="open" @update:open="onOpenChange">
    <DialogPortal>
      <DialogOverlay class="fixed inset-0 z-[100] bg-[#07121d]/72 backdrop-blur-md" />
      <DialogContent
        class="fixed top-[11vh] left-1/2 z-[101] max-h-[78vh] w-[min(720px,calc(100%-48px))] -translate-x-1/2 overflow-auto bg-paper p-7 shadow-[0_24px_80px_rgba(4,13,24,0.28)] outline-none max-[560px]:inset-0 max-[560px]:h-full max-[560px]:max-h-none max-[560px]:w-full max-[560px]:translate-x-0"
        @open-auto-focus="input?.focus()"
      >
        <motion.div
          :initial="{ opacity: 0, y: 10 }"
          :animate="{ opacity: 1, y: 0 }"
          :transition="{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }"
        >
          <header class="flex items-start justify-between gap-6">
            <div>
              <p class="mb-1 font-mono text-[10px] tracking-[0.13em] text-muted">SEARCH / NOTES</p>
              <DialogTitle class="text-[30px] tracking-[-0.05em]">搜索技术记录</DialogTitle>
              <DialogDescription class="sr-only">按标题、主题或关键词查找文章</DialogDescription>
            </div>
            <DialogClose class="grid size-11 place-items-center text-ink" aria-label="关闭搜索">
              <Icon :icon="X" :size="22" />
            </DialogClose>
          </header>

          <label class="mt-7 mb-3.5 grid h-14 grid-cols-[auto_1fr] items-center gap-3 border border-[#aeb7c2] px-4 text-muted focus-within:border-blue">
            <span class="sr-only">搜索关键词</span>
            <Icon :icon="Search" :size="22" />
            <input
              ref="input"
              v-model="query"
              type="search"
              placeholder="输入标题、主题或关键词"
              class="w-full border-0 bg-transparent text-[17px] text-ink outline-none"
            >
          </label>

          <div class="search-results" aria-live="polite">
            <RouterLink
              v-for="article in results"
              :key="article.id"
              class="grid grid-cols-[100px_1fr_auto] items-center gap-4 border-b border-line px-1 py-4 text-ink no-underline hover:text-blue max-[560px]:grid-cols-[1fr_auto]"
              :to="`/articles/${article.slug}`"
              @click="emit('close')"
            >
              <span class="font-mono text-[11px] text-muted max-[560px]:col-span-2">{{ article.category }}</span>
              <strong class="text-base font-semibold">{{ article.title }}</strong>
              <Icon :icon="ArrowUpRight" :size="20" />
            </RouterLink>
            <p v-if="results.length === 0" class="py-11 text-center text-muted">没有找到相关记录，换个关键词试试。</p>
          </div>
        </motion.div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
