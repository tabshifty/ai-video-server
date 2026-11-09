"""独立 Hermes 扩展：保留原 Telegram 适配器，增加确定性下载交互。"""
from .bridge import register

__all__ = ["register"]
