#!/usr/bin/env python3
"""安装受版本控制的独立扩展；不编辑凭据、不重启 Hermes。"""
import argparse
from pathlib import Path
import shutil
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description="安装 Hermes 论坛下载扩展与监控脚本补丁")
    parser.add_argument("--home", type=Path, default=Path.home() / ".hermes")
    parser.add_argument("--check", action="store_true", help="仅校验补丁兼容性")
    args = parser.parse_args()
    source = Path(__file__).parent
    script = args.home / "scripts/watch_sehuatang_forum95.py"
    if not script.is_file():
        parser.error("未找到论坛监控脚本")
    already = 'command[1] = "forum-download-send"' in script.read_text(encoding="utf-8")
    if not already:
        result = subprocess.run(["patch", "--dry-run", "--batch", str(script), str(source / "monitor.patch")], capture_output=True)
        if result.returncode:
            print("监控脚本与补丁不兼容；未修改任何文件。", file=sys.stderr)
            return 1
    if args.check:
        print("监控脚本补丁兼容；未修改任何文件。")
        return 0
    backup = script.with_suffix(".py.before-forum-download")
    if not already and backup.exists():
        print("已有回滚备份，拒绝覆盖；请先核对备份。", file=sys.stderr)
        return 1
    destination = args.home / "plugins/forum-download"
    destination.mkdir(parents=True, exist_ok=True)
    for name in ("__init__.py", "bridge.py", "plugin.yaml"):
        shutil.copy2(source / name, destination / name)
    if not already:
        shutil.copy2(script, backup)
        subprocess.run(["patch", "--batch", str(script), str(source / "monitor.patch")], check=True)
    print("扩展与监控补丁已安装。请按接入文档配置、验证，再重启 Hermes；本命令未重启服务。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
