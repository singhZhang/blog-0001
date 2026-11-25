import assert from 'node:assert/strict'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { createSSRApp } from 'vue'
import { renderToString } from '@vue/server-renderer'
import vue from '@vitejs/plugin-vue'
import { createServer } from 'vite'

test('scroll controls render quick links to the top and bottom of the page', async (t) => {
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

  let html = ''
  try {
    const { default: ScrollControls } = await server.ssrLoadModule(
      '/src/components/ScrollControls.vue',
    )
    html = await renderToString(createSSRApp(ScrollControls))
  } catch (reason) {
    if (!String(reason).includes('ScrollControls.vue')) throw reason
  }

  assert.match(html, /href="#top"[^>]*aria-label="一键到顶"/)
  assert.match(html, /href="#page-end"[^>]*aria-label="一键到底"/)
})
