import { describe, expect, it } from 'vitest'
import {
  canAdvanceStep,
  getNextStep,
  getPrevStep,
  shouldShowAVFields,
  shouldShowCollectionField
} from './videoUpload.wizard.helpers'

describe('videoUpload wizard helpers', () => {
  it('blocks step 0 when no file is selected', () => {
    expect(canAdvanceStep(0, { files: [] })).toBe(false)
  })

  it('allows step 0 when at least one file is selected', () => {
    expect(canAdvanceStep(0, { files: [{ name: 'demo.mp4' }] })).toBe(true)
  })

  it('blocks step 1 when type is empty', () => {
    expect(canAdvanceStep(1, { type: '' })).toBe(false)
  })

  it('allows step 1 when type is selected', () => {
    expect(canAdvanceStep(1, { type: 'short' })).toBe(true)
  })

  it('does not allow advancing from last step', () => {
    expect(canAdvanceStep(2, { files: [{ name: 'demo.mp4' }], type: 'short' })).toBe(false)
  })

  it('moves from step 0 to step 1', () => {
    expect(getNextStep(0)).toBe(1)
  })

  it('moves from step 1 to step 2', () => {
    expect(getNextStep(1)).toBe(2)
  })

  it('clamps next step at the last step', () => {
    expect(getNextStep(2)).toBe(2)
  })

  it('clamps previous step at the first step', () => {
    expect(getPrevStep(0)).toBe(0)
  })

  it('moves from step 1 to step 0', () => {
    expect(getPrevStep(1)).toBe(0)
  })

  it('moves from step 2 to step 1', () => {
    expect(getPrevStep(2)).toBe(1)
  })

  it('shows AV fields only for AV uploads', () => {
    expect(shouldShowAVFields('av')).toBe(true)
    expect(shouldShowAVFields('movie')).toBe(false)
  })

  it('shows collection field only for short videos', () => {
    expect(shouldShowCollectionField('short')).toBe(true)
    expect(shouldShowCollectionField('movie')).toBe(false)
  })
})
