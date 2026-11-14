// 秒传命中已有视频后的合集追加反馈文案。
// 计数口径是后端实际新增的关系条数，不是本次选择的数量。
export function buildExistingUploadMessage(result) {
  if (result?.pending_delete) {
    return '已存在（该视频待删除），未关联合集'
  }
  const added = Number(result?.added_collections)
  const addedCollections = Number.isFinite(added) ? added : 0
  const addedImageCollection = result?.added_image_collection === true
  if (addedCollections > 0) {
    return addedImageCollection
      ? `已存在，新增 ${addedCollections} 个合集，并补齐图片合集`
      : `已存在，新增 ${addedCollections} 个合集`
  }
  if (addedImageCollection) {
    return '已存在，已补齐图片合集'
  }
  return '已存在，合集无变化'
}

export function hasExistingAttachmentSelection(collectionIDs, imageCollectionID) {
  const collections = Array.isArray(collectionIDs) ? collectionIDs.filter((id) => String(id || '').trim() !== '') : []
  return collections.length > 0 || String(imageCollectionID || '').trim() !== ''
}
