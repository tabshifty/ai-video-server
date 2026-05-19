import { describe, expect, it } from 'vitest'

import {
  buildScrapeConfirmPayload,
  buildScrapePreviewPayload,
  resolveTVPreviewState
} from './scrapePreview.helpers'

describe('scrapePreview helpers', () => {
  it('adds season and episode to tv preview payloads', () => {
    expect(buildScrapePreviewPayload({
      video_id: 'video-1',
      title: '庆余年第一季第一集',
      type: 'tv',
      year: null,
      season_number: '1',
      episode_number: 2
    })).toEqual({
      video_id: 'video-1',
      title: '庆余年第一季第一集',
      type: 'tv',
      season_number: 1,
      episode_number: 2
    })
  })

  it('keeps season and episode out of non-tv confirm payloads', () => {
    expect(buildScrapeConfirmPayload(
      { type: 'movie', season_number: 1, episode_number: 2 },
      { video_id: 'video-1', tmdb_id: 10, title: '盗梦空间', metadata: null }
    )).toEqual({
      video_id: 'video-1',
      tmdb_id: 10,
      external_id: '',
      title: '盗梦空间',
      overview: '',
      poster_url: '',
      backdrop_url: '',
      release_date: '',
      metadata: {}
    })
  })

  it('includes movie backdrop url in confirm payloads', () => {
    expect(buildScrapeConfirmPayload(
      { type: 'movie' },
      {
        video_id: 'video-1',
        tmdb_id: 10,
        title: '盗梦空间',
        poster_url: ' /poster.jpg ',
        backdrop_url: ' /backdrop.jpg ',
        metadata: {}
      }
    )).toMatchObject({
      poster_url: '/poster.jpg',
      backdrop_url: '/backdrop.jpg'
    })
  })

  it('resolves parsed tv preview state from candidate first', () => {
    expect(resolveTVPreviewState(
      {
        parsed_title: '庆余年',
        parsed_season_number: 1,
        parsed_episode_number: 1
      },
      {
        title: '庆余年第一季第一集',
        season_number: 3,
        episode_number: 8
      }
    )).toEqual({
      title: '庆余年',
      season_number: 1,
      episode_number: 1
    })
  })
})
