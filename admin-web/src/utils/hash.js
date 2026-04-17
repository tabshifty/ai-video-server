import { sha256 } from 'js-sha256'

function toHex(bytes) {
  return Array.from(bytes)
    .map((x) => x.toString(16).padStart(2, '0'))
    .join('')
}

export async function sha256File(file, onProgress) {
  const chunkSize = 2 * 1024 * 1024
  const chunks = Math.max(1, Math.ceil(file.size / chunkSize))
  const webCrypto = globalThis?.crypto?.subtle
  const useWebCrypto = typeof webCrypto?.digest === 'function'

  const buffers = useWebCrypto ? [] : null
  const fallbackHasher = useWebCrypto ? null : sha256.create()

  for (let i = 0; i < chunks; i += 1) {
    const start = i * chunkSize
    const end = Math.min(start + chunkSize, file.size)
    const chunk = new Uint8Array(await file.slice(start, end).arrayBuffer())
    if (useWebCrypto) {
      buffers.push(chunk)
    } else {
      fallbackHasher.update(chunk)
    }
    if (onProgress) onProgress(Math.round(((i + 1) / chunks) * 100))
  }

  if (useWebCrypto) {
    const merged = new Uint8Array(file.size)
    let offset = 0
    for (const b of buffers) {
      merged.set(b, offset)
      offset += b.length
    }
    const digest = await webCrypto.digest('SHA-256', merged)
    return toHex(new Uint8Array(digest))
  }

  return fallbackHasher.hex()
}
