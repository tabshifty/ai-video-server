"""离线行为验证：无需真实 Bot 或生产下载。"""
import asyncio
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import AsyncMock
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
