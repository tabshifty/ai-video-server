# TV 电视剧播放器用共享控制层承载混合播放内核

电视剧播放器允许同一部剧内的不同分集按播放安全策略选择不同内核：普通非 DV 分集继续走 LibVLC，满足门控的 DV 分集走 Media3/ExoPlayer 专用链路。为避免用户在选集、连播、续播和遥控操作上感知到两个播放器，本决策把电视剧核心互动抽成共享控制层，由 LibVLC 和 Media3 分别做播放内核适配；这不是恢复通用 Player fallback，也不允许 DV 分集失败后回退 LibVLC 强行播放。

## 考虑过的替代方案

- **维持两套播放器 UI**：实现短期成本最低，但 DV 分集会缺少选集轨、连播和统一遥控体验；后续维护也会出现两套控制条、两套路由和两套错误处理。
- **切换内核时退出到详情页**：避免同页内核切换，但会让混合剧集播放体验割裂，也破坏自动连播语义。
- **把所有长视频都改成通用 Player 抽象**：表面最统一，但会扩大到单片、字幕、音轨和 LibVLC ASS 能力，明显超过当前 DV 剧集问题的范围。

## 关联

- 前置 ADR：`0004-tv-long-form-libvlc-for-ass-rendering.md`、`0010-dolby-vision-tv-playback-compatibility.md`、`0011-dolby-vision-episode-original-file-preservation.md`
- `CONTEXT.md` 术语：[[TV 播放控制层]]、[[TV 播放内核适配]]、[[电视剧核心控制层优先抽离]]、[[混合剧集播放链路切换]]、[[DV 专用链路核心互动层]]
