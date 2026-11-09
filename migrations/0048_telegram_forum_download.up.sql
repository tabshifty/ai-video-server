CREATE TABLE telegram_forum_download_operations (
 id UUID PRIMARY KEY,
 event_id TEXT NOT NULL UNIQUE,
 request_hash TEXT NOT NULL,
 actor_id BIGINT NOT NULL CHECK (actor_id > 0),
 chat_id BIGINT NOT NULL CHECK (chat_id <> 0),
 message_id BIGINT NOT NULL CHECK (message_id > 0),
 thread_id BIGINT,
 post_id UUID REFERENCES collected_forum_posts(id) ON DELETE SET NULL,
 result JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE telegram_forum_download_items (
 operation_id UUID NOT NULL REFERENCES telegram_forum_download_operations(id) ON DELETE CASCADE,
 resource_id UUID NOT NULL,
 task_id BIGINT REFERENCES cloud_download_tasks(id) ON DELETE SET NULL,
 attempt INTEGER NOT NULL DEFAULT 0,
 outcome TEXT NOT NULL,
 PRIMARY KEY (operation_id, resource_id)
);
CREATE TABLE telegram_forum_download_receipts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 operation_id UUID NOT NULL REFERENCES telegram_forum_download_operations(id) ON DELETE CASCADE,
 receipt_key TEXT NOT NULL,
 payload JSONB NOT NULL,
 delivered_at TIMESTAMPTZ,
 lease_token UUID,
 lease_until TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE (operation_id, receipt_key)
);
CREATE INDEX telegram_forum_download_receipts_pending ON telegram_forum_download_receipts(created_at) WHERE delivered_at IS NULL;
CREATE INDEX telegram_forum_download_items_task ON telegram_forum_download_items(task_id);
