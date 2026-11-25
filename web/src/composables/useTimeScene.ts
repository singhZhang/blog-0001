import { computed, onBeforeUnmount, onMounted, ref, type InjectionKey } from 'vue'
import { getTimeScene } from '../lib/time-scene.mjs'

export const timeSceneKey: InjectionKey<ReturnType<typeof useTimeScene>> = Symbol('timeScene')

export function useTimeScene() {
  const now = ref(new Date())
  let timer = 0

  onMounted(() => {
    timer = window.setInterval(() => {
      now.value = new Date()
    }, 60_000)
  })

  onBeforeUnmount(() => window.clearInterval(timer))

  const scene = computed(() => getTimeScene(now.value))
  const sceneStyle = computed(() => ({
    '--scene-sky': scene.value.sky,
    '--scene-tint': scene.value.tint,
    '--scene-ink': scene.value.ink,
    '--scene-accent': scene.value.accent,
    '--scene-surface': scene.value.surface,
    '--scene-surface-border': scene.value.surfaceBorder,
    '--landscape-filter': scene.value.landscapeFilter,
    '--celestial-x': `${scene.value.celestial.x}%`,
    '--celestial-y': `${scene.value.celestial.y}%`,
    '--celestial-filter': `sepia(${scene.value.celestialAppearance.sepia}) saturate(${scene.value.celestialAppearance.saturation}) hue-rotate(${scene.value.celestialAppearance.hue}deg) brightness(${scene.value.celestialAppearance.brightness})`,
    '--celestial-opacity': String(scene.value.celestialAppearance.opacity),
  }))

  return { now, scene, sceneStyle }
}
