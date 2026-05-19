# 项目上下文

## 技术沉淀约定
- 每次功能更新必须补充本文件，沉淀长期有效的术语、架构决策、接口约定、兼容策略、踩坑经验或后续维护注意事项。
- 技术沉淀只记录会影响后续实现和维护的内容，不记录临时进度、命令流水账或一次性调试过程。

## 管理端手动刮削术语
- `手动刮削`：管理员对已有视频主动发起候选预览、人工选择候选并确认覆盖视频 metadata 的操作，不等同于上传、改类型或后台任务触发的自动刮削。
- 电影手动刮削复用通用刮削页 `/scrape` 和现有 `POST /admin/scrape/preview`、`PUT /admin/scrape/confirm` 接口；详情页入口只通过 query 预填 `video_id`、`type=movie`、`title`，可解析到发行年份时追加 `year`，进入页面后必须由管理员手动点击“查询预览”。

## TV 首页术语
- `18+`：面向 TV 端显示的成人内容入口。代码、接口参数和存储内部仍使用 `av`，TV 界面文案不得显示为 `AV`。
- `最近更新`：当前类型下最新可播放内容；电视剧按最近剧集更新，电影和 `18+` 按最新可播放视频。
- `最近播放`：当前类型下用户已有播放进度的内容；电视剧保留季/集信息，电影和 `18+` 指向长视频本体。

## TV 首页结构
- 一级首页使用左侧窄菜单和右侧内容区。
- 一级菜单固定为：`电视剧`、`电影`、`18+`、`IPTV`、`搜索`、`设置`，默认选中 `电视剧`。
- `电视剧`、`电影`、`18+` 共享类型化首页结构：巨幅推荐、最近播放、最近更新、全部入口。
- 左侧菜单只属于 `tv-home` 一级页；海报墙、详情页、播放器页保持现有全屏返回体验。
- TV 首页左侧菜单按钮只能有一个焦点目标；`tvFocusableGlow()` 已经包含 `focusable()`，按钮自身不要再额外叠加 `.focusable()`，否则遥控确认键可能先落到重复焦点层，表现为必须按两次才触发菜单动作。
- TV 首页初始焦点只能请求当前已经组合到界面树里的节点。默认首页内容为空或接口失败导致没有巨幅推荐/最近播放/货架时，应回退到左侧菜单焦点；不要请求仅在“搜索”菜单下才会组合的搜索框焦点，否则 Compose 会因为 `FocusRequester` 未绑定节点在冷启动时崩溃。

## TV 端 AV 海报约定
- TV 端 `18+` 的大背景指首页/分类页巨幅推荐背景和详情页顶部背景，应优先使用 AV 元数据中的原始横幅海报 `poster_original_path`。
- TV API 给首页和分类巨幅推荐使用的 AV `backdrop_url` 由后端统一解析；fallback 顺序为 `poster_original_path`、`poster_url`/`poster_path`、抓取来源块内的 `poster_url`/`poster_path`、`poster_cropped_path`、`thumbnail_path`。
- AV 海报墙竖卡和详情页左侧小海报继续使用缩略图或裁剪图，不改成横幅图，避免破坏竖向卡片比例。

## TV 沉浸式详情首屏
- `沉浸式详情首屏`：TV 端电影和 `18+` 共用的长视频详情页首屏样式，使用全屏横向背景和底部半透明信息面板承载标题、年份/时长/标签、简介、演员头像与播放/收藏操作；电视剧详情页不套用此约定。
- 长视频详情背景优先使用可识别的横向背景字段：电影使用 `backdrop_url`、`backdrop_path`、`fanart_url`、`fanart_path`；`18+` 保持 AV 原始横幅海报优先规则。缺少横向背景时才允许使用竖向海报作为模糊暗化兜底，并通过展示模型显式标记为兜底，避免把竖海报当作正常横幅裁切。
- TV 电影/`18+` 详情页操作只保留“播放”和“收藏/取消收藏”；收藏状态复用详情 DTO 的 `user_state.is_favorited` 与 `DetailViewModel.toggleFavorite()`，不新增分享入口或更多信息滚动入口。
- 详情页演员区最多展示 5 个演员，优先使用接口 `actors.avatar_url`，无头像时显示圆形文字占位；无演员数据时隐藏演员区，不发起额外抓取。

