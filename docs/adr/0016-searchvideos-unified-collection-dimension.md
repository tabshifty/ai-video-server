# 短视频合集投屏统一走 SearchVideos 的 collection 维度

合集投屏需要 TV 端在已加载页放完后按合集补页。我们决定把 `collection_id` 作为 `SearchVideos` 的一个统一过滤维度（最右追加可选参数 `collectionID *uuid.UUID`，nil 不过滤），让 TV 端投屏补页对"关键词搜索"和"合集"走同一条链路，而不是为合集另开 `DiscoverShortVideos` 补页分支。

## 考虑过的备选

- **A2 双链路**：合集补页改调 `DiscoverShortVideos(collection_id)`，关键词补页继续走 `SearchVideos`。被否：TV 端补页要分两条分支，长期维护负担大，且两段 SQL 排序/过滤各自演化会产生手机端顺序 ≠ TV 端补页顺序的错位 bug。
- **A3 只投手机端已加载页、不补页**：被否，等于推翻"投屏 = 整个合集有序列表连放"的合集投屏定义。

## 后果

- `SearchVideos` 带 collection_id 的过滤/排序 SQL 必须与 `DiscoverShortVideos` 的 `mode=collection` 分支共享同一段查询构造，从根上防漂移（见 ADR-0017 投屏可见性边界）。
- `TvRemoteSearchContext` 新增 `CollectionID *uuid.UUID` 字段；`search_context` 是 JSONB 列，无需新迁移，老会话反序列化为 nil 向后兼容。
- 现有 `SearchVideos` 调用方（app 搜索、tv_auth 电影/AV 墙等）传 nil 保持原行为。
