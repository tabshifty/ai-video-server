"""离线行为验证：无需真实 Bot 或生产下载。"""
import asyncio
from datetime import datetime, timezone
import importlib.util
import json
from pathlib import Path
import tempfile
import sys
from types import SimpleNamespace
import unittest
from unittest.mock import AsyncMock, Mock, patch
import uuid

spec = importlib.util.spec_from_file_location("forum_bridge_test_subject", Path(__file__).with_name("bridge.py"))
bridge = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bridge)


def post(count):
    return {"id": str(uuid.uuid4()), "fingerprint": "snapshot", "title": "测试帖子", "resources": [{"id": str(uuid.uuid4()), "name": f"资源{i}.zip", "size": 123, "status": "", "task_id": None, "retry_count": 0} for i in range(count)]}


class BridgeTests(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.store = bridge.Store(Path(self.temp.name) / "state.db")

    async def send_post(self, text, media=True, photo_error=None, count=2, status=""):
        p = post(count)
        for resource in p["resources"]:
            resource.update(status=status, retry_count=2)
        photo = Path(self.temp.name) / "preview.jpg"
        photo.write_bytes(b"offline-image")
        sent = SimpleNamespace(chat_id=11, message_id=22)
        bot = AsyncMock()
        bot.__aenter__.return_value = bot
        bot.send_photo.side_effect = photo_error
        bot.send_photo.return_value = sent
        bot.send_message.return_value = sent
        base = SimpleNamespace(extract_media=Mock(return_value=([(str(photo), False)] if media else [], text)), truncate_message=lambda value, *a, **kw: [value])
        modules = {
            "gateway.config": SimpleNamespace(Platform=SimpleNamespace(TELEGRAM="telegram"), load_gateway_config=lambda: SimpleNamespace(platforms={"telegram": SimpleNamespace(token="123:test", home_channel=SimpleNamespace(chat_id=11, thread_id=3))})),
            "gateway.platforms.base": SimpleNamespace(BasePlatformAdapter=base, utf16_len=lambda value: len(value.encode("utf-16-le")) // 2, resolve_proxy_url=lambda *a, **kw: None),
            "gateway.channel_directory": SimpleNamespace(resolve_channel_name=Mock()),
            "tools.send_message_tool": SimpleNamespace(_parse_target_ref=Mock(), _send_telegram=AsyncMock()),
        }
        with patch.dict(sys.modules, modules), patch("telegram.Bot", return_value=bot):
            result = await bridge.send_post(SimpleNamespace(tid=123, to="telegram"), text, SimpleNamespace(call=Mock(return_value=p)), self.store)
        return p, bot, result

    async def test_single_resource_button_downloads_or_retries_without_opening_panel(self):
        for status, action, label in (("", "submit", "下载"), ("failed", "submit", "重试下载"), ("blocked", "submit", "重试下载"), ("completed", "open", "查看下载状态"), ("invalid", "open", "查看下载状态"), ("downloading", "open", "查看下载状态")):
            with self.subTest(status=status):
                p, bot, _ = await self.send_post("单资源", count=1, status=status)
                button = bot.edit_message_reply_markup.call_args.kwargs["reply_markup"].inline_keyboard[0][0]
                self.assertEqual(button.text, label)
                choice = self.store.get(button.callback_data[3:], 11, 22)
                self.assertEqual(choice["action"], action)
                if action == "submit":
                    self.assertTrue(choice["direct"])
                    self.assertEqual(choice["resource_ids"], [p["resources"][0]["id"]])
                    self.assertEqual(choice["retry_attempts"], {p["resources"][0]["id"]: 2} if status else {})

    async def test_direct_download_queues_once_without_editing_photo(self):
        p, bot, _ = await self.send_post("单资源", count=1)
        button = bot.edit_message_reply_markup.call_args.kwargs["reply_markup"].inline_keyboard[0][0]
        message = SimpleNamespace(chat_id=11, message_id=22, reply_text=AsyncMock())
        query = SimpleNamespace(from_user=SimpleNamespace(id=123), answer=AsyncMock(), message=message, data=button.callback_data, edit_message_text=AsyncMock())
        subject = bridge.Bridge(SimpleNamespace(owner=123), self.store)
        subject.call = AsyncMock(return_value=p)
        await subject.handle(query)
        await subject.handle(query)
        self.assertEqual(len(self.store.pending()), 1)
        self.assertTrue(subject.wake.is_set())
        query.edit_message_text.assert_not_awaited()
        self.assertIn("操作已保存", message.reply_text.call_args.args[0])

    async def test_changed_direct_resource_opens_panel_without_submitting(self):
        p, bot, _ = await self.send_post("单资源", count=1)
        button = bot.edit_message_reply_markup.call_args.kwargs["reply_markup"].inline_keyboard[0][0]
        p["fingerprint"] = "changed"
        panel = SimpleNamespace(chat_id=11, message_id=33, edit_text=AsyncMock())
        message = SimpleNamespace(chat_id=11, message_id=22, reply_text=AsyncMock(return_value=panel))
        query = SimpleNamespace(from_user=SimpleNamespace(id=123), answer=AsyncMock(), message=message, data=button.callback_data, edit_message_text=AsyncMock())
        subject = bridge.Bridge(SimpleNamespace(owner=123), self.store)
        subject.call = AsyncMock(return_value=p)
        await subject.handle(query)
        self.assertEqual(self.store.pending(), [])
        query.edit_message_text.assert_not_awaited()
        panel.edit_text.assert_awaited_once()
        self.assertIn("资源已变化", panel.edit_text.call_args.args[0])

    async def test_photo_and_short_text_are_one_message_with_bound_button(self):
        text = "测试帖子\nhttps://example.invalid/post"
        p, bot, result = await self.send_post(text)
        bot.send_photo.assert_awaited_once()
        self.assertEqual(bot.send_photo.call_args.kwargs["caption"], text)
        self.assertEqual(bot.send_photo.call_args.kwargs["message_thread_id"], 3)
        bot.send_message.assert_not_awaited()
        self.assertTrue(result["success"])
        markup = bot.edit_message_reply_markup.call_args.kwargs["reply_markup"]
        choice = self.store.get(markup.inline_keyboard[0][0].callback_data[3:], 11, 22)
        self.assertEqual(choice["action"], "open")
        self.assertEqual(choice["post_id"], p["id"])

    async def test_long_photo_caption_preserves_post_link_and_full_resource_count(self):
        text = "[打开帖子](https://example.invalid/post)\n" + "😀" * 1000
        _, bot, _ = await self.send_post(text, count=51)
        caption = bot.send_photo.call_args.kwargs["caption"]
        self.assertLessEqual(len(caption.encode("utf-16-le")) // 2, 1024)
        self.assertIn("https://example.invalid/post", caption)
        self.assertIn("51", caption)
        bot.send_message.assert_not_awaited()

    async def test_missing_or_failed_photo_falls_back_to_original_text_with_button(self):
        for media, error in ((False, None), (True, RuntimeError("offline failure"))):
            with self.subTest(media=media):
                _, bot, result = await self.send_post("完整原文", media=media, photo_error=error)
                self.assertEqual(bot.send_message.call_args.kwargs["text"], "完整原文")
                bot.edit_message_reply_markup.assert_awaited_once()
                self.assertTrue(result["success"])

    async def test_photo_button_opens_fresh_reply_panel_without_editing_original(self):
        p = post(51)
        token = self.store.choice(dict(action="open", post_id=p["id"], fingerprint="old", chat_id=11, message_id=22, thread_id=3))
        panel = SimpleNamespace(chat_id=11, message_id=33, edit_text=AsyncMock())
        message = SimpleNamespace(chat_id=11, message_id=22, reply_text=AsyncMock(return_value=panel))
        query = SimpleNamespace(from_user=SimpleNamespace(id=123), answer=AsyncMock(), message=message, data="fd:" + token, edit_message_text=AsyncMock())
        subject = bridge.Bridge(SimpleNamespace(owner=123), self.store)
        subject.call = AsyncMock(return_value=p)
        await subject.handle(query)
        query.edit_message_text.assert_not_awaited()
        self.assertEqual(message.reply_text.call_args.kwargs["reply_to_message_id"], 22)
        self.assertEqual(message.reply_text.call_args.kwargs["message_thread_id"], 3)
        panel.edit_text.assert_awaited_once()
        self.assertIn("共 51", panel.edit_text.call_args.args[0])
        markup = panel.edit_text.call_args.kwargs["reply_markup"]
        for row in markup.inline_keyboard:
            for item in row:
                choice = self.store.get(item.callback_data[3:], 11, 33)
                self.assertEqual(choice["fingerprint"], p["fingerprint"])
        self.assertEqual(self.store.pending(), [])
        # 面板中的下载请求关联回复消息，不能误绑原图文。
        query.message = panel
        query.data = markup.inline_keyboard[0][0].callback_data
        await subject.handle(query)
        pending = self.store.pending()
        self.assertEqual(len(pending), 1)
        self.assertEqual(pending[0][1]["message_id"], 33)
        self.assertEqual(pending[0][1]["thread_id"], 3)

    async def test_general_topic_panel_does_not_inherit_thread_one(self):
        from telegram import Chat, Message
        p = post(1)
        token = self.store.choice(dict(action="open", post_id=p["id"], fingerprint=p["fingerprint"], chat_id=11, message_id=22, thread_id=1))
        panel = SimpleNamespace(chat_id=11, message_id=33, edit_text=AsyncMock())
        bot = SimpleNamespace(send_message=AsyncMock(return_value=panel))
        message = Message(message_id=22, date=datetime.now(timezone.utc), chat=Chat(11, "supergroup"), is_topic_message=True, message_thread_id=1)
        message.set_bot(bot)
        query = SimpleNamespace(from_user=SimpleNamespace(id=123), answer=AsyncMock(), message=message, data="fd:" + token)
        subject = bridge.Bridge(SimpleNamespace(owner=123), self.store)
        subject.call = AsyncMock(return_value=p)
        await subject.handle(query)
        bot.send_message.assert_awaited_once()
        self.assertIsNone(bot.send_message.call_args.kwargs["message_thread_id"])

    async def test_unauthorized_callback_never_reads_resources(self):
        client = type("Client", (), {"owner": 123})()
        subject = bridge.Bridge(client, self.store)
        subject.call = AsyncMock()
        query = type("Query", (), {"from_user": type("User", (), {"id": 456})(), "answer": AsyncMock()})()
        await subject.handle(query)
        subject.call.assert_not_called()
        query.answer.assert_awaited_once_with("无权操作", show_alert=True)

    async def test_callback_requests_survive_restart_and_temporary_failure(self):
        client = type("Client", (), {"owner": 123})()
        subject = bridge.Bridge(client, self.store)
        request = {"event_id": "same-event"}
        self.store.queue(request)
        self.store.queue(request)
        subject.call = AsyncMock(side_effect=bridge.APIError(503, "temporary"))
        await subject.flush_operations()
        restarted = bridge.Store(self.store.path)
        self.assertEqual(restarted.pending(), [("same-event", request)])
        subject.call = AsyncMock(return_value={"id": "operation"})
        await subject.flush_operations()
        self.assertEqual(restarted.pending(), [])

    async def test_ack_failure_does_not_resend_after_local_confirmation(self):
        client = type("Client", (), {"owner": 123})()
        subject = bridge.Bridge(client, self.store)
        receipt = {"id": "receipt", "token": "lease", "key": "terminal:1", "chat_id": 2, "thread_id": 3, "payload": {"task_id": 1, "name": "文件", "status": "completed"}}
        subject.call = AsyncMock(side_effect=[[receipt], bridge.APIError(503, "temporary"), [receipt], {}])
        bot = type("Bot", (), {"send_message": AsyncMock()})()
        await subject.pump(bot)
        await subject.pump(bot)
        self.assertEqual(bot.send_message.await_count, 1)
        self.assertEqual(bot.send_message.call_args.kwargs["message_thread_id"], 3)

    async def test_send_failure_keeps_receipt_retryable(self):
        subject = bridge.Bridge(type("Client", (), {"owner": 123})(), self.store)
        receipt = {"id": "receipt", "token": "lease", "key": "terminal:1", "chat_id": 2, "thread_id": None, "payload": {"task_id": 1, "status": "failed"}}
        subject.call = AsyncMock(return_value=[receipt])
        bot = type("Bot", (), {"send_message": AsyncMock(side_effect=RuntimeError("network"))})()
        await subject.pump(bot)
        self.assertFalse(self.store.delivered("receipt"))
        self.assertEqual(subject.call.await_count, 1)

    async def test_pagination_and_full_post_boundary(self):
        p = post(51)
        text, markup = bridge.render(p, self.store, 11, 22)
        self.assertIn("共 51", text)
        self.assertIn("每批最多 50", text)
        self.assertNotIn("11. 资源10", text)
        labels = [b.text for row in markup.inline_keyboard for b in row]
        self.assertFalse(any("全部下载" in label for label in labels))
        text, _ = bridge.render(p, self.store, 11, 22, page=1)
        self.assertIn("11. 资源10", text)
        for row in markup.inline_keyboard:
            for button in row:
                self.assertLessEqual(len(button.callback_data.encode()), 64)
                self.store.get(button.callback_data[3:], 11, 22)
                with self.assertRaises(ValueError):
                    self.store.get(button.callback_data[3:], 11, 33)

    async def test_submit_callback_only_persists_and_wakes_background(self):
        p = post(1)
        payload = dict(action="submit", post_id=p["id"], fingerprint=p["fingerprint"], chat_id=11, message_id=22, thread_id=None, resource_ids=[p["resources"][0]["id"]], all=False)
        token = self.store.choice(payload)
        message = type("Message", (), {"chat_id": 11, "message_id": 22, "reply_text": AsyncMock()})()
        query = type("Query", (), {"from_user": type("User", (), {"id": 123})(), "answer": AsyncMock(), "message": message, "data": "fd:" + token, "edit_message_text": AsyncMock()})()
        subject = bridge.Bridge(type("Client", (), {"owner": 123})(), self.store)
        subject.call = AsyncMock(return_value=p)
        await subject.handle(query)
        self.assertEqual(subject.call.await_count, 1)
        self.assertEqual(subject.call.call_args.args[0], "GET")
        self.assertEqual(len(self.store.pending()), 1)
        self.assertTrue(subject.wake.is_set())
        await subject.handle(query)
        self.assertEqual(len(self.store.pending()), 1)

    async def test_resource_change_can_refresh_and_cannot_submit_old_selection(self):
        p = post(2)
        payload = dict(action="list", post_id=p["id"], fingerprint="old", chat_id=11, message_id=22, thread_id=None, selected=["removed"], page=0)
        message = type("Message", (), {"chat_id": 11, "message_id": 22, "reply_text": AsyncMock()})()
        subject = bridge.Bridge(type("Client", (), {"owner": 123})(), self.store)
        subject.call = AsyncMock(return_value=p)
        for action in ("list", "submit"):
            token = self.store.choice(dict(payload, action=action))
            query = type("Query", (), {"from_user": type("User", (), {"id": 123})(), "answer": AsyncMock(), "message": message, "data": "fd:" + token, "edit_message_text": AsyncMock()})()
            await subject.handle(query)
            query.edit_message_text.assert_awaited_once()
            self.assertIn("已选 0/50", query.edit_message_text.call_args.args[0])
            self.assertEqual(self.store.pending(), [])

    async def test_fifty_long_names_receipt_fits_telegram(self):
        receipt = {"key": "accepted", "payload": {"id": str(uuid.uuid4()), "items": [{"name": "长" * 500, "task_id": 9223372036854775807, "outcome": "submitted"} for _ in range(50)]}}
        text = bridge.receipt_text(receipt)
        self.assertLessEqual(len(text.encode("utf-16-le")) // 2, 4096)
        for item in receipt["payload"]["items"]:
            item.update(name="😀" * 500, outcome="failed", error="失败原因" * 30)
        text = bridge.receipt_text(receipt)
        self.assertLessEqual(len(text.encode("utf-16-le")) // 2, 4096)


if __name__ == "__main__":
    unittest.main()
