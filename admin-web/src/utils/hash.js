export async function sha256File(file, onProgress) {
  const chunkSize = 2 * 1024 * 1024
  const chunks = Math.ceil(file.size / chunkSize)
  const buffers = []

  for (let i = 0; i < chunks; i += 1) {
    const start = i * chunkSize
    const end = Math.min(start + chunkSize, file.size)
    const buffer = await file.slice(start, end).arrayBuffer()
    buffers.push(new Uint8Array(buffer))
    if (onProgress) onProgress(Math.round(((i + 1) / chunks) * 100))
  }

  const merged = new Uint8Array(file.size)
  let offset = 0
  for (const b of buffers) {
    merged.set(b, offset)
    offset += b.length
  }

  const digest = await crypto.subtle.digest('SHA-256', merged)
  return [...new Uint8Array(digest)].map((x) => x.toString(16).padStart(2, '0')).join('')
}
