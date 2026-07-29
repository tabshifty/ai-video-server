# 115 开放平台授权与令牌

> 官方文档核对日期：2026-07-29。实现前必须再次打开链接核对。

## 授权模式选择

| 场景 | 模式 | 客户端是否持有 AppSecret | 官方文档 |
|---|---|---:|---|
| 有可保密的服务端 | OAuth 2.0 授权码 | 否，只由服务端持有 | [授权码模式](https://www.yuque.com/115yun/open/okr2cq0wywelscpe) |
| 无后端的第三方客户端 | 手机扫码 OAuth 2.0 + PKCE | 不需要 | [手机扫码授权 PKCE 模式](https://www.yuque.com/115yun/open/shtpzfhewv5nag11) |

## 授权码模式

### 1. 请求授权

```http
GET https://passportapi.115.com/open/authorize
```

Query 参数：

- `client_id`：APP ID。
- `redirect_uri`：URL 编码后的回调地址，域名必须预先在应用管理中设置。
- `response_type=code`。
- `state`：强烈建议使用高熵一次性值，回调时必须与服务端会话中的原值常量时间比较。

用户未登录时由 115 官方页面承担登录；已登录时完成授权后回调 `redirect_uri` 并附带 `code` 和可选 `state`。

### 2. 授权码换令牌

```http
POST https://passportapi.115.com/open/authCodeToToken
Content-Type: application/x-www-form-urlencoded
```

Form 参数：`client_id`、`client_secret`、`code`、与第一步一致的 `redirect_uri`、`grant_type=authorization_code`。此请求必须在服务端发起。

## 手机扫码 PKCE 模式

### 1. 生成设备码

1. 生成 43–128 字符的高熵 `code_verifier`。
2. 对其执行 SHA-256，再进行无填充 Base64URL 编码得到 `code_challenge`。
3. 保留 `code_verifier` 仅用于本次授权，不传给日志或遥测。

```http
POST https://passportapi.115.com/open/authDeviceCode
Content-Type: application/x-www-form-urlencoded
```

Form 参数：`client_id`、`code_challenge`、`code_challenge_method=sha256`。使用返回的 `data.qrcode` 生成二维码，并保留 `uid`、`time`、`sign`。

### 2. 长轮询扫码状态

```http
GET https://qrcodeapi.115.com/get/status/?uid=...&time=...&sign=...
```

- `state=0`：二维码无效，结束轮询。
- `state=1`：继续轮询。
- `data.status=1`：已扫码，等待用户确认。
- `data.status=2`：用户已确认授权，结束轮询并换取令牌。

该接口是长轮询，无状态变化时可能会持续到超时。客户端必须支持取消、总超时和有上限的重连。

### 3. 设备码换令牌

```http
POST https://passportapi.115.com/open/deviceCodeToToken
Content-Type: application/x-www-form-urlencoded
```

Form 参数：`uid`、原始 `code_verifier`。

## 令牌结构与刷新

两种授权方式都返回：

- `access_token`：资源 API 的 Bearer 凭证。
- `refresh_token`：用于刷新 access token；官方授权文档标注有效期为 1 年。
- `expires_in`：access token 有效期，单位秒。

```http
POST https://passportapi.115.com/open/refreshToken
Content-Type: application/x-www-form-urlencoded

refresh_token=<current_refresh_token>
```

官方明确提示不要频繁刷新，否则可能进入频控。刷新响应会返回新 `access_token`、新 `refresh_token` 和 `expires_in`；新 refresh token 不延长原有有效期。必须原子持久化整组新令牌，不要只更新 access token。

官方文档中不同流程的 `expires_in` 展示值不同：授权响应示例为 7200，刷新响应示例为 2592000。这些只能视为示例，代码必须使用实际响应值。

## 资源 API 请求

```http
Authorization: Bearer <access_token>
```

资源 API 官方文档当前使用 `https://proapi.115.com` 作为基础域名；实现时仍需从目标接口页面再次核对。

## 实现必备状态

- 授权会话：`state`、发起时间、回调地址、已消费标记。
- PKCE 会话：`uid`、`time`、`sign`、`code_verifier`、截止时间、取消状态。
- 令牌记录：密文或密钥引用、绝对到期时间、授权用户/应用归属、令牌版本。
- 刷新协调：单航锁、双重到期检查、旋转令牌原子替换、失败后重新授权标记。
