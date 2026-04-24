import { describe, expect, it } from 'vitest'
import { extractTvPendingDiagnostics, getVideoStatusMeta } from './videoList.helpers'

describe('videoList helpers', () => {
  it('includes tv_pending status label and tag type', () => {
    expect(getVideoStatusMeta('tv_pending')).toEqual({
      label: '待绑定',
      tagType: 'warning'
    })
  })

  it('extracts tv pending diagnostics from metadata', () => {
    const diagnostics = extractTvPendingDiagnostics({
      scrape_error: 'ambiguous tv candidate',
      scrape_stage: 'candidate_ambiguous',
      parsed_title: '三体',
      parsed_season_number: 1,
      parsed_episode_number: 2,
      candidate_count: 2,
      candidate_preview: [{ tmdb_id: 1, title: '三体' }]
    })

    expect(diagnostics).toEqual({
      error: 'ambiguous tv candidate',
      stage: 'candidate_ambiguous',
      parsedTitle: '三体',
      parsedSeasonNumber: 1,
      parsedEpisodeNumber: 2,
      candidateCount: 2,
      candidatePreview: [{ tmdb_id: 1, title: '三体' }]
    })
  })
})
