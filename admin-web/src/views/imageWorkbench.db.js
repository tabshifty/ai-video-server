const DB_NAME = 'admin-image-workbench'
const DB_VERSION = 1
const STORE_TASKS = 'tasks'
const STORE_IMAGES = 'images'
const STORE_THUMBNAILS = 'thumbnails'
const THUMBNAIL_MAX_SIZE = 480
const THUMBNAIL_QUALITY = 0.82

function openImageWorkbenchDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(STORE_TASKS)) {
        db.createObjectStore(STORE_TASKS, { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains(STORE_IMAGES)) {
        db.createObjectStore(STORE_IMAGES, { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains(STORE_THUMBNAILS)) {
        db.createObjectStore(STORE_THUMBNAILS, { keyPath: 'id' })
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

function transaction(storeName, mode, run) {
  return openImageWorkbenchDB().then((db) =>
    new Promise((resolve, reject) => {
      const tx = db.transaction(storeName, mode)
      const store = tx.objectStore(storeName)
      const req = run(store)
      req.onsuccess = () => resolve(req.result)
      req.onerror = () => reject(req.error)
    })
  )
}

export function getAllWorkbenchTasks() {
  return transaction(STORE_TASKS, 'readonly', (store) => store.getAll()).then((items) =>
    (items || []).sort((a, b) => Number(b.createdAt || 0) - Number(a.createdAt || 0))
  )
}

export function putWorkbenchTask(task) {
  return transaction(STORE_TASKS, 'readwrite', (store) => store.put(task))
}

export function deleteWorkbenchTask(id) {
  return transaction(STORE_TASKS, 'readwrite', (store) => store.delete(id))
}

export function getWorkbenchImage(id) {
  return transaction(STORE_IMAGES, 'readonly', (store) => store.get(id))
}

export function putWorkbenchImage(image) {
  return transaction(STORE_IMAGES, 'readwrite', (store) => store.put(image))
}

export function putWorkbenchThumbnail(thumbnail) {
  return transaction(STORE_THUMBNAILS, 'readwrite', (store) => store.put(thumbnail))
}

export function getWorkbenchThumbnail(id) {
  return transaction(STORE_THUMBNAILS, 'readonly', (store) => store.get(id))
}

export async function storeWorkbenchImage(dataUrl, source = 'generated', name = '') {
  const id = await hashDataUrl(dataUrl)
  const existing = await getWorkbenchImage(id)
  if (!existing) {
    const metadata = await createImageMetadata(dataUrl)
    await putWorkbenchImage({
      id,
      dataUrl,
      source,
      name,
      mime: metadata.mime,
      width: metadata.width,
      height: metadata.height,
      createdAt: Date.now()
    })
    if (metadata.thumbnailDataUrl) {
      await putWorkbenchThumbnail({
        id,
        thumbnailDataUrl: metadata.thumbnailDataUrl,
        width: metadata.width,
        height: metadata.height,
        createdAt: Date.now()
      })
    }
  }
  return id
}

export async function getWorkbenchImagePreview(id) {
  const thumbnail = await getWorkbenchThumbnail(id)
  if (thumbnail?.thumbnailDataUrl) {
    return thumbnail.thumbnailDataUrl
  }
  const image = await getWorkbenchImage(id)
  return image?.dataUrl || ''
}

async function hashDataUrl(dataUrl) {
  if (globalThis.crypto?.subtle) {
    const bytes = new TextEncoder().encode(dataUrl)
    const hash = await crypto.subtle.digest('SHA-256', bytes)
    return Array.from(new Uint8Array(hash))
      .map((value) => value.toString(16).padStart(2, '0'))
      .join('')
  }
  let hash = 2166136261
  for (let i = 0; i < dataUrl.length; i += 1) {
    hash ^= dataUrl.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return `fallback-${(hash >>> 0).toString(16)}`
}

function createImageMetadata(dataUrl) {
  return new Promise((resolve) => {
    const image = new Image()
    image.onload = () => {
      const width = image.naturalWidth || 0
      const height = image.naturalHeight || 0
      resolve({
        mime: parseDataUrlMime(dataUrl),
        width,
        height,
        thumbnailDataUrl: createThumbnailDataUrl(image, width, height)
      })
    }
    image.onerror = () => resolve({ mime: parseDataUrlMime(dataUrl), width: 0, height: 0, thumbnailDataUrl: '' })
    image.src = dataUrl
  })
}

function parseDataUrlMime(dataUrl) {
  const match = String(dataUrl || '').match(/^data:([^;]+);base64,/)
  return match?.[1] || 'image/png'
}

function createThumbnailDataUrl(image, width, height) {
  if (!width || !height) return ''
  const scale = Math.min(1, THUMBNAIL_MAX_SIZE / Math.max(width, height))
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, Math.round(width * scale))
  canvas.height = Math.max(1, Math.round(height * scale))
  const ctx = canvas.getContext('2d')
  if (!ctx) return ''
  ctx.drawImage(image, 0, 0, canvas.width, canvas.height)
  return canvas.toDataURL('image/jpeg', THUMBNAIL_QUALITY)
}
