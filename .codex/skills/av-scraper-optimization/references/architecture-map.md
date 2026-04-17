# mdcx 到 Go 的架构映射

## 目标
在不破坏现有外部接口的前提下，把 AV 刮削实现拆成可扩展结构。

## 建议映射
- Python `CrawlerProvider` -> Go `avCrawlerProvider`
  - 负责站点 crawler 生命周期与注册。
- Python `BaseCrawler._run` 模板流程 -> Go `SearchCandidates + FetchByDetailURL` 统一流程
  - 生成搜索 query
  - 请求搜索页
  - 解析详情 URL
  - 请求详情页
  - 解析字段
  - post-process
- Python `Context` -> Go `avScrapeRunContext`
  - 记录 search queries/search urls/detail urls/steps/errors。
- Python `DetailPageParser` -> Go `avDetailParser`
  - 字段级提取并区分字段状态（filled/empty/not_supported）。

## 字段状态建议
- `filled`: 站点支持且提取成功
- `empty`: 站点支持但当前页面未提取到值
- `not_supported`: 该站点/实现当前不支持该字段

## 接入新站点最小清单
1. 实现 `avCrawler` 的 `SearchCandidates` 与 `FetchByDetailURL`
2. 实现站点 `avDetailParser`
3. 增加 post-process（URL 修正、分类字段等）
4. 在 provider 注册 crawler
5. 补 5 类测试：搜索、详情、编号、关联演员、异常 trace
