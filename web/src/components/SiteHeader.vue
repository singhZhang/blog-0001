<script setup lang="ts">
import { RouterLink, type RouteLocationRaw } from 'vue-router'
import { Search } from '@lucide/vue'
import Icon from './ui/Icon.vue'
import type { NavSection } from '../router'

defineProps<{ isNight: boolean, activeSection?: NavSection }>()
defineEmits<{ openSearch: [] }>()

const navItems: { section: NavSection, label: string, to: RouteLocationRaw }[] = [
  { section: 'articles', label: '文章', to: { name: 'home', hash: '#articles' } },
  { section: 'topics', label: '主题', to: { name: 'topics' } },
  { section: 'archive', label: '归档', to: { name: 'archive' } },
  { section: 'about', label: '关于', to: { name: 'about' } },
]
</script>

<template>
  <header
    class="absolute top-0 left-1/2 z-20 grid h-[84px] w-[min(calc(100%-48px),1400px)] -translate-x-1/2 grid-cols-[1fr_auto_1fr] items-center text-ink max-[760px]:h-[72px] max-[760px]:w-[min(calc(100%-32px),1400px)] max-[760px]:grid-cols-[1fr_auto]"
    :class="isNight && 'text-[#f2f7fb]'"
  >
    <RouterLink class="justify-self-start" to="/" aria-label="零玖肆玖肆首页">
      <img
        class="block h-auto w-[clamp(210px,17vw,248px)] transition-[filter,opacity] duration-200 hover:opacity-75 max-[760px]:w-[196px]"
        :class="isNight && 'brightness-0 invert'"
        src="/brand/logo-09494.png"
        alt="零玖肆玖肆"
        width="300"
        height="64"
        decoding="async"
        fetchpriority="high"
      >
    </RouterLink>
    <nav class="flex items-center gap-[clamp(28px,4vw,52px)] max-[760px]:hidden" aria-label="主要导航">
      <RouterLink
        v-for="item in navItems"
        :key="item.section"
        :to="item.to"
        class="relative text-[15px] no-underline after:absolute after:inset-x-0 after:-bottom-2 after:h-px after:origin-right after:scale-x-0 after:bg-current after:transition-transform after:duration-200 hover:after:origin-left hover:after:scale-x-100"
        :class="activeSection === item.section && 'font-bold after:origin-left after:scale-x-100'"
        :aria-current="activeSection === item.section ? 'page' : undefined"
      >
        {{ item.label }}
      </RouterLink>
    </nav>
    <button
      class="grid size-11 place-items-center justify-self-end"
      type="button"
      aria-label="搜索文章"
      @click="$emit('openSearch')"
    >
      <Icon :icon="Search" :size="24" />
    </button>
  </header>
</template>
