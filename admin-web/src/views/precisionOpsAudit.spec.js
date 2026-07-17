import { readFileSync, readdirSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const directory = new URL('.', import.meta.url)
const viewFiles = readdirSync(directory).filter((name) => name.endsWith('.vue'))
const themeSource = readFileSync(new URL('../assets/theme.css', directory), 'utf8')
const indexSource = readFileSync(new URL('../../index.html', directory), 'utf8')
const MASK_DATA_COLOR_DECLARATION = "const MASK_OPAQUE_COLOR = '#ffffff'"
const ZERO_LETTER_SPACING_PATTERN = /^[+-]?(?:0+(?:\.0*)?|\.0+)(?:px|rem|em)?$/i
const REMOTE_FONT_DOMAIN_PATTERN = /fonts\.(?:googleapis|gstatic)\.com/i
const DIRECT_VIEW_COLOR_PATTERNS = [
  /#(?:[0-9a-f]{3,8})\b/i,
  /rgba?\(\s*[+-]?(?:\d|\.\d)/i,
  /\b(?:hsla?|oklch|oklab|lab|lch|color)\(/i
]
const DECORATIVE_GRADIENT_PATTERN = /(?:linear|radial|conic)-gradient\(/i
const ABSOLUTE_LENGTH_TO_PX = {
  px: 1,
  rem: 16,
  em: 16,
  pt: 96 / 72,
  pc: 16,
  in: 96,
  cm: 96 / 2.54,
  mm: 96 / 25.4,
  q: 96 / 101.6
}
const functionalGradientBlocks = {
  'ImageManage.vue': /\.image-grid-card__preview\s*\{[^{}]*\}/g,
  'ImageWorkbenchMaskEditor.vue': /\.mask-editor__stage\s*\{[^{}]*\}/g
}

function colorAuditSource(file, source) {
  return file === 'ImageWorkbenchMaskEditor.vue'
    ? source.replace(MASK_DATA_COLOR_DECLARATION, '')
    : source
}

function gradientAuditSource(file, source) {
  const block = functionalGradientBlocks[file]
  return block ? source.replace(block, '') : source
}

function nonzeroLetterSpacingDeclarations(source) {
  return [...source.matchAll(/letter-spacing:\s*([^;]+);/gi)]
    .map((match) => match[1].trim())
    .filter((value) => !ZERO_LETTER_SPACING_PATTERN.test(value))
}

function directViewColorDeclarations(source) {
  return DIRECT_VIEW_COLOR_PATTERNS
    .map((pattern) => source.match(pattern)?.[0])
    .filter(Boolean)
}

function decorativeGradientDeclarations(source) {
  return source.match(DECORATIVE_GRADIENT_PATTERN) || []
}

function openingTags(source, elementName) {
  return source.match(new RegExp(`<${elementName}\\b[^>]*>`, 'gi')) || []
}

function viewTagViolations(elementName, predicate) {
  return viewFiles.flatMap((file) => {
    const source = readFileSync(new URL(file, directory), 'utf8')
    return openingTags(source, elementName)
      .filter(predicate)
      .map((tag) => `${file}: ${tag.replace(/\\s+/g, ' ')}`)
  })
}

function cssAuditSource(source) {
  const styles = [...source.matchAll(/<style\b[^>]*>([\s\S]*?)<\/style>/gi)]
    .map((match) => match[1])
  return styles.length > 0 ? styles.join('\n') : source
}

function cssRules(source) {
  return [...cssAuditSource(source).matchAll(/([^{}]+)\{([^{}]*)\}/g)]
    .map((match) => ({ selector: match[1].trim(), declarations: match[2] }))
}

function declarationValues(declarations, property) {
  const pattern = new RegExp(`(?:^|;)\\s*${property}\\s*:\\s*([^;]+)`, 'gi')
  return [...declarations.matchAll(pattern)].map((match) => match[1].trim())
}

function selectorBranches(selector) {
  return selector.split(',').map((branch) => branch.trim()).filter(Boolean)
}

function isTransientSelector(selector) {
  return /:(hover|focus|focus-visible|focus-within|active)\b/i.test(selector)
    || /\.(?:is-)?(?:active|selected)\b/i.test(selector)
}

function imageActionVisibilityViolations(source) {
  const actionRules = cssRules(source).flatMap((rule) => selectorBranches(rule.selector)
    .filter((selector) => selector.includes('.image-grid-card__actions'))
    .map((selector) => ({ selector, declarations: rule.declarations })))
  const baseRules = actionRules.filter(({ selector }) => !isTransientSelector(selector))
  const violations = baseRules.flatMap(({ selector, declarations }) => {
    const hiddenDisplay = declarationValues(declarations, 'display').filter((value) => /^none$/i.test(value))
    const hiddenVisibility = declarationValues(declarations, 'visibility').filter((value) => /^(?:hidden|collapse)$/i.test(value))
    const hiddenOpacity = declarationValues(declarations, 'opacity').filter((value) => {
      const numeric = /^[+-]?(?:\d+(?:\.\d*)?|\.\d+)$/.test(value) ? Number(value) : Number.NaN
      return !Number.isFinite(numeric) || numeric <= 0
    })
    return [...hiddenDisplay, ...hiddenVisibility, ...hiddenOpacity].map((value) => `${selector}: ${value}`)
  })
  return baseRules.length > 0 ? violations : ['missing visible base action rule']
}

function persistentPanelShadowSelectors(source) {
  return cssRules(source).flatMap((rule) => {
    const shadows = declarationValues(rule.declarations, 'box-shadow')
      .filter((value) => !/^none$/i.test(value) && !/^inset\b/i.test(value))
    if (shadows.length === 0) return []
    return selectorBranches(rule.selector)
      .filter((selector) => !isTransientSelector(selector))
      .filter((selector) => !/(drawer|dialog|popover|popper|tooltip|dropdown|bulk-action|floating|overlay|modal)/i.test(selector))
  })
}

function oversizedDirectRadii(source) {
  return [...cssAuditSource(source).matchAll(/border-radius:\s*([^;]+);/gi)]
    .map((match) => match[1].trim())
    .filter((value) => {
      if (/^var\(/i.test(value)) return false
      return value.split(/[\s/]+/).some((part) => {
        const match = part.match(/^([+-]?(?:\d+(?:\.\d*)?|\.\d+))([a-z%]*)$/i)
        if (!match) return true
        const numeric = Number(match[1])
        const unit = match[2].toLowerCase()
        if (numeric === 0) return false
        if (numeric < 0 || !Object.hasOwn(ABSOLUTE_LENGTH_TO_PX, unit)) return true
        return numeric * ABSOLUTE_LENGTH_TO_PX[unit] > 8
      })
    })
}

describe('Product runtime hygiene contracts', () => {
  it('uses local Chinese and system font fallbacks without remote font loading', () => {
    const runtimeSources = `${themeSource}\n${indexSource}`
    const fontSans = themeSource.match(/--font-sans:\s*([^;]+);/i)?.[1].trim()
    const remoteFontImports = (themeSource.match(/@import[^;]*;/gi) || [])
      .filter((declaration) => REMOTE_FONT_DOMAIN_PATTERN.test(declaration))
    const remoteFontPreconnects = openingTags(indexSource, 'link')
      .filter((tag) => /(?:^|\s)rel=['"]preconnect['"]/i.test(tag))
      .filter((tag) => REMOTE_FONT_DOMAIN_PATTERN.test(tag))

    expect(runtimeSources).not.toMatch(REMOTE_FONT_DOMAIN_PATTERN)
    expect(remoteFontImports).toEqual([])
    expect(remoteFontPreconnects).toEqual([])
    expect(fontSans).toBe("'PingFang SC', 'Microsoft YaHei', system-ui, -apple-system, sans-serif")
  })

  it('binds numeric el-input rows as numbers in every view', () => {
    const violations = viewTagViolations('el-input', (tag) => (
      /(?:^|\s)rows\s*=\s*(['"])\d+\1/i.test(tag)
    ))

    expect(violations).toEqual([])
  })

  it('uses value for every el-radio-button option value', () => {
    const violations = viewTagViolations('el-radio-button', (tag) => (
      /(?:^|\s):?label\s*=/i.test(tag)
    ))

    expect(violations).toEqual([])
  })
})

describe('Precision Ops final audit', () => {
  viewFiles.forEach((file) => {
    const source = readFileSync(new URL(file, directory), 'utf8')
    it(`${file} avoids nonzero tracking and direct view colors`, () => {
      const auditedSource = colorAuditSource(file, source)
      expect(nonzeroLetterSpacingDeclarations(source)).toEqual([])
      expect(directViewColorDeclarations(auditedSource)).toEqual([])
    })
    it(`${file} avoids persistent panel shadows`, () => {
      expect(persistentPanelShadowSelectors(source)).toEqual([])
    })
    it(`${file} avoids oversized direct panel radii`, () => {
      expect(oversizedDirectRadii(source)).toEqual([])
    })
    it(`${file} avoids decorative gradients`, () => {
      expect(decorativeGradientDeclarations(gradientAuditSource(file, source))).toEqual([])
    })
  })

  it('keeps image card actions visible without hover dependency', () => {
    const imageManage = readFileSync(new URL('./ImageManage.vue', import.meta.url), 'utf8')
    expect(imageActionVisibilityViolations(imageManage)).toEqual([])
  })

  it('keeps shell letter spacing at zero', () => {
    const layout = readFileSync(new URL('../components/Layout.vue', import.meta.url), 'utf8')
    expect(nonzeroLetterSpacingDeclarations(layout)).toEqual([])
  })
})

describe('Precision Ops audit helper contracts', () => {
  it.each([
    'letter-spacing: 0;',
    'letter-spacing: 0.0;',
    'letter-spacing: -0;',
    'letter-spacing: 0.00px;',
    'letter-spacing: -0.0rem;',
    'letter-spacing: 0em;'
  ])('accepts numeric zero letter spacing: %s', (source) => {
    expect(nonzeroLetterSpacingDeclarations(source)).toEqual([])
  })

  it.each([
    'letter-spacing: 0.01em;',
    'letter-spacing: -1px;',
    'letter-spacing: normal;',
    'letter-spacing: var(--tracking);',
    'letter-spacing: calc(0px);'
  ])('rejects non-literal-zero letter spacing: %s', (source) => {
    expect(nonzeroLetterSpacingDeclarations(source)).not.toEqual([])
  })

  it.each([
    'color: var(--text-primary);',
    'background: color-mix(in srgb, var(--primary) 20%, transparent);'
  ])('accepts tokenized view colors: %s', (source) => {
    expect(directViewColorDeclarations(source)).toEqual([])
  })

  it.each([
    'color: hsl(220 10% 20%);',
    'color: HSLA(220 10% 20% / 80%);',
    'color: oklch(60% 0.15 240);',
    'color: color(display-p3 1 0 0);',
    'color: RGB(10 20 30);'
  ])('rejects direct functional view colors: %s', (source) => {
    expect(directViewColorDeclarations(source)).not.toEqual([])
  })

  it.each([
    '.panel:hover { box-shadow: 0 4px 12px var(--shadow-color); }',
    '.dialog { box-shadow: 0 4px 12px var(--shadow-color); }',
    '.selection-control { box-shadow: inset 0 0 0 1px var(--line-soft); }',
    '.panel { box-shadow: none; }'
  ])('accepts transient or functional shadows: %s', (source) => {
    expect(persistentPanelShadowSelectors(source)).toEqual([])
  })

  it.each([
    '.panel { box-shadow: 0 4px 12px rgb(0 0 0 / 20%); }',
    '.panel, .panel:hover { box-shadow: var(--shadow-sm); }',
    '.panel, .tooltip { box-shadow: var(--shadow-sm); }'
  ])('rejects persistent panel shadows without group-wide exemptions: %s', (source) => {
    expect(persistentPanelShadowSelectors(source)).not.toEqual([])
  })

  it.each([
    'border-radius: 8px;',
    'border-radius: 0.5rem;',
    'border-radius: 6pt;',
    'border-radius: 0;'
  ])('accepts direct radii no larger than 8px: %s', (source) => {
    expect(oversizedDirectRadii(source)).toEqual([])
  })

  it.each([
    'border-radius: 9px;',
    'border-radius: 0.75rem;',
    'border-radius: 0.6em;',
    'border-radius: 7pt;',
    'border-radius: 1vw;'
  ])('rejects direct radii that exceed or can bypass 8px: %s', (source) => {
    expect(oversizedDirectRadii(source)).not.toEqual([])
  })

  it.each([
    'background: var(--bg-surface);',
    'background: color-mix(in srgb, var(--bg-surface) 80%, transparent);'
  ])('accepts non-gradient backgrounds: %s', (source) => {
    expect(decorativeGradientDeclarations(source)).toEqual([])
  })

  it.each([
    'background: LINEAR-GRADIENT(red, blue);',
    'background: Radial-Gradient(red, blue);',
    'background: conic-gradient(red, blue);'
  ])('rejects decorative gradient variants: %s', (source) => {
    expect(decorativeGradientDeclarations(source)).not.toEqual([])
  })

  it('accepts an explicitly visible base image action rule', () => {
    const source = '.image-grid-card__actions { display: flex; visibility: visible; opacity: 1; }'
    expect(imageActionVisibilityViolations(source)).toEqual([])
  })

  it.each([
    '.image-grid-card__actions { opacity: 0; }',
    '.image-grid-card__actions { opacity: var(--action-opacity); }',
    '.image-grid-card__actions { visibility: hidden; }',
    '.image-grid-card__actions { display: none; }',
    '.image-grid-card:hover .image-grid-card__actions { opacity: 1; }'
  ])('rejects image actions hidden or available only on hover: %s', (source) => {
    expect(imageActionVisibilityViolations(source)).not.toEqual([])
  })
})
