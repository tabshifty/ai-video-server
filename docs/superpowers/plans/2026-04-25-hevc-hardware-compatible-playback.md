# HEVC 硬件转码与兼容播放方案 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 保持后端“硬件加速优先、主产物优先使用 H.265/HEVC、体积尽量小”的前提下，补齐 Android/模拟器无法解码 HEVC 时的兼容播放链路，避免电视剧、电影、AV 在 app 端立即播放失败。

**Architecture:** 后端转码改为“双产物”策略：主产物继续使用硬件 HEVC，小体积优先；兼容产物额外使用硬件 AVC/H.264。播放源接口新增播放档位选择，Android app 根据本机解码能力优先请求 HEVC，遇到不支持或高风险档位时自动请求 AVC 兼容流。失败场景补充显式错误提示与管理端重转码入口，避免再次出现“首播即 CodecException”但界面无明确反馈。

**Tech Stack:** Go、FFmpeg/Videotoolbox、PostgreSQL、Android Media3/ExoPlayer、Kotlin、Compose

---

### Task 1: 固化现状与失败样本

**Files:**
- Modify: `plan.md`
- Test: `internal/services/transcode_test.go`
- Test: `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`

- [ ] **Step 1: 追加后端转码侧失败样本测试**

目标：
- 锁定“主转码产物不再只存在单一 codec 输出假设”
- 为后续双产物元数据和路径选择提供测试入口

建议新增测试：
- `TestResolveProbeFieldsRecordsOutputCodecMetadata`
- `TestTranscodeResultMetadataSupportsPrimaryAndCompatProfiles`

- [ ] **Step 2: 追加 Android 侧播放档位选择测试**

目标：
- 锁定“当设备不支持 HEVC 时，app 不再直接请求默认 HEVC 地址”
- 锁定“播放器错误落 UI 状态，而不是只有底层 codec 堆栈”

建议新增测试：
- `selectPlaybackProfile_prefersHevcWhenSupported`
- `selectPlaybackProfile_fallsBackToCompatWhenHevcUnsupported`

- [ ] **Step 3: 运行失败测试确认红灯**

Run:
```bash
go test ./internal/services -run 'TestResolveProbeFieldsRecordsOutputCodecMetadata|TestTranscodeResultMetadataSupportsPrimaryAndCompatProfiles' -v
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.tv.*'
```

Expected:
- 新增测试失败，且失败原因指向“当前只有单一 HEVC 输出 / app 不区分播放档位”

- [ ] **Step 4: Commit**

```bash
git add plan.md internal/services/transcode_test.go android-app/app/src/test/java/com/chee/videos/feature/tv/*.kt
git commit -m "补充 HEVC 兼容播放失败样本测试"
```

### Task 2: 将 FFmpeg 转码抽象为硬件编码档位

**Files:**
- Modify: `pkg/ffmpeg/ffmpeg.go`
- Test: `pkg/ffmpeg/ffmpeg_test.go`

- [ ] **Step 1: 写失败测试，锁定硬件编码档位定义**

新增测试目标：
- `preferredHardwareHevcEncoder()` 返回 `hevc_videotoolbox`
- `preferredHardwareAvcEncoder()` 返回 `h264_videotoolbox`
- `buildTranscodeVideoArgs()` 能根据 profile 生成不同 codec 参数
- 不允许静默落到软件编码器

- [ ] **Step 2: 运行测试确认失败**

Run:
```bash
go test ./pkg/ffmpeg -run 'TestPreferredHardware|TestBuildTranscodeVideoArgs' -v
```

Expected:
- 当前缺少 profile 抽象与 AVC 硬编参数，测试失败

- [ ] **Step 3: 实现最小代码**

实现要点：
- 新增 `TranscodeProfile`，至少包含 `hevc_primary` 和 `avc_compat`
- 抽出 `buildTranscodeVideoArgs(input, output, profile, options)`
- 新增 `TranscodeVideo(ctx, inputPath, outputPath, profile, options)`，严格只走硬件编码器
- 若目标硬件编码器不可用，直接返回明确错误，例如 `hardware encoder unavailable for avc_compat`

- [ ] **Step 4: 运行测试确认通过**

Run:
```bash
go test ./pkg/ffmpeg -run 'TestPreferredHardware|TestBuildTranscodeVideoArgs|TestIsEncoderUnavailableOutput' -v
```

- [ ] **Step 5: Commit**

