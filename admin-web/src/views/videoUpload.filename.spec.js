import { describe, expect, it } from 'vitest'

import { buildTitleFromFilename, stripVideoExtension } from './videoUpload.filename'

describe('stripVideoExtension', () => {
  it('去掉常见视频扩展名（大小写不敏感）', () => {
    expect(stripVideoExtension('ABC.www.98T.la@DEF.mp4')).toBe('ABC.www.98T.la@DEF')
    expect(stripVideoExtension('Movie.2024.05.12.MKV')).toBe('Movie.2024.05.12')
    expect(stripVideoExtension('ep01-1080p.ts')).toBe('ep01-1080p')
    expect(stripVideoExtension('clip.m4v')).toBe('clip')
  })

  it('非视频扩展名保留原样', () => {
    expect(stripVideoExtension('archive.zip')).toBe('archive.zip')
    expect(stripVideoExtension('notes.txt')).toBe('notes.txt')
    expect(stripVideoExtension('noext')).toBe('noext')
  })

  it('空值与无点文件名原样返回', () => {
    expect(stripVideoExtension('')).toBe('')
    expect(stripVideoExtension(null)).toBe('')
    expect(stripVideoExtension('plain')).toBe('plain')
  })
})

describe('buildTitleFromFilename', () => {
  it('按确认规则清洗：去扩展名、去水印、合并分隔符', () => {
    expect(buildTitleFromFilename('ABC.www.98T.la@DEF.mp4')).toBe('ABC.DEF')
    expect(buildTitleFromFilename('ABC www.98T.la@ DEF.mp4')).toBe('ABC DEF')
    expect(buildTitleFromFilename('ABC_www.98T.la@_DEF.mp4')).toBe('ABC.DEF')
  })

  it('水印大小写不敏感且全局替换', () => {
    expect(buildTitleFromFilename('a.WWW.98T.LA@b.mp4')).toBe('a.b')
    expect(buildTitleFromFilename('www.98t.la@X.www.98t.la@Y.mkv')).toBe('X.Y')
  })

  it('非视频扩展名与多段点号保留', () => {
    expect(buildTitleFromFilename('Movie.2024.05.12.mkv')).toBe('Movie.2024.05.12')
    expect(buildTitleFromFilename('archive.zip')).toBe('archive.zip')
  })

  it('清理首尾点号与空白', () => {
    expect(buildTitleFromFilename('.www.98T.la@ABC.mp4')).toBe('ABC')
    expect(buildTitleFromFilename('  ABC.www.98T.la@DEF  .mkv')).toBe('ABC.DEF')
  })

  it('清洗结果为空时回退为去扩展名的原始文件名', () => {
    expect(buildTitleFromFilename('www.98T.la@.mp4')).toBe('www.98T.la@')
  })

  it('空值返回空串', () => {
    expect(buildTitleFromFilename('')).toBe('')
    expect(buildTitleFromFilename(null)).toBe('')
  })
})
