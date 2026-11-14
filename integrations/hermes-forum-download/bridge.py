"""论坛按钮、持久化操作队列及后端回执；不启动第二套 Telegram 消费者。"""
from __future__ import annotations

import asyncio
from contextlib import contextmanager
from dataclasses import replace
import json
import logging
import os
from pathlib import Path
import re
import secrets
import sqlite3
import sys
import time
import urllib.error
import urllib.request
from urllib.parse import urlparse
import uuid

log = logging.getLogger(__name__)
PREFIX = "/api/v1/integrations/telegram-forum"
PAGE_SIZE = 10
LABELS = {"submitted": "已受理", "retried": "已提交重试", "skipped_completed": "已下载，跳过", "skipped_active": "进行中，跳过", "duplicate": "重复资源，跳过", "invalid": "无效链接", "failed": "提交失败"}
STATUSES = {"completed": "已下载", "failed": "失败", "blocked": "需处理", "invalid": "无效链接", "": "未下载"}


class APIError(Exception):
    def __init__(self, status, message):
        super().__init__(message)
        self.status = status


class Client:
    def __init__(self, base, token, owner):
        parsed = urlparse(base)
        if parsed.scheme not in {"http", "https"} or not parsed.hostname or parsed.username or parsed.password or parsed.query or parsed.fragment:
            raise ValueError("论坛下载 api_base 配置无效")
        if not token or int(owner) <= 0:
            raise ValueError("论坛下载独立令牌及本人用户 ID 未配置")
        self.base, self.token, self.owner = base.rstrip("/"), token, int(owner)

    def call(self, method, path, data=None):
        body = None if data is None else json.dumps(data).encode()
        request = urllib.request.Request(self.base + PREFIX + path, data=body, method=method, headers={"Authorization": "Bearer " + self.token, "X-Telegram-User-ID": str(self.owner), "Content-Type": "application/json"})
        # 不跟随跳转，防止把机器令牌发送到非配置目标。
        class NoRedirect(urllib.request.HTTPRedirectHandler):
            def redirect_request(self, *args, **kwargs):
                return None
        try:
            with urllib.request.build_opener(NoRedirect).open(request, timeout=8 if method == "GET" else 35) as response:
                return json.load(response)["data"]
        except urllib.error.HTTPError as exc:
            messages = {400: "请求无效，请重新选择", 401: "下载接口鉴权失败", 403: "无权操作", 404: "帖子或资源已不存在", 409: "资源或操作已变化，请重新选择并确认", 503: "下载服务未启用或暂时不可用"}
            raise APIError(exc.code, messages.get(exc.code, "下载服务暂时不可用")) from None
        except (OSError, ValueError, KeyError):
            raise APIError(503, "下载服务暂时不可用，请稍后重试") from None


