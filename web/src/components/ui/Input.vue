<script setup lang="ts">
defineOptions({ inheritAttrs: false })

defineProps<{
  modelValue?: string
}>()

defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <input
    :value="modelValue"
    class="h-11 w-full border border-line bg-white px-3.5 text-[15px] text-ink outline-none transition-colors placeholder:text-muted focus:border-blue"
    v-bind="$attrs"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  >
</template>
