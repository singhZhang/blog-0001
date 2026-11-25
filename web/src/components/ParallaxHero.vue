<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

defineProps<{
  scene: {
    key: string
    name: string
    dateLabel: string
    timeLabel: string
    periodLabel: string
  }
  sceneStyle: Record<string, string>
}>()

const hero = ref<HTMLElement | null>(null)
const sky = ref<HTMLImageElement | null>(null)
const far = ref<HTMLImageElement | null>(null)
const river = ref<HTMLImageElement | null>(null)
const foreground = ref<HTMLImageElement | null>(null)
const celestial = ref<HTMLDivElement | null>(null)
const copy = ref<HTMLDivElement | null>(null)
let context: gsap.Context | null = null

onMounted(() => {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  gsap.registerPlugin(ScrollTrigger)
  context = gsap.context(() => {
    const timeline = gsap.timeline({
      scrollTrigger: {
        trigger: hero.value,
        start: 'top top',
        end: 'bottom top',
        scrub: 0.9,
      },
    })
    timeline
      .to(sky.value, { yPercent: 5, scale: 1.025, ease: 'none' }, 0)
      .to(celestial.value, { yPercent: 14, ease: 'none' }, 0)
      .to(far.value, { yPercent: 9, scale: 1.035, ease: 'none' }, 0)
      .to(river.value, { yPercent: 16, scale: 1.045, ease: 'none' }, 0)
      .to(foreground.value, { yPercent: 24, scale: 1.06, ease: 'none' }, 0)
      .to(copy.value, { yPercent: 18, opacity: 0.3, ease: 'none' }, 0)
  }, hero.value ?? undefined)
})

onBeforeUnmount(() => context?.revert())
</script>

<template>
  <section ref="hero" class="hero" :style="sceneStyle" :aria-label="`${scene.name}山水景观`">
    <div class="scene" aria-hidden="true">
      <img ref="sky" class="scene-layer scene-layer--sky" src="/images/parallax/sky-day.png" alt="" fetchpriority="high" />
      <img ref="far" class="scene-layer scene-layer--land scene-layer--far" src="/images/parallax/mountains-far.png" alt="" fetchpriority="high" />
      <img ref="river" class="scene-layer scene-layer--land scene-layer--river" src="/images/parallax/river-mid.png" alt="" fetchpriority="high" />
      <div ref="celestial" class="celestial">
        <img :src="scene.key === 'night' ? '/images/parallax/moon.png' : '/images/parallax/sun.png'" alt="" />
      </div>
      <div class="scene-tint" />
      <img ref="foreground" class="scene-layer scene-layer--land scene-layer--foreground" src="/images/parallax/grass-foreground.png" alt="" fetchpriority="high" />
    </div>

    <div ref="copy" class="hero-copy">
      <p class="eyebrow" aria-live="polite">
        <span>TECHNICAL NOTES · {{ scene.name }}</span>
        <time class="eyebrow-time">{{ scene.dateLabel }} {{ scene.timeLabel }} · {{ scene.periodLabel }}</time>
      </p>
      <h1>记录日常，也沉淀思考。</h1>
      <p class="hero-summary">生活片段、技术实践与阅读随想。</p>
      <p class="hero-summary">把平常日子里的所见所得，慢慢整理成值得回看的文字。</p>
    </div>
  </section>
</template>

<style scoped>
.hero {
  --hero-surface-top: 114px;
  position: relative;
  min-height: 548px;
  overflow: hidden;
  background: var(--scene-sky);
  color: var(--scene-ink);
  isolation: isolate;
}

.scene,
.scene-layer,
.scene-tint {
  position: absolute;
  inset: 0;
}

.scene {
  z-index: -1;
  overflow: hidden;
}

.scene-layer {
  width: 100%;
  height: 100%;
  max-width: none;
  object-fit: cover;
  object-position: center bottom;
  pointer-events: none;
  will-change: transform;
}

.scene-layer--sky {
  object-position: center center;
  transform: scale(1.03);
}

.scene-layer--land {
  filter: var(--landscape-filter);
  transition: filter 1.2s ease;
}

.scene-layer--far { z-index: 2; }
.scene-layer--river { z-index: 4; }
.scene-layer--foreground { z-index: 6; }

.scene-tint {
  z-index: 5;
  background: var(--scene-tint);
  transition: background 1.2s ease;
}

.celestial {
  position: absolute;
  /* Above the distant mountains, behind the riverbank trees and foreground. */
  z-index: 3;
  top: var(--celestial-y);
  left: var(--celestial-x);
  width: clamp(72px, 7vw, 116px);
  aspect-ratio: 1;
  transform: translate(-50%, -50%);
  transition: top 1.2s ease, left 1.2s ease;
  will-change: transform;
}

.celestial img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: var(--celestial-filter, none);
  opacity: var(--celestial-opacity, 1);
  transition: filter 1.2s ease, opacity 1.2s ease;
}

.hero-copy {
  position: absolute;
  z-index: 8;
  top: calc(var(--hero-surface-top) + 18px);
  left: max(44px, calc((100vw - var(--content-wide)) / 2 + 15px));
  width: min(790px, calc(100% - 96px));
  isolation: isolate;
  will-change: transform, opacity;
}

.hero-copy::before {
  position: absolute;
  z-index: -1;
  inset: -18px -20px -16px;
  border: 1px solid var(--scene-surface-border);
  background: var(--scene-surface);
  box-shadow: 0 12px 36px rgba(18, 42, 58, 0.07);
  backdrop-filter: blur(16px) saturate(0.78);
  -webkit-backdrop-filter: blur(16px) saturate(0.78);
  content: '';
}

.eyebrow {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin: 0 0 18px;
  color: color-mix(in srgb, var(--scene-ink) 92%, transparent);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
}

.eyebrow-time {
  display: inline-flex;
  align-items: center;
  font-family: var(--font-mono);
  font-weight: 600;
  letter-spacing: 0.06em;
}

.eyebrow-time::before {
  width: 1px;
  height: 12px;
  margin-right: 12px;
  background: color-mix(in srgb, var(--scene-ink) 28%, transparent);
  content: '';
}

h1 {
  max-width: 720px;
  margin: 0;
  font-size: clamp(42px, 4.4vw, 62px);
  font-weight: 820;
  line-height: 1.03;
  letter-spacing: -0.075em;
}

.hero-summary {
  max-width: 700px;
  margin: 18px 0 0;
  color: color-mix(in srgb, var(--scene-ink) 92%, transparent);
  font-size: clamp(16px, 1.45vw, 21px);
  line-height: 1.62;
}

@media (max-width: 900px) {
  .hero {
    --hero-surface-top: 106px;
    min-height: 590px;
  }
}

@media (max-width: 600px) {
  .hero { min-height: 620px; }
  .hero-copy {
    top: 126px;
    left: 24px;
    width: calc(100% - 48px);
  }
  .hero-copy::before { inset: -14px -12px -12px; }
  .eyebrow { margin-bottom: 14px; }
  h1 {
    font-size: clamp(42px, 13vw, 58px);
    line-height: 1.05;
    max-width: 380px;
  }
  .hero-summary {
    max-width: 390px;
    font-size: 16px;
  }
  .scene-layer { width: 182%; left: -41%; }
}

@media (prefers-reduced-motion: reduce) {
  .scene-layer,
  .hero-copy,
  .celestial { will-change: auto; }
}
</style>