class Store:
    def __init__(self, path):
        self.path = Path(path)
        self.path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
        with self.db() as db:
            db.executescript("""
                CREATE TABLE IF NOT EXISTS choices(token TEXT PRIMARY KEY,payload TEXT NOT NULL);
                CREATE TABLE IF NOT EXISTS pending(event_id TEXT PRIMARY KEY,payload TEXT NOT NULL,done INTEGER NOT NULL DEFAULT 0,error TEXT);
                CREATE TABLE IF NOT EXISTS delivered(id TEXT PRIMARY KEY);
            """)
        self.path.chmod(0o600)

    @contextmanager
    def db(self):
        with sqlite3.connect(self.path, timeout=10) as db:
            yield db

    def choice(self, payload):
        token = secrets.token_hex(12)
        with self.db() as db:
            db.execute("INSERT INTO choices VALUES (?,?)", (token, json.dumps(payload)))
        return token

    def bind(self, token, chat, message):
        with self.db() as db:
            row = db.execute("SELECT payload FROM choices WHERE token=?", (token,)).fetchone()
            payload = json.loads(row[0])
            payload.update(chat_id=chat, message_id=message)
            db.execute("UPDATE choices SET payload=? WHERE token=?", (json.dumps(payload), token))

    def get(self, token, chat, message):
        with self.db() as db:
            row = db.execute("SELECT payload FROM choices WHERE token=?", (token,)).fetchone()
        if row is None:
            raise ValueError("按钮已失效，请查看新的资源推送")
        data = json.loads(row[0])
        if data.get("chat_id") != chat or data.get("message_id") != message:
            raise ValueError("按钮与原推送不匹配")
        return data

    def queue(self, request):
        with self.db() as db:
            db.execute("INSERT OR IGNORE INTO pending(event_id,payload) VALUES (?,?)", (request["event_id"], json.dumps(request)))

    def pending(self):
        with self.db() as db:
            return [(event, json.loads(payload)) for event, payload in db.execute("SELECT event_id,payload FROM pending WHERE done=0 ORDER BY rowid LIMIT 5")]

    def finish(self, event, error=None):
        with self.db() as db:
            db.execute("UPDATE pending SET done=1,error=? WHERE event_id=?", (error, event))

    def delivered(self, receipt):
        with self.db() as db:
            return db.execute("SELECT 1 FROM delivered WHERE id=?", (receipt,)).fetchone() is not None

    def mark_delivered(self, receipt):
        with self.db() as db:
            db.execute("INSERT OR IGNORE INTO delivered VALUES (?)", (receipt,))


def configured(ctx):
    from hermes_cli.config import get_hermes_home
    client = Client(str(ctx.get_config("api_base", "")), os.environ.get("TELEGRAM_FORUM_DOWNLOAD_TOKEN", "").strip(), ctx.get_config("owner_id", 0))
    return client, Store(get_hermes_home() / "plugin-data" / "forum-download" / "state.sqlite3")


def short(value, length=100):
    return " ".join(str(value).split()).encode("utf-16-le")[:length * 2].decode("utf-16-le", errors="ignore")


def size_text(value):
    if value is None:
        return "大小未知"
    for unit in ("B", "KB", "MB", "GB", "TB"):
        if value < 1024 or unit == "TB":
            return f"{value:.1f} {unit}"
        value /= 1024


def photo_caption(post, text):
    """短正文原样附图；长正文保留标题、完整帖子链接和资源概况。"""
    if len(text.encode("utf-16-le")) // 2 <= 1024:
        return text
    match = re.search(r"\[打开帖子\]\((https?://[^\s)]+)\)", text)
    link = match.group(1) if match else ""
    summary = f"共 {len(post['resources'])} 条 ED2K；附件请打开帖子查看。\n点击下方按钮查看下载资源。"
    tail = "\n" + (link + "\n" if link else "") + summary
    budget = 1024 - len(tail.encode("utf-16-le")) // 2
    if budget < 1:
        raise ValueError("帖子链接过长，无法生成图片说明")
    return short(post["title"], min(200, budget)) + tail


def button(store, payload, label):
    from telegram import InlineKeyboardButton
    return InlineKeyboardButton(label, callback_data="fd:" + store.choice(payload))


