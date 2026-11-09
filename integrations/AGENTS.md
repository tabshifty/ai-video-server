# 外部集成维护约定

- 本目录维护外部运行时扩展的发布源码，不直接修改 Hermes 核心或提交运行时凭据。
- `hermes-forum-download/` 使用 Hermes 现有 Telegram 平台工厂与独立 CLI 命令，接收仍属于唯一 gateway，不新增长轮询。
- Python 测试：在仓库根目录运行 `~/.hermes/hermes-agent/venv/bin/python -m unittest discover -s integrations/hermes-forum-download -p 'test_*.py' -v`；`HERMES_RUNTIME_PATH` 可覆盖真实插件发现验证的运行时位置。
- 安装前检查：`python3 integrations/hermes-forum-download/install.py --check`；安装、配置及回滚见 `docs/telegram-forum-download-setup.md`。
- 合约属于 `internal/forumdownload/`；后端使用独立 `TELEGRAM_FORUM_DOWNLOAD_TOKEN` 与本人 ID，Hermes 侧行为配置归插件 `settings`，凭据归 `.env`。
- 外部集成独立安装，不由 Go server/worker 的代码推送自动替换 Hermes 运行文件；不得把源码提交等同于已启用生产 Bot。
