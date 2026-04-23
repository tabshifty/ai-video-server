import { describe, expect, it } from 'vitest'
import {
  buildTvEpisodePayload,
  buildTvSeasonPayload,
  buildTvSeriesPayload,
  mapEpisodeVideoOption
} from './tvSeriesManage.helpers'

describe('tvSeriesManage helpers', () => {
  it('buildTvSeriesPayload trims text fields', () => {
    expect(
      buildTvSeriesPayload({
        title: '  雾城档案  ',
        overview: '  调查悬案  ',
        poster_url: '  /poster.jpg  ',
        backdrop_url: '  /backdrop.jpg  ',
        first_air_date: '2025-04-01',
        active: true
      })
    ).toEqual({
      title: '雾城档案',
      overview: '调查悬案',
      poster_url: '/poster.jpg',
      backdrop_url: '/backdrop.jpg',
      first_air_date: '2025-04-01',
      active: true
    })
  })

  it('buildTvSeasonPayload normalizes season numbers', () => {
    expect(
      buildTvSeasonPayload({
        season_number: '2',
        title: ' 第二季 ',
        overview: '  ',
        poster_url: '',
        air_date: ''
      })
    ).toEqual({
      season_number: 2,
      title: '第二季',
      overview: '',
      poster_url: '',
      air_date: ''
    })
  })

  it('buildTvEpisodePayload keeps blank video id as empty string', () => {
    expect(
      buildTvEpisodePayload({
        episode_number: '6',
        title: ' 第六集 ',
        overview: ' 逆转 ',
        runtime: '45',
        air_date: '2025-04-23',
        still_url: ' /still.jpg ',
        video_id: ' '
      })
    ).toEqual({
      episode_number: 6,
      title: '第六集',
      overview: '逆转',
      runtime: 45,
      air_date: '2025-04-23',
      still_url: '/still.jpg',
      video_id: ''
    })
  })

  it('mapEpisodeVideoOption formats title and status text', () => {
    expect(
      mapEpisodeVideoOption({
        id: 'video-1',
        title: '雾城档案 S02E06',
        status: 'ready'
      })
    ).toEqual({
      value: 'video-1',
      label: '雾城档案 S02E06 · 可播放'
    })
  })
})
