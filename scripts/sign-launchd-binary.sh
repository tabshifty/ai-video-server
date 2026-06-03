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

codesign --force --sign "$CODESIGN_IDENTITY" --timestamp=none "$BIN_PATH"
codesign --verify --strict --verbose=2 "$BIN_PATH"
