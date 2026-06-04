#!/usr/bin/env bash
# 用稳定签名身份给 launchd 二进制重新签名，避免 ad-hoc cdhash 漂移后 TCC 失效。

set -Eeuo pipefail

if [[ $# -ne 1 ]]; then
  cat >&2 <<'EOF'
用法: sign-launchd-binary.sh <binary-path>
EOF
  exit 1
fi

BIN_PATH="$1"
CODESIGN_IDENTITY="${CODESIGN_IDENTITY:-}"
CODESIGN_KEYCHAIN="${CODESIGN_KEYCHAIN:-}"
CODESIGN_KEYCHAIN_PASSWORD="${CODESIGN_KEYCHAIN_PASSWORD:-}"

if [[ -z "$CODESIGN_IDENTITY" ]]; then
  printf '缺少 CODESIGN_IDENTITY；请在部署机 .env 里配置稳定签名身份\n' >&2
  exit 1
fi
if [[ "$CODESIGN_IDENTITY" == "-" ]]; then
  printf 'CODESIGN_IDENTITY 不能是 ad-hoc (-)；这条链路要求稳定签名身份\n' >&2
  exit 1
fi
if ! command -v codesign >/dev/null 2>&1; then
  printf '找不到 codesign\n' >&2
  exit 1
fi
if [[ ! -f "$BIN_PATH" ]]; then
  printf '找不到二进制: %s\n' "$BIN_PATH" >&2
  exit 1
fi

if [[ -n "$CODESIGN_KEYCHAIN" ]]; then
  if [[ ! -f "$CODESIGN_KEYCHAIN" ]]; then
    printf '找不到 CODESIGN_KEYCHAIN 指向的 keychain: %s\n' "$CODESIGN_KEYCHAIN" >&2
    exit 1
  fi
  if [[ -n "$CODESIGN_KEYCHAIN_PASSWORD" ]]; then
    security unlock-keychain -p "$CODESIGN_KEYCHAIN_PASSWORD" "$CODESIGN_KEYCHAIN"
    security set-key-partition-list -S apple-tool:,apple: -s -k "$CODESIGN_KEYCHAIN_PASSWORD" "$CODESIGN_KEYCHAIN" >/dev/null
  else
    printf '提示: CODESIGN_KEYCHAIN 已设置，但没有 CODESIGN_KEYCHAIN_PASSWORD；若 keychain 处于锁定状态，codesign 可能失败\n' >&2
  fi
fi

if [[ -n "$CODESIGN_KEYCHAIN" ]]; then
  if ! security find-identity -v -p codesigning "$CODESIGN_KEYCHAIN" | grep -F "$CODESIGN_IDENTITY" >/dev/null 2>&1; then
    printf '在 %s 里找不到可用的 codesigning identity: %s\n' "$CODESIGN_KEYCHAIN" "$CODESIGN_IDENTITY" >&2
    printf '请确认这是 Apple-issued 的 Developer ID Application / Apple Development 证书，并已在本机完成一次 trust 授权。\n' >&2
    exit 1
  fi
else
  if ! security find-identity -v -p codesigning | grep -F "$CODESIGN_IDENTITY" >/dev/null 2>&1; then
    printf '当前 keychain 搜索列表里找不到可用的 codesigning identity: %s\n' "$CODESIGN_IDENTITY" >&2
    printf '请确认这是 Apple-issued 的 Developer ID Application / Apple Development 证书，并已在本机完成一次 trust 授权。\n' >&2
    exit 1
  fi
fi

if [[ -n "$CODESIGN_KEYCHAIN" ]]; then
  codesign --force --sign "$CODESIGN_IDENTITY" --keychain "$CODESIGN_KEYCHAIN" --timestamp=none "$BIN_PATH"
else
  codesign --force --sign "$CODESIGN_IDENTITY" --timestamp=none "$BIN_PATH"
fi
codesign --verify --strict --verbose=2 "$BIN_PATH"
