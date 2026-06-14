import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const taskMonitor = readFileSync(new URL('./TaskMonitor.vue', import.meta.url), 'utf8')
const loadBlock = taskMonitor.match(/async function load[\s\S]*?\n}\n\nfunction toNumber/)?.[0] || ''

describe('task monitor page', () => {
  it('skips overlapping auto refreshes so the loading state can settle', () => {
    expect(loadBlock).toContain('skipIfLoading')
    expect(loadBlock).toContain('if (skipIfLoading && loading.value)')
    expect(loadBlock).toContain('return')
    expect(taskMonitor).toContain('setInterval(() => load({ skipIfLoading: true }), 5000)')
  })
})
