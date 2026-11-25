const SCENES = {
  dawn: {
    key: 'dawn',
    name: '清晨',
    sky: '#d9dfe8',
    tint: 'rgba(255, 184, 125, 0.26)',
    ink: '#132738',
    accent: '#d57842',
    surface: 'rgba(245, 239, 226, 0.64)',
    surfaceBorder: 'rgba(102, 77, 57, 0.14)',
    landscapeFilter: 'saturate(.82) sepia(.1) hue-rotate(338deg) brightness(.96)',
  },
  day: {
    key: 'day',
    name: '白昼',
    sky: '#cde9f7',
    tint: 'rgba(255, 244, 207, 0.08)',
    ink: '#0b1724',
    accent: '#165dff',
    surface: 'rgba(226, 241, 247, 0.64)',
    surfaceBorder: 'rgba(50, 91, 118, 0.14)',
    landscapeFilter: 'saturate(.92) brightness(1.02)',
  },
  dusk: {
    key: 'dusk',
    name: '黄昏',
    sky: '#d8a37e',
    tint: 'rgba(173, 76, 71, 0.30)',
    ink: '#251a1d',
    accent: '#b14f3f',
    surface: 'rgba(242, 222, 201, 0.64)',
    surfaceBorder: 'rgba(104, 59, 50, 0.16)',
    landscapeFilter: 'saturate(.82) sepia(.25) hue-rotate(332deg) brightness(.84)',
  },
  night: {
    key: 'night',
    name: '夜晚',
    sky: '#17283c',
    tint: 'rgba(21, 34, 58, 0.46)',
    ink: '#eef5fb',
    accent: '#a9cfff',
    surface: 'rgba(12, 25, 41, 0.68)',
    surfaceBorder: 'rgba(218, 233, 249, 0.16)',
    landscapeFilter: 'saturate(.55) hue-rotate(175deg) brightness(.48)',
  },
}

const clamp = (value, min, max) => Math.min(max, Math.max(min, value))
const pad = (value) => String(value).padStart(2, '0')

// Local-time art direction: warm sunrise, pale midday light, deeper sunset.
const SUN_LIGHT_STOPS = [
  { hour: 5, brightness: 0.85, saturation: 1.8, hue: -22, opacity: 0.8 },
  { hour: 8, brightness: 1.05, saturation: 1.1, hue: -8, opacity: 0.95 },
  { hour: 12.5, brightness: 1.12, saturation: 0.8, hue: 0, opacity: 1 },
  { hour: 17, brightness: 1.05, saturation: 1.7, hue: -18, opacity: 0.92 },
  { hour: 20, brightness: 0.78, saturation: 2.5, hue: -32, opacity: 0.65 },
]

/** @param {Date} date */
export function getCelestialAppearance(date) {
  const hour = date.getHours() + date.getMinutes() / 60
  if (hour < 5 || hour >= 20) {
    return { sepia: 0, brightness: 1.05, saturation: 0.65, hue: 0, opacity: 0.95 }
  }

  const endIndex = SUN_LIGHT_STOPS.findIndex((stop) => stop.hour > hour)
  const start = SUN_LIGHT_STOPS[endIndex - 1]
  const end = SUN_LIGHT_STOPS[endIndex]
  const progress = (hour - start.hour) / (end.hour - start.hour)
  const blend = (key) => Number((start[key] + (end[key] - start[key]) * progress).toFixed(4))

  return {
    sepia: 0.15,
    brightness: blend('brightness'),
    saturation: blend('saturation'),
    hue: blend('hue'),
    opacity: blend('opacity'),
  }
}

/** @param {Date} date */
export function getCelestialPosition(date) {
  const decimalHour = date.getHours() + date.getMinutes() / 60
  const isNight = decimalHour < 5 || decimalHour >= 20
  const normalizedHour = isNight && decimalHour < 5 ? decimalHour + 24 : decimalHour
  const start = isNight ? 20 : 5
  const duration = isNight ? 9 : 15
  const progress = clamp((normalizedHour - start) / duration, 0, 1)
  // Keep the sun above the riverbank instead of descending into the water layer.
  const baseline = isNight ? 54 : 22
  const arcHeight = isNight ? 42 : 12

  return {
    x: Number((58 + progress * 32).toFixed(2)),
    y: Number((baseline - Math.sin(progress * Math.PI) * arcHeight).toFixed(2)),
    isNight,
  }
}

/** @param {number} hour */
function getPeriodLabel(hour) {
  if (hour < 5) return '凌晨'
  if (hour < 9) return '早上'
  if (hour < 12) return '上午'
  if (hour < 18) return '下午'
  return '晚上'
}

/** @param {Date} [date] */
export function getTimeScene(date = new Date()) {
  const hour = date.getHours()
  const key = hour >= 5 && hour < 8
    ? 'dawn'
    : hour >= 8 && hour < 17
      ? 'day'
      : hour >= 17 && hour < 20
        ? 'dusk'
        : 'night'

  return {
    ...SCENES[key],
    dateLabel: `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`,
    timeLabel: `${pad(hour)}:${pad(date.getMinutes())}`,
    periodLabel: getPeriodLabel(hour),
    celestial: getCelestialPosition(date),
    celestialAppearance: getCelestialAppearance(date),
  }
}
