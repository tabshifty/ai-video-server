---
name: av-scraper-optimization
description: Use when optimizing or extending Go-based AV scraping pipelines with site adapters, parser abstraction, number normalization, and actor auto-linking.
---

# AV 刮削优化技能

## 何时使用
- 需要优化 Go AV 刮削准确率、可维护性与可扩展性。
- 需要从单站点逻辑演进到可扩展多站点架构。
- 需要增强编号解析、匹配优先级、演员自动关联与调试能力。

## 快速流程
1. 先确认现有链路：上传触发、预览、确认、队列任务。
2. 引入 AV crawler provider + crawler 模板流程，不改外部 API。
3. 为站点实现 parser 抽象和 post-process。
4. 升级编号标准化与候选排序。
5. 增加 scrape trace，记录 query/url/错误/匹配策略。
6. 验证自动刮削中文输出与演员关联是否稳定。

## 实施规范
- 默认保持接口兼容：不修改现有 handler 路径与请求结构。
- 新增能力优先放在服务层内部（provider/crawler/parser/context）。
- 站点首期可只实现 JavDB，但必须预留多站点注册点。
- 对所有新增行为补单元测试，优先覆盖：
  - 编号解析（FC2/HEYZO/素人编号等）
  - 搜索 query 生成与优先级
  - 详情字段提取与空字段语义
  - 演员自动关联
  - 失败路径 trace

## 参考资料
- 架构映射：`references/architecture-map.md`
- 编号规范化：`references/number-normalization.md`