## IPTV 术语
- `IPTV 播放列表`：后台维护的一份全局 M3U/M3U8 频道源；本期只允许一个播放列表生效，上传文件或远程刷新都会替换当前频道清单。
- `频道/电视台`：M3U 中一条可播放的直播流，必须包含频道名和 `http/https` 播放 URL；`group-title`、`tvg-logo`、`tvg-id` 是可选信息。
- `频道分组`：M3U 的 `group-title`；TV 端频道列表按分组展示，缺失分组时显示为 `未分组`。
- `当前播放频道`：TV 端 IPTV 播放页正在直连播放的频道；默认使用频道清单第一项，遥控器上下键按原始顺序首尾循环切换。

## IPTV 接口约定
- 后端负责保存和解析 M3U，Admin Web 负责上传文件、保存远程 URL、手动刷新和预览解析结果；TV App 只获取解析后的频道清单。
- TV App 播放 IPTV 时直接连接 M3U 中的原始频道 URL；后端不代理、不转码、不隐藏源地址。
- 远程 M3U URL 只在后台手动刷新时拉取，本期不做定时刷新和 EPG 节目单。
- TV App 播放 M3U8/HLS 频道必须打包 Media3 HLS 扩展模块；仅引入 `media3-exoplayer` 时，`DefaultMediaSourceFactory` 会在运行时找不到 `HlsMediaSource.Factory` 并导致 IPTV 播放页崩溃。
- Media3 `PlayerView` 的 `texture_view` 只能排除部分 Compose/Surface 渲染黑屏；如果 IPTV 日志中音频解码器已初始化但没有视频解码器初始化，且出现多个 `VideoCapabilities` 不支持提示，应按直播源视频编码兼容性处理。
- TV App 的 IPTV 播放页使用 LibVLC `VLCVideoLayout` 独立播放，在 Compose `AndroidView` 中必须让 LibVLC 使用 TextureView 输出，并关闭硬解优先走软解；其他长视频播放器继续使用 Media3。IPTV 源常见 MPEG2、特殊 H.264 profile、Dolby Vision 或设备厂商解码缺口，不能只依赖 Android TV 设备硬解能力。
- IPTV 播放无画面排查必须保留播放事件诊断：至少记录 LibVLC event、vout 数、视频轨/音频轨数量和当前视频轨 codec/分辨率。若 `videoTracks=0` 或无 `Vout`，优先检查 M3U 频道 URL、源站多码率/分片返回和是否需要后端转码；若有视频轨和 `Vout` 但仍黑屏，优先检查输出视图/系统合成层。
- IPTV 频道清单不应包含音频专用源。明显音频源包括 `group-title=Audio/音频`、频道名包含 `音频` 或 `audio only`、URL 路径包含 `/audio/`、`_audio/`，以及 `.mp3/.aac/.m4a/.flac/.wav/.ogg/.opus` 等音频文件。后端解析时应跳过，TV 端也要过滤旧数据，避免默认播放只有声音没有画面的条目。
- TV App IPTV 顶部频道信息不是常驻条：进入播放页和切换频道后仅显示 3 秒临时提示，频道列表打开、加载中、无可播放频道、状态错误或播放错误时必须隐藏。临时提示仍使用后端返回的 `logo_url`（M3U `tvg-logo`）展示台标，`logo_url` 为空或图片加载失败时使用 TV 图标回退，不新增占位资源。
- IPTV 频道列表的 `LazyColumn` item 顺序固定为：标题、分组标题、频道行。遥控器打开列表时应直接初始化到当前播放频道附近，不应先从顶部出现再动画滚动；焦点上下移动时必须按频道 id 映射到实际 item index 并保留短动画滚动到可见区域，避免分组标题导致索引偏移。
- TV App 分发仍使用 APK，不切换 AAB；由于 IPTV 播放兼容性优先保留 `libvlc-all`，APK 必须按 ARM ABI 拆分输出 `armeabi-v7a` 和 `arm64-v8a` 两个安装包，不生成 x86/x86_64 或 universal APK。手动侧载时按设备 ABI 选择对应 APK；仓库不保存 keystore，Release 签名继续由 Android Studio 本机签名配置完成。
