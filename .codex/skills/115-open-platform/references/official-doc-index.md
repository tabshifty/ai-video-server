# 115 开放平台官方文档索引

> 来源：[115 开放平台](https://www.yuque.com/115yun/open)。目录核对日期：2026-07-29，当时共 38 篇。下表用于定位官方页面，不代替实时原文。

## 平台与接入

| 文档 | 用途 |
|---|---|
| [概述](https://www.yuque.com/115yun/open/gv0l5007pczskivz) | 平台能力和资源范围 |
| [更新记录](https://www.yuque.com/115yun/open/huv9x8m1opivdokg) | 接口新增和变更记录 |
| [接入流程](https://www.yuque.com/115yun/open/fd7fidbgsritauxm) | 开发者入驻、实名与应用审核 |
| [开发须知](https://www.yuque.com/115yun/open/vq62qwp8ia2efoli) | 隐私、合规、禁止行为和未公开频控 |
| [授权错误码](https://www.yuque.com/115yun/open/rnq0cbz8tt7cu43i) | OAuth 授权错误对照 |
| [开发者商业价值转化](https://www.yuque.com/115yun/open/cguk6qshgapwg4qn) | 增值服务推广与收益 |

## 接入授权

| 文档 | 核心接口 |
|---|---|
| [手机扫码授权 PKCE 模式](https://www.yuque.com/115yun/open/shtpzfhewv5nag11) | `POST /open/authDeviceCode`、`GET https://qrcodeapi.115.com/get/status/`、`POST /open/deviceCodeToToken` |
| [授权码模式](https://www.yuque.com/115yun/open/okr2cq0wywelscpe) | `GET /open/authorize`、`POST /open/authCodeToToken` |
| [刷新 access_token](https://www.yuque.com/115yun/open/opnx8yezo4at2be6) | `POST https://passportapi.115.com/open/refreshToken` |

## 用户管理

| 文档 | 方法与路径 |
|---|---|
| [用户信息](https://www.yuque.com/115yun/open/ot1litggzxa1czww) | `GET /open/user/info` |

## 文件管理

| 文档 | 方法与路径 |
|---|---|
| [获取文件列表](https://www.yuque.com/115yun/open/kz9ft9a7s57ep868) | `GET /open/ufile/files` |
| [文件搜索](https://www.yuque.com/115yun/open/ft2yelxzopusus38) | `GET /open/ufile/search` |
| [获取文件(夹)详情](https://www.yuque.com/115yun/open/rl8zrhe2nag21dfw) | `GET/POST /open/folder/get_info` |
| [新建文件夹](https://www.yuque.com/115yun/open/qur839kyx9cgxpxi) | `POST /open/folder/add` |
| [文件(夹)更新](https://www.yuque.com/115yun/open/gyrpw5a0zc4sengm) | `POST /open/ufile/update` |
| [删除文件](https://www.yuque.com/115yun/open/kt04fu8vcchd2fnb) | `POST /open/ufile/delete` |
| [文件移动](https://www.yuque.com/115yun/open/vc6fhi2mrkenmav2) | `POST /open/ufile/move` |
| [文件复制](https://www.yuque.com/115yun/open/lvas49ar94n47bbk) | `POST /open/ufile/copy` |
| [获取文件下载地址](https://www.yuque.com/115yun/open/um8whr91bxb5997o) | `POST /open/ufile/downurl` |
| [获取上传凭证](https://www.yuque.com/115yun/open/kzacvzl0g7aiyyn4) | `GET /open/upload/get_token` |
| [文件上传](https://www.yuque.com/115yun/open/ul4mrauo5i2uza0q) | `POST /open/upload/init` |
| [断点续传](https://www.yuque.com/115yun/open/tzvi9sbcg59msddz) | `POST /open/upload/resume` |
| [上传流程](https://www.yuque.com/115yun/open/xb89onhdxsfpwsyc) | 秒传、二次认证和非秒传回调流程 |
| [回收站列表](https://www.yuque.com/115yun/open/bg7l4328t98fwgex) | `GET /open/rb/list` |
| [回收站还原](https://www.yuque.com/115yun/open/gq293z80a3kmxbaq) | `POST /open/rb/revert` |
| [删除/清空回收站](https://www.yuque.com/115yun/open/gwtof85nmboulrce) | `POST /open/rb/del` |

## 视频播放

| 文档 | 方法与路径 |
|---|---|
| [获取视频在线播放地址](https://www.yuque.com/115yun/open/hqglxv3cedi3p9dz) | `GET /open/video/play` |
| [视频字幕列表](https://www.yuque.com/115yun/open/nx076h3glapoyh7u) | `GET /open/video/subtitle` |
| [获取视频播放进度](https://www.yuque.com/115yun/open/gssqdrsq6vfqigag) | `GET /open/video/history` |
| [记忆视频播放进度](https://www.yuque.com/115yun/open/bshagbxv1gzqglg4) | `POST /open/video/history` |
| [提交视频转码](https://www.yuque.com/115yun/open/nxt8r1qcktmg3oan) | `POST /open/video/video_push` |

## 云下载

| 文档 | 方法与路径 |
|---|---|
| [添加云下载链接任务](https://www.yuque.com/115yun/open/zkyfq2499gdn3mty) | `POST /open/offline/add_task_urls` |
| [添加云下载 BT 任务](https://www.yuque.com/115yun/open/svfe4unlhayvluly) | `POST /open/offline/add_task_bt` |
| [获取用户云下载任务列表](https://www.yuque.com/115yun/open/av2mluz7uwigz74k) | `GET /open/offline/get_task_list` |
| [删除用户云下载任务](https://www.yuque.com/115yun/open/pmgwc86lpcy238nw) | `POST /open/offline/del_task` |
| [清空云下载任务](https://www.yuque.com/115yun/open/uu5i4urb5ylqwfy4) | `POST /open/offline/clear_task` |
| [获取云下载配额信息](https://www.yuque.com/115yun/open/gif2n3smh54kyg0p) | `GET /open/offline/get_quota_info` |
| [解析 BT 种子](https://www.yuque.com/115yun/open/evez3u50cemoict1) | `POST /open/offline/torrent` |

## 维护规则

1. 每次开始 115 集成任务时，先检查总目录和「更新记录」。
2. 官方新增、删除或重命名文档时，更新本索引与顶部的核对日期。
3. 不在本仓库镜像完整官方文档；只记录稳定导航、实现约束和仓库已验证的兼容经验。
