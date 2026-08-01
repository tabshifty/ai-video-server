// 待上传视频标题清洗规则：从文件名生成标题。
// 规则（顺序固定）：
// 1. 去掉末尾常见视频扩展名（大小写不敏感）；
// 2. 全局替换水印文字 `www.98t.la@`（大小写不敏感）；
// 3. 收尾清理：连续空白压成单个空格、连续 `.`/`-`/`_` 压成单个 `.`、去掉首尾空白与点号；
// 4. 兜底：清洗结果为空时回退为去掉扩展名的原始文件名。

export const VIDEO_EXTENSIONS = [
  'mp4',
  'mkv',
  'avi',
  'ts',
  'mov',
  'wmv',
  'flv',
  'webm',
  'm4v',
  'rmvb',
  'rm',
  '3gp',
  'mpg',
  'mpeg',
  'm2ts',
  'mts',
  'ogv',
  'asf',
  'f4v',
  'vob'
]

export const WATERMARK_PATTERN = /www\.98t\.la@/gi

export function stripVideoExtension(filename) {
  const text = String(filename || '')
  const match = text.match(/^(.*)\.([^.\s]+)$/)
  if (!match) return text
  const extension = match[2].toLowerCase()
  return VIDEO_EXTENSIONS.includes(extension) ? match[1] : text
}

export function buildTitleFromFilename(filename) {
  const withoutExt = stripVideoExtension(filename)
  const dewatermarked = withoutExt.replace(WATERMARK_PATTERN, '')
  const cleaned = dewatermarked
    .replace(/\s+/g, ' ')
    .replace(/[._-]+/g, '.')
    .replace(/^[.\s]+|[.\s]+$/g, '')
  return cleaned || withoutExt
}
