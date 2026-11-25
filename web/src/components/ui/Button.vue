<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '../../lib/cn'

const props = withDefaults(defineProps<{
  variant?: 'default' | 'secondary' | 'danger' | 'ghost'
  size?: 'md' | 'sm'
  type?: 'button' | 'submit'
  disabled?: boolean
}>(), {
  variant: 'default',
  size: 'md',
  type: 'button',
  disabled: false,
})

const variantClass = computed(() => ({
  default: 'border-ink bg-ink text-white hover:bg-blue hover:border-blue',
  secondary: 'border-ink bg-transparent text-ink hover:border-blue hover:text-blue',
  danger: 'border-[#9d3b3b] bg-transparent text-[#9d3b3b] hover:bg-[#fff6f6]',
  ghost: 'border-transparent bg-transparent text-muted hover:text-blue',
}[props.variant]))

const sizeClass = computed(() => props.size === 'sm'
  ? 'min-h-8 px-3 py-0 text-[13px]'
  : 'min-h-10 px-4 text-sm')
</script>

<template>
  <button
    :type="type"
    :disabled="disabled"
    :class="cn(
      'inline-flex items-center justify-center gap-2 border leading-none transition-colors duration-150 disabled:cursor-progress disabled:opacity-50',
      sizeClass,
      variantClass,
    )"
  >
    <slot />
  </button>
</template>
