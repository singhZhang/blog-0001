import test from 'node:test'
import assert from 'node:assert/strict'
import { getCelestialAppearance, getCelestialPosition, getTimeScene } from '../src/lib/time-scene.mjs'

test('selects dawn, day, dusk and night at boundary hours', () => {
  assert.equal(getTimeScene(new Date(2026, 9, 27, 5, 0)).key, 'dawn')
  assert.equal(getTimeScene(new Date(2026, 9, 27, 8, 0)).key, 'day')
  assert.equal(getTimeScene(new Date(2026, 9, 27, 17, 0)).key, 'dusk')
  assert.equal(getTimeScene(new Date(2026, 9, 27, 20, 0)).key, 'night')
})

test('formats local time with a Chinese period label', () => {
  const scene = getTimeScene(new Date(2026, 9, 27, 11, 38))
  assert.equal(scene.dateLabel, '2026年10月27日')
  assert.equal(scene.timeLabel, '11:38')
  assert.equal(scene.periodLabel, '上午')
})

test('moves the sun through a bounded daylight arc', () => {
  const morning = getCelestialPosition(new Date(2026, 9, 27, 8, 0))
  const noon = getCelestialPosition(new Date(2026, 9, 27, 12, 0))
  const evening = getCelestialPosition(new Date(2026, 9, 27, 17, 0))

  assert.ok(morning.x < noon.x)
  assert.ok(noon.x < evening.x)
  assert.ok(noon.y < morning.y)
  assert.ok(noon.y < evening.y)
  assert.ok([morning, noon, evening].every(({ x, y }) => x >= 55 && x <= 90 && y >= 10 && y <= 22))
})

test('keeps every reading surface translucent enough to reveal the landscape', () => {
  for (const hour of [6, 12, 18, 22]) {
    const scene = getTimeScene(new Date(2026, 9, 27, hour, 0))
    assert.equal(typeof scene.surface, 'string')
    assert.equal(typeof scene.surfaceBorder, 'string')
    const alpha = Number(scene.surface.match(/,\s*(0?\.\d+)\)$/)?.[1])
    assert.ok(alpha >= 0.60, `${scene.key} surface alpha should preserve text contrast`)
    assert.ok(alpha <= 0.68, `${scene.key} surface alpha should reveal the landscape`)
  }
})

test('sunlight is paler and brighter at noon than at sunrise and sunset', () => {
  const dawn = getCelestialAppearance(new Date(2026, 9, 27, 5, 0))
  const noon = getCelestialAppearance(new Date(2026, 9, 27, 12, 30))
  const dusk = getCelestialAppearance(new Date(2026, 9, 27, 19, 0))
  for (const lowSun of [dawn, dusk]) {
    assert.ok(noon.brightness > lowSun.brightness)
    assert.ok(noon.saturation < lowSun.saturation)
    assert.ok(noon.opacity > lowSun.opacity)
    assert.ok(lowSun.hue < noon.hue)
  }
})

test('sunlight changes by the minute without jumping at day and dusk boundaries', () => {
  for (const hour of [8, 17]) {
    const before = getCelestialAppearance(new Date(2026, 9, 27, hour - 1, 59))
    const at = getCelestialAppearance(new Date(2026, 9, 27, hour, 0))
    const after = getCelestialAppearance(new Date(2026, 9, 27, hour, 1))
    assert.notDeepEqual(before, at)
    assert.notDeepEqual(at, after)
    assert.ok(Math.abs(before.brightness - after.brightness) < 0.01)
    assert.ok(Math.abs(before.hue - after.hue) < 0.2)
  }
})

test('night uses neutral moonlight and exposes valid appearance values across the whole day', () => {
  for (let minute = 0; minute < 1440; minute++) {
    const date = new Date(2026, 9, 27, 0, minute)
    const scene = getTimeScene(date)
    const light = scene.celestialAppearance
    assert.ok(Object.values(light).every(Number.isFinite))
    assert.ok(light.opacity > 0 && light.opacity <= 1)
    if (scene.key === 'night') {
      assert.equal(light.sepia, 0)
      assert.equal(light.hue, 0)
    } else {
      assert.ok(light.sepia > 0 && light.sepia < 1)
      assert.ok(scene.celestial.y >= 10 && scene.celestial.y <= 22)
    }
  }
})
