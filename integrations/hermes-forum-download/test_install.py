"""在临时 Hermes 目录验证真实补丁、幂等安装及回滚备份。"""
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


class InstallTests(unittest.TestCase):
    def test_install_is_repeatable_and_preserves_original_script(self):
        source = Path.home() / ".hermes/scripts/watch_sehuatang_forum95.py"
        if not source.is_file():
            self.skipTest("本机没有可供验证的现有监控脚本")
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            script = home / "scripts/watch_sehuatang_forum95.py"
            script.parent.mkdir()
            shutil.copy2(source, script)
            original = script.read_bytes()
            installer = Path(__file__).with_name("install.py")
            for arguments in (("--check",), (), ()):
                result = subprocess.run([sys.executable, str(installer), "--home", temp, *arguments], capture_output=True, text=True, timeout=10)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual(script.with_suffix(".py.before-forum-download").read_bytes(), original)
            self.assertEqual(script.read_text().count('command[1] = "forum-download-send"'), 1)
            self.assertTrue((home / "plugins/forum-download/bridge.py").is_file())
            compile(script.read_text(), str(script), "exec")


if __name__ == "__main__":
    unittest.main()
