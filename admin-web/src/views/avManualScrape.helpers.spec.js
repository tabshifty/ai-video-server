import { describe, expect, it } from 'vitest'

import {
  AV_POSTER_CROP_MODE_OPTIONS,
  AV_SITE_OPTIONS,
  applyAVScrapeConfig,
  buildAVManualScrapeConfirmPayload,
  buildAVManualScrapePreviewPayload,
  buildAVScrapeConfigPayload
} from './avManualScrape.helpers'

describe('avManualScrape helpers', () => {
  it('exposes all MDCx migrated AV sites', () => {
    expect(AV_SITE_OPTIONS).toEqual(expect.arrayContaining([
      'theporndb',
      'dmm',
      'javdb',
      'jav321',
      'mgstage',
      'fc2ppvdb',
      'fc2club',
      'fc2',
      'fc2hub'
    ]))
  })

  it('builds preview payload with cache bypass and site hints', () => {
    expect(buildAVManualScrapePreviewPayload({
      video_id: 'video-1',
      title: 'SSIS-123',
      site_category: 'japanese',
      site_source: 'javlibrary',
      bypass_cache: true
    })).toEqual({
      video_id: 'video-1',
      title: 'SSIS-123',
      site_category: 'japanese',
      site_source: 'javlibrary',
      bypass_cache: true
    })
  })

  it('builds confirm payload with explicit site routing', () => {
    expect(buildAVManualScrapeConfirmPayload(
      { video_id: 'video-1', site_category: 'fc2', site_source: 'fc2club' },
      { external_id: 'abc123', title: 'FC2-123456', metadata: null }
    )).toEqual({
      video_id: 'video-1',
      external_id: 'abc123',
      title: 'FC2-123456',
      overview: '',
      poster_url: '',
      release_date: '',
      site_category: 'fc2',
      site_source: 'fc2club',
      metadata: {}
    })
  })

  it('builds config payload from comma separated site order', () => {
    expect(buildAVScrapeConfigPayload({
      enabled_sites: ['javdb', 'theporndb'],
      fc2_order: 'fc2, fc2club',
      western_order: 'theporndb, javdb',
      japanese_order: 'javdb, javlibrary',
      poster_crop_enabled: false,
      poster_crop_mode: 'portrait_center'
    })).toEqual({
      enabled_sites: ['javdb', 'theporndb'],
      category_site_order: {
        fc2: ['fc2', 'fc2club'],
        western: ['theporndb', 'javdb'],
        japanese: ['javdb', 'javlibrary']
      },
      poster_crop_enabled: false,
      poster_crop_mode: 'portrait_center'
    })
  })

  it('only accepts the three allowed poster crop modes in config payload', () => {
    expect(AV_POSTER_CROP_MODE_OPTIONS).toEqual([
      'portrait_center',
      'portrait_left',
      'portrait_right'
    ])

    expect(buildAVScrapeConfigPayload({
      poster_crop_mode: 'portrait_left'
    }).poster_crop_mode).toBe('portrait_left')

    expect(buildAVScrapeConfigPayload({
      poster_crop_mode: 'portrait_right'
    }).poster_crop_mode).toBe('portrait_right')

    expect(buildAVScrapeConfigPayload({
      poster_crop_mode: 'unexpected-mode'
    }).poster_crop_mode).toBe('portrait_center')
  })

  it('applies remote config into editable form state', () => {
    const form = {}
    applyAVScrapeConfig(form, {
      enabled_sites: ['javdb', 'fc2'],
      category_site_order: {
        fc2: ['fc2', 'fc2club'],
        western: ['theporndb'],
        japanese: ['javdb', 'javlibrary']
      },
      poster_crop_enabled: true,
      poster_crop_mode: 'portrait_center'
    })

    expect(form).toEqual({
      enabled_sites: ['javdb', 'fc2'],
      fc2_order: 'fc2, fc2club',
      western_order: 'theporndb',
      japanese_order: 'javdb, javlibrary',
      poster_crop_enabled: true,
      poster_crop_mode: 'portrait_center'
    })
  })

  it('falls back to portrait_center when remote crop mode is empty or invalid', () => {
    const form = {}
    applyAVScrapeConfig(form, {
      poster_crop_mode: 'unexpected-mode'
    })
    expect(form.poster_crop_mode).toBe('portrait_center')

    applyAVScrapeConfig(form, {
      poster_crop_mode: 'portrait_right'
    })
    expect(form.poster_crop_mode).toBe('portrait_right')

    applyAVScrapeConfig(form, {})
    expect(form.poster_crop_mode).toBe('portrait_center')
  })
})