```bash
git add pkg/ffmpeg/ffmpeg.go pkg/ffmpeg/ffmpeg_test.go
git commit -m "抽象硬件视频转码档位"
```

### Task 3: 后端转码服务改为双产物输出

**Files:**
- Modify: `internal/services/transcode.go`
- Modify: `internal/services/transcode_test.go`
- Modify: `internal/models/models.go`
- Modify: `internal/models/admin.go`
- Modify: `internal/models/app.go`

- [ ] **Step 1: 写失败测试，锁定双产物行为**

新增测试目标：
- `Process()` 对长视频默认输出 HEVC 主产物
- `Process()` 额外输出 AVC 兼容产物
- 元数据中记录 `primary_codec=hevc`、`compat_codec=h264`
- 若兼容产物硬编失败，返回显式错误，不伪装成成功

- [ ] **Step 2: 运行测试确认失败**

Run:
```bash
go test ./internal/services -run 'TestProcessProducesPrimaryAndCompatOutputs|TestProcessFailsWhenCompatHardwareEncoderUnavailable' -v
```

- [ ] **Step 3: 实现最小代码**

实现要点：
- `TranscodeResult` 增加：
  - `PrimaryPath`
  - `CompatPath`
  - `PrimaryCodec`
  - `CompatCodec`
- `Process()`：
  - 主产物：`video-hevc.mp4`
  - 兼容产物：`video-avc.mp4`
  - 缩略图继续从主产物或兼容产物选一个稳定来源生成
- metadata 增加：
  - `playback_profiles`
  - `primary_codec`
  - `compat_codec`
  - `compat_available`

- [ ] **Step 4: 运行测试确认通过**

Run:
```bash
go test ./internal/services -run 'TestProcessProducesPrimaryAndCompatOutputs|TestResolveProbeFields' -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/services/transcode.go internal/services/transcode_test.go internal/models/models.go internal/models/admin.go internal/models/app.go
git commit -m "增加 HEVC 主产物与 AVC 兼容产物"
```

### Task 4: 持久化兼容播放路径并开放播放档位

**Files:**
- Modify: `internal/repository/video_repository.go`
- Modify: `internal/handlers/video_source.go`
- Modify: `internal/handlers/router.go`
- Test: `internal/handlers/admin_retranscode_test.go`
- Test: `internal/utils/video_url_test.go`

- [ ] **Step 1: 写失败测试**

新增测试目标：
- `/api/v1/videos/:id/source?profile=primary` 返回 HEVC 主产物
- `/api/v1/videos/:id/source?profile=compat` 返回 AVC 兼容产物
- 未传 `profile` 时仍默认主产物
- 请求 compat 且兼容产物不存在时，返回明确错误，不静默回退

- [ ] **Step 2: 运行测试确认失败**

Run:
```bash
go test ./internal/handlers -run 'TestVideoSourceSelectsPlaybackProfile|TestSelectRetranscodeInputPath' -v
```

- [ ] **Step 3: 实现最小代码**

实现要点：
- 数据库新增 `compat_transcoded_path` 或在 `metadata` 中持久化兼容路径
- `resolvePlayableSource()` 读取 `profile` 查询参数
- `profile=primary|compat`
- 管理端重转码接口保留，但新增 profile 相关可见信息

- [ ] **Step 4: 运行测试确认通过**

Run:
```bash
go test ./internal/handlers -run 'TestVideoSourceSelectsPlaybackProfile|TestSelectRetranscodeInputPath' -v
go test ./internal/repository -run Video -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/repository/video_repository.go internal/handlers/video_source.go internal/handlers/router.go internal/handlers/*.go
git commit -m "为视频源增加主档与兼容档选择"
```

### Task 5: Android 增加 HEVC 能力探测与播放档位选择

**Files:**
- Modify: `android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt`
- Modify: `android-app/app/src/main/java/com/chee/videos/core/repository/VideoRepository.kt`
- Modify: `android-app/app/src/main/java/com/chee/videos/feature/tv/TvRepository.kt`
- Modify: `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt`
- Create: `android-app/app/src/main/java/com/chee/videos/core/player/PlaybackProfileResolver.kt`
- Test: `android-app/app/src/test/java/com/chee/videos/core/player/PlaybackProfileResolverTest.kt`
- Test: `android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt`

- [ ] **Step 1: 写失败测试**

新增测试目标：
- 设备支持 HEVC 时，请求 `profile=primary`
- 设备不支持 HEVC 时，请求 `profile=compat`
- ViewModel 切集时继续保留上一轮修复的请求版本控制

