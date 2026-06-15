# 杜比视界剧集保留原始文件以支持专用链路播放

当前转码流程对所有视频执行 uploadGC（删除原始上传文件），且转码输出强制 8-bit yuv420p、不保留 Dolby Vision RPU 元数据——这意味着 DV 源文件一旦被转码，原始 DV 信息永久丢失，TV 端已实现的 Media3/ExoPlayer 专用 DV 播放链路永远无法被触达。本决策针对电视剧集（episode）放宽存储约束：转码时若源文件被探测为 Dolby Vision，跳过 uploadGC，将原始文件移至 `storageRoot/videos/<uuid>/source-dv.<ext>` 永久保留，并通过 `playback_compat.source_playback_path` 暴露给 TV 端作为专用链路播放源。

放宽范围仅限 episode——movie 和 short 等类型仍遵守现有 `杜比视界原始文件不保留约束`。选择 episode 而非全部类型，是因为电视剧集是 DV 内容最常见的载体，且单集体积可控；movie 类 DV 源文件通常极大（40-80GB remux），全量保留的存储压力不可接受。该决策还附带禁止对持有 DV 原始文件的剧集执行重转码，因为 DV 播放已使用原始文件，重转码既无意义也会破坏现有播放链路。

## 考虑过的替代方案

- **对 DV 源执行保留 RPU 的 remux 或 re-encode**：理论上可以生成体积更小的 DV 兼容文件，但当前 `hevc_videotoolbox` 编码器和 ffmpeg 的 RPU 通路均不支持可靠保留 Dolby Vision 元数据，验证成本高且结果不可预期。
- **全类型保留 DV 原始文件**：安全边界最完整，但 movie 类 DV remux 文件体积极大，与 `媒体存储空间优先` 策略直接冲突。
- **不保留任何原始文件，维持全量阻断**：零存储成本，但 DV 剧集永远无法在 TV 端播放，已实现的专用播放链路形同虚设。
- **在 `playback_compat` 外新增数据库列存储 DV 播放源路径**：结构更显式，但破坏了 `playback_compat` 作为播放策略唯一输入的设计；新增列还需要迁移和索引维护。

## 关联

- `CONTEXT.md` 新术语：[[杜比视界剧集原始文件保留]]、[[杜比视界剧集播放源路径]]、[[杜比视界剧集播放源 profile]]、[[杜比视界剧集禁止重转码]]、[[杜比视界剧集精确删除]]、[[杜比视界存量数据维持阻断]]
- `CONTEXT.md` 更新术语：[[媒体存储空间优先]]（增加剧集 DV 例外）、[[杜比视界原始文件不保留约束]]（限定为非 episode 或非 DV）
- 前置 ADR：`0010-dolby-vision-tv-playback-compatibility.md`
