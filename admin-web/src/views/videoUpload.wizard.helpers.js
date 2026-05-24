export const WIZARD_STEPS = ['file', 'basic', 'relate']

export function canAdvanceStep(currentIndex, form) {
  if (currentIndex === 0) {
    return Array.isArray(form?.files) && form.files.length > 0
  }
  if (currentIndex === 1) {
    return typeof form?.type === 'string' && form.type.trim() !== ''
  }
  return false
}

export function getNextStep(currentIndex) {
  return Math.min(currentIndex + 1, WIZARD_STEPS.length - 1)
}

export function getPrevStep(currentIndex) {
  return Math.max(currentIndex - 1, 0)
}

export function shouldShowAVFields(type) {
  return type === 'av'
}

export function shouldShowCollectionField(type) {
  return type === 'short'
}
