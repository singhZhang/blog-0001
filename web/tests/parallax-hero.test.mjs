import assert from 'node:assert/strict'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { createSSRApp } from 'vue'
import { renderToString } from '@vue/server-renderer'
import vue from '@vitejs/plugin-vue'
import { createServer } from 'vite'

test('renders the current time inside the hero eyebrow without a separate time card', async (t) => {
  const server = await createServer({
    root: fileURLToPath(new URL('..', import.meta.url)),
    appType: 'custom',
    cacheDir: '/tmp/blog-0001-vite-test',
    configFile: false,
    logLevel: 'silent',
    plugins: [vue()],
    server: { middlewareMode: true },
  })
  t.after(() => server.close())

  const { default: ParallaxHero } = await server.ssrLoadModule(
    '/src/components/ParallaxHero.vue',
  )
  const html = await renderToString(createSSRApp(ParallaxHero, {
    scene: {
      key: 'day',
      name: '白昼',
      dateLabel: '2026年10月27日',
      timeLabel: '14:27',
      periodLabel: '下午',
    },
    sceneStyle: {},
  }))

  assert.match(
    html,
    /class="eyebrow"[^>]*>[\s\S]*?<time class="eyebrow-time"[^>]*>2026年10月27日 14:27 · 下午<\/time>[\s\S]*?<\/p>/,
  )
  assert.doesNotMatch(html, /class="local-time"/)
  assert.match(html, /<h1[^>]*>记录日常，也沉淀思考。<\/h1>/)
  assert.doesNotMatch(html, /开始阅读/)
})
