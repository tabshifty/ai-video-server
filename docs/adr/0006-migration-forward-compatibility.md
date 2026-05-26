# 新增 migration 必须前向兼容上一版本的二进制

从本 ADR 生效之日起，**每个新增到 `migrations/*.up.sql` 的迁移都必须保证"上一版本的 Go 二进制在新 schema 上仍能正常跑"**。这条契约由 [[migration 前向兼容契约]] 在 CONTEXT.md 中作为术语固化，本 ADR 记录设立这条规则的取舍与理由，供未来 review 与教育新加入者使用。

之所以需要这条规则，是因为 [[家用部署机]] 的 rollback 路径是**手动切回旧 binary**（见 ADR-0005，自动 rollback 不做）。手动 rollback 触发时迁移已经跑过 —— schema 处于新版本。如果旧 binary 的代码假设旧 schema（譬如查询一个已被 rename 的列），rollback 会直接崩溃，rollback 通道本身就废了。前向兼容契约把"binary 可独立回退"作为可达成的目标，让家庭部署在最简单的 supervisor + 手动 rollback 模型下也保持安全。

具体规则：

| 类别 | 允许 | 受限 | 禁止 |
|---|---|---|---|
| 新增表 | ✅ | — | — |
| 新增列 | ✅ NULL 或带 DEFAULT | — | ❌ NOT NULL 且不带 DEFAULT 到非空表 |
| 删除列 | — | ⚠ 必须分两次：先在代码停止使用 → 下次部署再删 | ❌ 一次性 DROP COLUMN |
| 改列类型 | — | ⚠ `ALTER TYPE` + 数据迁移；只允许扩容方向（int → bigint，varchar(10) → varchar(20)）| ❌ DROP + ADD 同名列 |
| 重命名列/表 | — | — | ❌ 任何 RENAME；如需改名走"加新列+双写+迁数据+删旧列"的多次部署 |
| 新增索引 | ✅ | — | — |
| 删除索引 | ✅（只影响性能不影响语义） | — | — |
| 新增枚举值 | ✅ Postgres `ALTER TYPE … ADD VALUE` | — | ❌ 删除枚举值 |

Review 阶段（任务三段流的 review.md）必须对照本规则核对每个新 migration；CI 不强制（家庭场景，靠 review 与本契约 + ADR 的可读性兜底）。

**存量 21 个迁移不回溯审计**，仅约束新增。理由：回溯审计需要识别每个旧 migration 是否违反契约、是否需要补救；这是巨大的考古工作而 ROI 极低（这 21 个迁移在历史上已经被部署过，当时也没有人需要回退）。新规则只对未来 migration 起约束。

## 考虑过的替代方案

- **允许 breaking migration + 实现自动 rollback (含跑 down)**：让 hook 在新 binary 启动失败时自动回滚 binary 并跑对应的 `down.sql`。这要求 down 也必须前向兼容（即新 binary 在跑过 down 之后的旧 schema 上不能死），等于把规则倒置且约束面更广；同时 hook 复杂度大增（需要识别"刚跑过的 up 是哪些"、"对应 down 是否存在且安全"），任何 down 写错都会把数据库带到不可预期状态。家庭场景的可靠性诉求够不到这套机器。
- **严格"原子 deploy"——禁止 rollback**：每个 commit 是单向 forward，schema 与代码必须一起前进，rollback 等同于"反向写一个 commit"。优点：完全没有 rollback 路径意味着没有兼容性负担，可以任意 rename。缺点：代码出严重 bug 时家庭无可用客户端的窗口可能持续到下一个修复 commit 推出来。家庭场景不能容忍这种长尾故障。
- **双 binary 灰度 / feature flag**：保留新旧版本并存，运行时按用户/请求路由。家庭场景没有"多用户灰度"诉求；引入双 binary 立即把所有 stateful 资源（Asynq 队列、DB 连接池、上传目录）的所有权问题暴露出来，复杂度爆炸。
- **event-sourced 或物化视图层做兼容**：在 schema 上面架一层不变接口，业务代码只看接口。这是大型生产系统的典型答案但对单机家庭项目极度过剩。
- **无契约，靠"出问题再说"**：不写本 ADR，每个 migration 写法自由。代价是 rollback 通道随机失效；某次需要回退时才发现"哎呀这个 migration 把列 rename 了"，那时已经晚了。设立契约的成本极低（review 时多看一行），不设立的代价是 rollback 路径不可信。

## 关联

- `docs/adr/0005-home-deployment-architecture.md` — 本 ADR 是 0005 "rollback 不做自动" 决策的支撑条件
- `CONTEXT.md` 新术语：[[migration 前向兼容契约]]
- `migrations/` — 本契约约束的对象；当前 21 个文件不回溯审计
- 未来任务的 review.md checklist 应加入"新 migration 是否符合前向兼容契约"一条
