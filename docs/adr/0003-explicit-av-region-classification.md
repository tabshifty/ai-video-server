# 上传时由管理员显式选择 AV 地区分类，不沿用文件名启发式

代码原先有 `detectAVSiteCategory`（`internal/services/scraper_av_strategy.go:112`）在 scrape 时按文件名启发式推断 AV 是 fc2 / western / japanese，但启发式对真实文件名（缩写、混杂、乱码、各种站点风格）猜错率不低，错路由到错的站点列表就直接刮不到。决定**在上传表单上**让管理员从 `日本 / 欧美 / FC2` 显式三选一（默认 `日本`），落 `videos.metadata.site_category`——把"猜"换成"事实"，启发式降级为 retag 或历史数据缺失时的兜底。这个分类**存在的唯一理由**是为 [`欧美 AV`] 专属的 oshash 计算与 [`刮削确认门控`] 提供门控依据，不允许被未来其它特性借用作"内容地区标签"等通用语义。

## 考虑过的替代方案

- **保留启发式不暴露给用户**：上传体验最简单，但保留了"启发式不命中"的老问题；且 FC2 与 western 文件名变种太多，启发式准确率难以靠数据驱动调优
- **两档简化（吞 FC2 进日本）**：UI 极简，但 FC2 站点配置完全独立（`fc2ppvdb` / `fc2club` / `fc2` / `fc2hub`），合并后要么用错站点列表刮不到，要么仍走启发式区分——只是把复杂度移到内部
- **两档 + 自动 FC2 detect**：UI 暴露日本 / 欧美，FC2 由启发式识别。优点 UI 简单，缺点保留启发式失败的老问题，且用户把 FC2 上传成"日本"时无法解释为啥刮不到

## 关联

- `internal/handlers/upload.go` 接收 `site_category` 字段
- `admin-web/src/views/VideoUpload.vue` AV 地区 radio
- `CONTEXT.md` 中 [`AV 地区分类`] 术语条款关于"不允许被独立扩展用途"的约束
