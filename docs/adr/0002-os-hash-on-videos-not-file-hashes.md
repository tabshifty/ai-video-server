# `os_hash` 放在 `videos` 表新列，不扩展 `file_hashes`

OpenSubtitles 风格 oshash 用于 ThePornDB `/scenes/hash/{hash}` 查找，是**原始文件物理指纹**，跟现有 `file_hashes` 表里承担去重职责的 SHA256 **用途、长度、唯一性约束都不同**——oshash 16 位 hex 且允许撞库（极罕见但非零），SHA256 64 位 hex 且全表 UNIQUE。决定在 `videos` 表加 `os_hash CHAR(16)` 列 + partial index `WHERE os_hash IS NOT NULL`，与 `file_hashes` 完全解耦。partial index 让"绝大多数 row 是 NULL（短视频 / movie / 日本 AV / FC2 AV / 全部存量）"的现实下 B-tree 不浪费空间，同时给"用 hash 反查现有视频"留出 O(log n) 查询路径。

## 考虑过的替代方案

- **扩 `file_hashes` 表加 `hash_type` 列**：所有"文件指纹"集中管理，但 (1) 现有 `UNIQUE INDEX idx_file_hashes_hash` 把 hash 列设为全表唯一，oshash 撞库会插入失败，要拆 UNIQUE 等于动手术；(2) `file_hashes` 当前所有调用方都默认 SHA256 语义（去重、秒传），加 `hash_type` 后每个调用点都要补判断
- **`videos.metadata.os_hash` JSONB 子键**：零迁移，但 (1) oshash 是物理属性而非"刮削结果"，放 metadata 会混入语义层级；(2) "已上传过这个 hash"查询要 `WHERE metadata->>'os_hash' = $1`，写法和性能都不如 column；(3) 后续若要 JOIN videos ON os_hash 会很别扭

## 关联

- `migrations/0021_western_av_oshash_gate.up.sql`
- `pkg/oshash/oshash.go`
- `CONTEXT.md` 中 [`欧美 AV`] 术语关于"一次计算多路复用"的约束