def render(post, store, chat, message, thread=None, page=0, selected=None):
    """每个按钮保存固定资源快照和消息绑定，不信任回调携带的 URL。"""
    from telegram import InlineKeyboardMarkup
    resources = post["resources"]
    selected = set(selected or [])
    pages = max(1, (len(resources) + PAGE_SIZE - 1) // PAGE_SIZE)
    page = max(0, min(page, pages - 1))
    common = dict(post_id=post["id"], fingerprint=post["fingerprint"], chat_id=chat, message_id=message, thread_id=thread)
    rows = []
    lines = [short(post["title"], 160), f"共 {len(resources)} 条 ED2K，第 {page + 1}/{pages} 页；已选 {len(selected)}/50"]
    for index, resource in enumerate(resources[page * PAGE_SIZE:(page + 1) * PAGE_SIZE], page * PAGE_SIZE + 1):
        status = STATUSES.get(resource["status"], "进行中")
        lines.append(f"{index}. {short(resource['name'], 100)} · {size_text(resource['size'])} · {status}")
        if resource["status"] == "invalid":
            continue
        payload = dict(common, action="submit", resource_ids=[resource["id"]], all=False, retry_attempts={resource["id"]: resource["retry_count"]} if resource["status"] in {"failed", "blocked"} else {})
        action = "重试" if resource["status"] in {"failed", "blocked"} else "下载"
        if resource["status"] in {"", "failed", "blocked"}:
            toggled = selected ^ {resource["id"]}
            toggle = button(store, dict(common, action="list", page=page, selected=sorted(toggled)), ("取消选择 " if resource["id"] in selected else "选择 ") + str(index))
            rows.append([button(store, payload, f"{action} {index}"), toggle])
    navigation = []
    for offset, label in ((-1, "上一页"), (1, "下一页")):
        if 0 <= page + offset < pages:
            navigation.append(button(store, dict(common, action="list", page=page + offset, selected=sorted(selected)), label))
    if navigation:
        rows.append(navigation)
    if selected:
        rows.append([button(store, dict(common, action="confirm", resource_ids=sorted(selected), all=False), "确认下载选中资源")])
    if 0 < len(resources) <= 50:
        rows.append([button(store, dict(common, action="confirm", resource_ids=[], all=True), "全部下载（需确认）")])
    elif len(resources) > 50:
        lines.append("整帖超过 50 条，请跨页选择，每批最多 50 条。")
    rows.append([button(store, dict(common, action="list", page=page, selected=[]), "刷新状态")])
    return "\n".join(lines), InlineKeyboardMarkup(rows)


def receipt_text(receipt):
    payload = receipt["payload"]
    if receipt["key"] == "accepted":
        counts = {"受理": 0, "跳过": 0, "失败": 0}
        lines = []
        for item in payload["items"]:
            outcome = item["outcome"]
            counts["受理" if outcome in {"submitted", "retried"} else "跳过" if outcome in {"duplicate", "skipped_active", "skipped_completed"} else "失败"] += 1
            task = f" · 任务 #{item['task_id']}" if item.get("task_id") else ""
            detail = " · " + short(item["error"], 14) if item.get("error") else ""
            lines.append(f"{short(item['name'], 22)}：{LABELS[outcome]}{task}{detail}")
        return f"下载操作 {payload['id']}\n受理 {counts['受理']}，跳过 {counts['跳过']}，失败 {counts['失败']}\n" + "\n".join(lines)
    labels = {"completed": "已下载到服务器本地", "failed": "下载失败，可刷新资源列表后重试", "blocked": "下载需人工处理，可处理后刷新并重试", "superseded": "原尝试已被新的重试替代，请在网站查看当前状态"}
    reason = {"login_required": "115 需要登录", "quota_exhausted": "115 配额不足", "browser_down": "115 浏览器不可用", "needs_attention": "需要人工处理"}.get(payload.get("blocked_reason"), "")
    return f"任务 #{payload['task_id']} · {short(payload.get('name', ''), 160)}\n{labels.get(payload['status'], '请查看网站任务状态')}" + (f"\n{reason}" if reason else "")


class Bridge:
    def __init__(self, client, store):
        self.client, self.store = client, store
        self.wake = asyncio.Event()

    async def call(self, method, path, data=None):
        return await asyncio.to_thread(self.client.call, method, path, data)

    async def handle(self, query):
        if query.from_user.id != self.client.owner:
            await query.answer("无权操作", show_alert=True)
            return
        await query.answer()
        try:
            message = query.message
            chat, message_id = message.chat_id, message.message_id
            key = query.data.removeprefix("fd:")
            choice = self.store.get(key, chat, message_id)
            post = await self.call("GET", "/posts/" + choice["post_id"])
            action = choice["action"]
            changed = post["fingerprint"] != choice["fingerprint"]
            if action == "open" or (choice.get("direct") and changed):
                # 先取得回复消息 ID，再将面板按钮绑定到回复，保留原始图文。
                # 显式 None 防止 PTB 从 General 话题原消息继承无效的 thread_id=1。
                panel = await message.reply_text("正在加载下载资源…", reply_to_message_id=message_id, message_thread_id=thread_args(choice.get("thread_id")).get("message_thread_id"))
                text, markup = render(post, self.store, panel.chat_id, panel.message_id, choice.get("thread_id"))
                if changed:
                    text = "资源已变化，请重新选择并确认。\n" + text
                await panel.edit_text(text, reply_markup=markup)
                return
            if post["fingerprint"] != choice["fingerprint"]:
                if action == "list":
                    choice["selected"] = []
                else:
                    text, markup = render(post, self.store, chat, message_id, choice.get("thread_id"))
                    await query.edit_message_text("资源已变化，请重新选择并确认。\n" + text, reply_markup=markup)
                    return
            if action == "list":
                if len(choice.get("selected", [])) > 50:
                    await message.reply_text("每批最多选择 50 条资源。")
                    return
                text, markup = render(post, self.store, chat, message_id, choice.get("thread_id"), choice.get("page", 0), choice.get("selected"))
                await query.edit_message_text(text, reply_markup=markup)
            elif action == "confirm":
                from telegram import InlineKeyboardMarkup
                resources = post["resources"] if choice["all"] else [r for r in post["resources"] if r["id"] in choice["resource_ids"]]
                if not resources or len(resources) > 50:
                    raise ValueError("每次只能确认 1 至 50 条资源")
                total = sum(r["size"] or 0 for r in resources)
                unknown = sum(r["size"] is None for r in resources)
                retries = sum(r["status"] in {"failed", "blocked"} for r in resources)
                skips = sum(r["status"] not in {"", "failed", "blocked", "invalid"} for r in resources)
                text = f"确认下载 {len(resources)} 条 ED2K？\n声明总大小：{size_text(total)}" + (f"，其中 {unknown} 条大小未知" if unknown else "") + f"\n将重试旧失败任务 {retries} 条，跳过已下载或进行中 {skips} 条。\n单项失败不影响其它资源。"
                retry_attempts = {r["id"]: r["retry_count"] for r in resources if r["status"] in {"failed", "blocked"}}
                rows = [[button(self.store, dict(choice, action="submit", retry_attempts=retry_attempts), "确认提交")], [button(self.store, dict(choice, action="list", selected=[], page=0), "返回资源列表")]]
                await query.edit_message_text(text, reply_markup=InlineKeyboardMarkup(rows))
            elif action == "submit":
                request = {k: choice[k] for k in ("chat_id", "message_id", "thread_id", "post_id", "fingerprint", "resource_ids", "all")}
                request.update(event_id="choice:" + key, actor_id=query.from_user.id, retry_attempts=choice.get("retry_attempts", {}))
                self.store.queue(request)
                self.wake.set()
                if choice.get("direct"):
                    await message.reply_text("操作已保存，受理结果将回复到此会话。", reply_to_message_id=message_id, message_thread_id=thread_args(choice.get("thread_id")).get("message_thread_id"))
                    return
                # 任务受理回执经持久化通道发送；刷新后的按钮提供显式重试。
                text, markup = render(post, self.store, chat, message_id, choice.get("thread_id"))
                await query.edit_message_text(text + "\n操作已保存，受理结果将回复到此会话。", reply_markup=markup)
            else:
                raise ValueError("按钮操作无效")
        except (ValueError, APIError) as exc:
            await query.message.reply_text(str(exc))
        except Exception:
            # Telegram 的异常可能包含 token URL，不把异常原文写入日志或消息。
            log.warning("论坛下载交互失败；已持久化的操作将继续恢复")

    async def flush_operations(self):
        for event, request in self.store.pending():
            try:
                await self.call("POST", "/operations", request)
                self.store.finish(event)
            except APIError as exc:
                if exc.status in {400, 403, 404, 409}:
                    # 留下持久化拒绝结果，稍后由同会话回执通知用户。
                    self.store.finish(event, str(exc))
                else:
                    break

    async def pump(self, bot):
        await self.flush_operations()
        # 接口拒绝不是已受理下载；仍需向原会话反馈。
        with self.store.db() as db:
            rejected = list(db.execute("SELECT event_id,payload,error FROM pending WHERE done=1 AND error IS NOT NULL"))
        for event, raw, error in rejected:
            request = json.loads(raw)
            await bot.send_message(chat_id=request["chat_id"], text=error, **thread_args(request.get("thread_id")))
            with self.store.db() as db:
                db.execute("UPDATE pending SET error=NULL WHERE event_id=?", (event,))
        receipts = await self.call("GET", "/receipts")
        for receipt in receipts:
            try:
                if not self.store.delivered(receipt["id"]):
                    text = receipt_text(receipt)
                    # 50 条结果的文件名有长度上限，确保回执能一次发送。
                    await bot.send_message(chat_id=receipt["chat_id"], text=text, **thread_args(receipt.get("thread_id")))
                    self.store.mark_delivered(receipt["id"])
                await self.call("POST", f"/receipts/{receipt['id']}/ack", {"token": receipt["token"]})
            except Exception:
                log.warning("论坛下载回执发送或确认失败，租约过期后重试")

    async def run(self, bot):
        while True:
            self.wake.clear()
            try:
                await self.pump(bot)
            except asyncio.CancelledError:
                raise
            except Exception:
                log.warning("论坛下载后台恢复暂不可用，稍后重试")
            try:
                await asyncio.wait_for(self.wake.wait(), timeout=5)
            except asyncio.TimeoutError:
                pass


def thread_args(thread):
    return {"message_thread_id": int(thread)} if thread and int(thread) != 1 else {}


def adapter_factory(client, store):
    from plugins.platforms.telegram.adapter import TelegramAdapter, _resolve_notifications_mode
    from telegram.ext import ApplicationHandlerStop, CallbackQueryHandler

    class ForumTelegramAdapter(TelegramAdapter):
        def __init__(self, config):
            super().__init__(config)
            self.forum_bridge = Bridge(client, store)
            self.forum_task = None
            self._notifications_mode = _resolve_notifications_mode()

        def _register_handlers(self, app):
            super()._register_handlers(app)

            async def callback(update, context):
                await self.forum_bridge.handle(update.callback_query)
                raise ApplicationHandlerStop
            app.add_handler(CallbackQueryHandler(callback, pattern=r"^fd:"), group=-1)

        async def connect(self, **kwargs):
            connected = await super().connect(**kwargs)
            if connected and (self.forum_task is None or self.forum_task.done()):
                self.forum_task = asyncio.create_task(self.forum_bridge.run(self._app.bot))
            return connected

        async def disconnect(self):
            if self.forum_task is not None:
                self.forum_task.cancel()
                try:
                    await self.forum_task
                except asyncio.CancelledError:
                    pass
                self.forum_task = None
            await super().disconnect()

    return ForumTelegramAdapter


def register(ctx):
    client, store = configured(ctx)
    from gateway.platform_registry import platform_registry
    original = platform_registry.get("telegram")
    if original is None:
        raise RuntimeError("请先启用 Hermes Telegram 平台")
    # 官方 registry 支持平台覆盖；替换工厂，保留原 sender、配置及授权元数据。
    platform_registry.register(replace(original, adapter_factory=adapter_factory(client, store)))
    ctx.register_cli_command("forum-download-send", "推送带论坛下载按钮的资源消息", setup_send, lambda args: command_send(args, client, store))


def setup_send(parser):
    parser.add_argument("--tid", required=True)
    parser.add_argument("--to", default="telegram")
    parser.add_argument("--file", default="-")
    parser.add_argument("--json", action="store_true")


def command_send(args, client, store):
    try:
        from hermes_cli.send_cmd import _load_hermes_env
        _load_hermes_env()
        message = sys.stdin.read() if args.file == "-" else Path(args.file).read_text(encoding="utf-8")
        result = asyncio.run(send_post(args, message, client, store))
        print(json.dumps(result, ensure_ascii=False))
        return 0
    except Exception:
        print(json.dumps({"success": False, "error": "论坛消息发送失败，请检查下载接口与扩展配置"}, ensure_ascii=False))
        raise SystemExit(1)


async def send_post(args, message, client, store):
    from gateway.config import Platform, load_gateway_config
    from gateway.platforms.base import BasePlatformAdapter, utf16_len
    from telegram import Bot, InlineKeyboardButton, InlineKeyboardMarkup
    from telegram.request import HTTPXRequest
    from tools.send_message_tool import _parse_target_ref
    from gateway.channel_directory import resolve_channel_name
    from gateway.platforms.base import resolve_proxy_url
    post = await asyncio.to_thread(client.call, "GET", "/posts?tid=" + str(int(args.tid)))
    config = load_gateway_config()
    pc = config.platforms[Platform.TELEGRAM]
    if args.to == "telegram":
        if pc.home_channel is None:
            raise ValueError("Telegram 默认会话未配置")
        chat, thread = pc.home_channel.chat_id, pc.home_channel.thread_id
    elif args.to.startswith("telegram:"):
        ref = args.to.split(":", 1)[1]
        chat, thread, explicit = _parse_target_ref("telegram", ref)
        if not explicit:
            chat = resolve_channel_name("telegram", ref)
        if not chat:
            raise ValueError("Telegram 目标不可用")
    else:
        raise ValueError("论坛下载只支持 Telegram 推送")
    media, text = BasePlatformAdapter.extract_media(message)
    proxy = resolve_proxy_url("TELEGRAM_PROXY", target_hosts=["api.telegram.org"])
    bot = Bot(token=pc.token, request=HTTPXRequest(proxy=proxy))
    async with bot:
        common = dict(action="open", post_id=post["id"], fingerprint=post["fingerprint"], thread_id=int(thread) if thread else None)
        resources = post["resources"]
        label = f"选择下载（{len(resources)}）"
        if len(resources) == 1:
            resource = resources[0]
            label = "查看下载状态"
            if resource["status"] in {"", "failed", "blocked"}:
                common.update(action="submit", direct=True, all=False, resource_ids=[resource["id"]], retry_attempts={resource["id"]: resource["retry_count"]} if resource["status"] in {"failed", "blocked"} else {})
                label = "重试下载" if resource["status"] else "下载"
        token = store.choice(common) if resources else None
        markup = InlineKeyboardMarkup([[InlineKeyboardButton(label, callback_data="fd:" + token)]]) if token else None
        last = None
        if len(media) == 1:
            caption = photo_caption(post, text)
            try:
                with open(media[0][0], "rb") as photo:
                    last = await bot.send_photo(chat_id=chat, photo=photo, caption=caption, **thread_args(thread))
            except Exception:
                # 不记录可能携带 Bot Token 的异常原文；失败时保留完整文字。
                log.warning("论坛图片发送失败，降级纯文字推送")
        if last is None:
            chunks = BasePlatformAdapter.truncate_message(text, 3500, len_fn=utf16_len)
            for chunk in chunks:
                last = await bot.send_message(chat_id=chat, text=chunk, **thread_args(thread))
        if last is None:
            raise ValueError("论坛文字为空")
        if token:
            store.bind(token, last.chat_id, last.message_id)
            await bot.edit_message_reply_markup(chat_id=last.chat_id, message_id=last.message_id, reply_markup=markup)
        return {"success": True, "message_id": str(last.message_id), "post_id": post["id"]}
