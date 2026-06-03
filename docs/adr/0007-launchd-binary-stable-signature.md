# launchd 二进制必须使用稳定签名身份

从本 ADR 生效之日起，[[家用部署机]] 的 Go server / worker 在每次构建后都必须用同一份稳定签名身份重新 codesign，再切换 symlink 并重启 launchd。这样做的目的，是让 macOS TCC 对 `/Volumes/large` 的外盘授权绑定到稳定代码身份，而不是绑定到每次 push 都变化的 ad-hoc cdhash。

这个决策是对 [[家用部署机媒体访问权限契约]] 的补充：我们已经确认 server / worker 会从 `/Volumes/large` 读取封面、视频、图片、转码输出等运行期资产；一旦 launchd 运行体失去外盘权限，图片会 503、视频会 pending、worker 会卡在输出目录创建。若二进制继续走 ad-hoc，每次新 build 都可能换一个新的 cdhash，TCC 授权就会在下一次部署后失效，导致“手动能读，launchd 不能读”的故障反复出现。

具体决策：

| 项 | 选择 |
|---|---|
| 签名方式 | ✅ 同一份稳定签名身份（通过 `CODESIGN_IDENTITY` 提供） |
| 允许方式 | ✅ 每次 build 后重新 codesign 新二进制，再切 symlink |
| 禁止方式 | ❌ 继续把 launchd 二进制当 ad-hoc 产物直接部署 |
| 回滚要求 | ✅ 回滚目标二进制也必须走同一份稳定签名身份，避免回退后重新丢失 TCC 授权 |

稳定签名带来的代价是：部署机需要长期保存同一份签名身份，首次切换到这条新链路后要对新的实际二进制重新授予一次外盘访问；但这个成本只发生一次，后续每次 push 都复用同一授权。相比之下，继续 ad-hoc 的代价是每次部署都可能把外盘访问打回原点，维护成本更高。

## 考虑过的替代方案

- **继续 ad-hoc + 每次 push 后手动重新授权**：最省事，但每次部署都要重做系统层授权，故障会反复回归。
- **给整个部署目录做路径级 workaround**：路径不会解决 TCC 按代码身份记忆授权的问题，而且一旦二进制换了位置或被 symlink-swap，仍然会回到原问题。
- **把 server / worker 包成 app bundle 再申请权限**：能让权限管理更像桌面 App，但对当前单机家庭部署过度复杂。
- **不用 launchd，改常驻手工进程**：可以绕开一部分授权体验问题，但会失去自动拉起与标准日志管理，违背 [[家用部署机]] 的基础架构选择。

## 关联

- `docs/adr/0005-home-deployment-architecture.md` — 本 ADR 是 0005 中“launchd 硬切重启”路径的补充约束
- `docs/adr/0006-migration-forward-compatibility.md` — 与 rollback 能力并列生效
- `CONTEXT.md` 新术语：[[家用部署机稳定签名契约]]
- `scripts/sign-launchd-binary.sh` — 部署机重签 helper
- `scripts/rollback.sh` — 回滚时复用同一签名身份
