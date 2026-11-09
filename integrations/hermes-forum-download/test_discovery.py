"""通过 Hermes 真实插件发现与平台工厂验证安装契约。"""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class DiscoveryTests(unittest.TestCase):
    def test_real_discovery_preserves_telegram_sender_and_adds_callback(self):
        runtime = Path(os.environ.get("HERMES_RUNTIME_PATH", Path.home() / ".hermes/hermes-agent"))
        if not (runtime / "hermes_cli/plugins.py").exists():
            self.skipTest("设置 HERMES_RUNTIME_PATH 运行真实发现验证")
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            shutil.copytree(Path(__file__).parent, home / "plugins/forum-download", ignore=shutil.ignore_patterns("__pycache__"))
            (home / "config.yaml").write_text("plugins:\n  enabled: [forum-download]\n  entries:\n    forum-download:\n      settings:\n        api_base: http://127.0.0.1:65530\n        owner_id: 123\n", encoding="utf-8")
            env = dict(os.environ, HERMES_HOME=temp, TELEGRAM_FORUM_DOWNLOAD_TOKEN="offline-test-token", PYTHONPATH=str(runtime))
            code = '''
from hermes_cli.plugins import discover_plugins, get_plugin_manager
from gateway.platform_registry import platform_registry
from gateway.config import PlatformConfig

discover_plugins()
entry = platform_registry.get("telegram")
assert entry is not None
assert entry.standalone_sender_fn is not None
assert "forum-download-send" in get_plugin_manager()._cli_commands
adapter = entry.adapter_factory(PlatformConfig(token="123:test"))
assert hasattr(adapter, "forum_bridge"), type(adapter)
class App:
    def __init__(self): self.handlers=[]
    def add_handler(self, handler, group=0): self.handlers.append((handler,group))
app=App()
adapter._register_handlers(app)
assert any(group==-1 and handler.pattern.pattern=="^fd:" for handler,group in app.handlers)
assert any(group==0 for handler,group in app.handlers)
print("真实插件发现、原 sender 保留及回调注册通过")
'''
            result = subprocess.run([str(runtime / "venv/bin/python"), "-c", code], cwd=temp, env=env, capture_output=True, text=True, timeout=60)
            self.assertEqual(result.returncode, 0, (result.stdout + result.stderr)[-5000:])


if __name__ == "__main__":
    unittest.main()
