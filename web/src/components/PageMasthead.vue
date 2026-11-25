<script setup lang="ts">
import { RouterLink, type RouteLocationRaw } from 'vue-router'

defineProps<{
  index: string
  title: string
  description?: string
  crumbs?: { label: string; to: RouteLocationRaw }[]
}>()
</script>

<template>
  <section class="masthead">
    <div class="masthead-scene" aria-hidden="true">
      <img class="masthead-ridge" src="/images/parallax/mountains-far.png" alt="" decoding="async" />
    </div>
    <div class="masthead-inner">
      <nav v-if="crumbs?.length" class="crumbs" aria-label="面包屑">
        <template v-for="(crumb, position) in crumbs" :key="position">
          <span v-if="position > 0" aria-hidden="true">/</span>
          <RouterLink :to="crumb.to">{{ crumb.label }}</RouterLink>
        </template>
      </nav>
      <p class="masthead-index">{{ index }}</p>
      <h1>{{ title }}</h1>
      <p v-if="description" class="masthead-description">{{ description }}</p>
      <slot />
    </div>
  </section>
</template>

<style scoped>
.masthead {
  position: relative;
  min-height: 380px;
  padding: 150px 0 112px;
  overflow: hidden;
  background: var(--scene-sky);
  color: var(--scene-ink);
  isolation: isolate;
  transition: background 1.2s ease, color 1.2s ease;
}

.masthead-scene {
  position: absolute;
  z-index: -1;
  inset: 0;
}

.masthead-scene::after {
  position: absolute;
  inset: 0;
  background: var(--scene-tint);
  content: '';
}

.masthead-ridge {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 78%;
  max-width: none;
  object-fit: cover;
  object-position: center 82%;
  filter: var(--landscape-filter);
  opacity: 0.5;
  pointer-events: none;
}

.masthead-inner {
  width: min(calc(100% - 48px), var(--content-wide));
  margin: 0 auto;
  padding: 0 34px;
}

.crumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 18px;
  font-size: 13px;
}

.crumbs a {
  color: inherit;
  opacity: 0.78;
  text-decoration: none;
}

.crumbs a:hover { opacity: 1; text-decoration: underline; text-underline-offset: 4px; }
.crumbs span { opacity: 0.45; }

.masthead-index {
  margin: 0 0 10px;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
  opacity: 0.82;
}

h1 {
  max-width: 900px;
  margin: 0;
  font-size: clamp(34px, 3.6vw, 52px);
  font-weight: 820;
  line-height: 1.12;
  letter-spacing: -0.06em;
}

.masthead-description {
  max-width: 680px;
  margin: 16px 0 0;
  font-size: clamp(15px, 1.25vw, 18px);
  line-height: 1.7;
  opacity: 0.88;
}

@media (max-width: 760px) {
  .masthead {
    min-height: 320px;
    padding: 112px 0 92px;
  }
  .masthead-inner {
    width: calc(100% - 24px);
    padding: 0 20px;
  }
  .masthead-ridge { height: 60%; }
}
</style>
