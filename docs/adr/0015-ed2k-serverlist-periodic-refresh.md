# ED2K 服务器列表周期刷新走 asynq 定时任务 + 重启 amuled 生效

aMule 的 `server.met` 会随 ED2K 服务器增减而过期，需要周期性从上游 `http://upd.emule-security.org/server.met` 刷新。aMule 自带 `AutoUpdateServerListAtStartup` 只在启动时拉取、且 macOS daemon 模式下不可靠；运行中的 `amuled` 又不热重载磁盘上的 `server.met`。决定在 Go worker 内新增一个 asynq 周期任务（默认每日 01:17 本地时，`ED2K_SERVERLIST_REFRESH_CRON` 可配）驱动刷新，实际“下载 server.met 落盘 + 重启 amuled”由独立执行器脚本 `scripts/ed2k-serverlist-refresh.sh` 完成，Go 侧不写 `~/.aMule/server.met`、不调 `amulecmd`，延续 [[ED2K aMule 执行器契约]] 把 aMule 远控细节拦在 Go 外的边界。

## 生效方式：重启 amuled，而非 amulecmd EC reload

刷新脚本下载 `server.met` 落盘后，用 `launchctl kickstart -k gui/$(id -u)/com.aivideo.amuled` 重启 amuled 让它在启动时重读，不依赖 `amulecmd` 的 EC reload 动作命令。原因是部署机上 `amulecmd` 的动作命令历史上不可靠（见 [[ED2K aMule 3.0.0 事件循环修复]]：2.3.3 时期 `status`/`show dl`/`connect` 静默超时，只有 `help` 返回；3.0.0 已恢复但脚本契约仍不依赖动作命令以保稳定）。重启虽会中断在途下载轮询，但 aMule 的 `.part`/`.part.met` 持久化保证重启后断点续传，配合刷新前的在途任务跳过判定，对家用深夜场景影响可控。

## 被拒绝的替代方案

- **aMule `AutoUpdateServerListAtStartup`**：只在 amuled 启动时触发，无法周期刷新；且 macOS daemon 模式下定时拉取行为不可观测、不可控，与项目“Go 调度为唯一可观测入口”的方向冲突。
- **`amulecmd` EC reload**：不重启 amuled、不中断下载，最优雅；但依赖部署机上不可靠的 amulecmd 动作命令，风险高于重启。
- **Go 直接下载 + 调 amulecmd**：省一个脚本，但把 aMule 路径细节引入 Go，违反 [[ED2K 外部执行器适配]] 边界。
- **部署机 launchd cron + curl**：完全不进 Go，但不可观测、运维细节散到部署机、无法与下载任务的在途跳过判定联动。
- **只下载 server.met 不重启**：amuled 运行中不重读磁盘文件，新列表不生效，刷新“成功”却无实际效果，语义不诚实。

## 在途任务跳过判定

重启 amuled 会中断在途下载轮询，可能让在跑的下载任务超时落 `failed`。所以刷新任务执行前先查 `ed2k_download_tasks` 是否有 `status IN ('queued','running')` 的任务，任一存在就跳过本轮（写日志不报错），等下一周期。`queued` 也算在途，覆盖 queued 随时转 running 的窗口。口径是“保护在途下载”优先于“列表及时更新”——列表过期只影响下载可用性下降，不破坏已有任务。