- [ ] **Step 2: 运行测试确认失败**

Run:
```bash
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.core.player.*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'
```

- [ ] **Step 3: 实现最小代码**

实现要点：
- 新增 `PlaybackProfileResolver`
- 使用 `MediaCodecList` / `MediaCodecInfo.CodecCapabilities` 探测 HEVC Main/Main10 支持情况
- `UrlBuilder.source(baseUrl, videoId, profile)`
- 电视剧、电影、AV 的长视频链路统一走该 resolver

- [ ] **Step 4: 运行测试确认通过**

Run:
```bash
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.core.player.*' --tests 'com.chee.videos.feature.tv.TvSeriesPlayerViewModelTest'
```

- [ ] **Step 5: Commit**

```bash
git add android-app/app/src/main/java/com/chee/videos/core/player android-app/app/src/main/java/com/chee/videos/core/util/UrlBuilder.kt android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModel.kt android-app/app/src/test/java/com/chee/videos/core/player android-app/app/src/test/java/com/chee/videos/feature/tv/TvSeriesPlayerViewModelTest.kt
git commit -m "按设备能力选择 HEVC 或兼容播放档位"
```

### Task 6: app 端补齐播放器错误态与管理端可见性

**Files:**
- Modify: `android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`
- Modify: `android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt`
- Modify: `android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt`
- Modify: `web-admin` 中视频详情/转码信息相关页面（按实际目录补齐）

- [ ] **Step 1: 写失败测试**

新增测试目标：
- `onPlayerError` 时 UI 出现明确错误文案
- 错误文案区分“编解码器不支持”和“源文件不存在/未就绪”

- [ ] **Step 2: 运行测试确认失败**

Run:
```bash
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest --tests 'com.chee.videos.feature.*player*'
```

- [ ] **Step 3: 实现最小代码**

实现要点：
- 三个长视频页面监听 `onPlayerError`
- 遇到 codec unsupported 时提示：
  - `当前设备不支持该清晰度/编码，已尝试兼容流；若仍失败，请在管理端重新生成兼容转码`
- 管理端展示：
  - 主档 codec
  - 兼容档 codec
  - 主/兼容文件路径
  - 兼容档是否已生成

- [ ] **Step 4: 运行测试确认通过**

Run:
```bash
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug
```

- [ ] **Step 5: Commit**

```bash
git add android-app/app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt android-app/app/src/main/java/com/chee/videos/feature/detail/DetailScreen.kt android-app/app/src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt
git commit -m "补充长视频播放档位错误提示"
```

### Task 7: 端到端验证与回填历史内容

**Files:**
- Modify: `plan.md`

- [ ] **Step 1: Go 全量验证**

Run:
```bash
go test ./...
```

- [ ] **Step 2: Android 全量验证**

Run:
```bash
cd android-app && GRADLE_USER_HOME="$PWD/.gradle-local" ./gradlew :app:testDebugUnitTest :app:assembleDebug
```

- [ ] **Step 3: 手工验证**

验证路径：
- Android 模拟器播放 HEVC 主档不可解码的视频，自动请求 compat 成功播放
- 真机支持 HEVC 时继续走主档 HEVC
- 电视剧、电影、AV 三类长视频都能走同一套播放档位选择
- 管理端可见主/兼容档信息

- [ ] **Step 4: 追加 `plan.md` 实施记录并提交**

```bash
git add plan.md
git commit -m "记录 HEVC 兼容播放实施结果"
```

---

## Self-Review

- Spec coverage:
  - “转码一定要硬件加速” 已覆盖：Task 2、Task 3 明确只走硬件编码器，不引入软件兜底。
  - “优先转成占用空间最小的 h265” 已覆盖：Task 3 保持 HEVC 主产物为默认主档。
  - “Android 立即播放不能失败” 已覆盖：Task 4、Task 5、Task 6 通过兼容档和错误态处理补齐。
- Placeholder scan:
  - 已给出明确文件、测试方向、命令和落点，没有使用 TBD/TODO 占位。
- Type consistency:
  - 计划统一使用 `primary` / `compat` 作为播放档位命名，避免后续实现命名漂移。

Plan complete and saved to `docs/superpowers/plans/2026-04-25-hevc-hardware-compatible-playback.md`. Two execution options:

**1. Subagent-Driven (recommended)** - 我按任务拆分并逐步审查执行

**2. Inline Execution** - 在当前会话里按计划顺序直接实现
