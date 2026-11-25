import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const headerSource = await readFile(
  new URL('../src/components/SiteHeader.vue', import.meta.url),
  'utf8',
)

test('site header renders the generated logo with an accessible brand label', () => {
  assert.match(headerSource, /src="\/brand\/logo-09494\.png"/)
  assert.match(headerSource, /alt="零玖肆玖肆"/)
  assert.match(headerSource, /width="300"\s+height="64"/)
  assert.match(headerSource, /aria-label="零玖肆玖肆首页"/)
})
