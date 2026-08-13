import { mkdtemp, mkdir, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { validateSnapshot } from './validate.mjs'

const temporaryDirectories = []

async function createSnapshot() {
  const directory = await mkdtemp(join(tmpdir(), 'china-map-validate-'))
  temporaryDirectories.push(directory)
  await mkdir(join(directory, 'layers'))
  await writeFile(
    join(directory, 'manifest.json'),
    JSON.stringify({
      status: 'complete',
      snapshot_timestamp: '2026-08-12T00:00:00Z',
      geometry_source: 'OpenStreetMap',
      license: 'ODbL 1.0',
      files: []
    })
  )
  await writeFile(
    join(directory, 'catalog.json'),
    JSON.stringify({
      root_region_code: 'CN',
      regions: [
        {
          region_code: 'CN',
          name: '中国',
          parent_region_code: null,
          child_region_codes: []
        }
      ]
    })
  )
  return directory
}

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map((directory) => rm(directory, { recursive: true })))
})

describe('中国地图快照完整性门禁', () => {
  it('拒绝缺少匹配报告、完整省级范围和清单覆盖的快照', async () => {
    const directory = await createSnapshot()
    const result = await validateSnapshot(directory)

    expect(result.issues.some((issue) => issue.includes('匹配报告'))).toBe(true)
    expect(result.issues.some((issue) => issue.includes('省级范围缺失'))).toBe(true)
    expect(result.issues.some((issue) => issue.includes('清单未覆盖'))).toBe(true)
  })
})
