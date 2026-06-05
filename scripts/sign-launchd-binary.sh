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
CODESIGN_ENV_FILE="${CODESIGN_ENV_FILE:-}"
CODESIGN_IDENTITY="${CODESIGN_IDENTITY:-}"
CODESIGN_IDENTIFIER="${CODESIGN_IDENTIFIER:-com.chee.videos.server}"
CODESIGN_KEYCHAIN="${CODESIGN_KEYCHAIN:-}"
CODESIGN_KEYCHAIN_PASSWORD="${CODESIGN_KEYCHAIN_PASSWORD:-}"
CODESIGN_KEYCHAIN_PASSWORD_FILE="${CODESIGN_KEYCHAIN_PASSWORD_FILE:-}"

strip_env_quotes() {
  local value="$1"
  if [[ "${value:0:1}" == '"' && "${value: -1}" == '"' ]]; then
    value="${value:1:${#value}-2}"
  elif [[ "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s' "$value"
}

expand_home_path() {
  local value="$1"
  case "$value" in
    '$HOME'/*)
      value="$HOME/${value#'$HOME'/}"
      ;;
    '${HOME}'/*)
      value="$HOME/${value#'${HOME}'/}"
      ;;
    '~'/*)
      value="$HOME/${value#'~'/}"
      ;;
  esac
  printf '%s' "$value"
}

load_codesign_env_file() {
  local env_file="$1"
  local line key value
  if [[ ! -f "$env_file" ]]; then
    printf '找不到 CODESIGN_ENV_FILE 指向的环境文件: %s\n' "$env_file" >&2
    exit 1
  fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    [[ -z "$line" || "${line:0:1}" == "#" ]] && continue
    key="${line%%=*}"
    value="${line#*=}"
    value="$(strip_env_quotes "$value")"
    case "$key" in
      CODESIGN_IDENTITY)
        [[ -n "$value" && -z "$CODESIGN_IDENTITY" ]] && CODESIGN_IDENTITY="$value"
        ;;
      CODESIGN_IDENTIFIER)
        [[ -n "$value" && "$CODESIGN_IDENTIFIER" == "com.chee.videos.server" ]] && CODESIGN_IDENTIFIER="$value"
        ;;
      CODESIGN_KEYCHAIN)
        [[ -n "$value" && -z "$CODESIGN_KEYCHAIN" ]] && CODESIGN_KEYCHAIN="$(expand_home_path "$value")"
        ;;
      CODESIGN_KEYCHAIN_PASSWORD)
        [[ -n "$value" && -z "$CODESIGN_KEYCHAIN_PASSWORD" ]] && CODESIGN_KEYCHAIN_PASSWORD="$value"
        ;;
      CODESIGN_KEYCHAIN_PASSWORD_FILE)
        [[ -n "$value" && -z "$CODESIGN_KEYCHAIN_PASSWORD_FILE" ]] && CODESIGN_KEYCHAIN_PASSWORD_FILE="$(expand_home_path "$value")"
        ;;
    esac
  done < "$env_file"
}

if [[ -n "$CODESIGN_ENV_FILE" ]]; then
  load_codesign_env_file "$CODESIGN_ENV_FILE"
fi

if [[ -z "$CODESIGN_IDENTITY" ]]; then
  printf '缺少 CODESIGN_IDENTITY；请在部署机 .env 里配置稳定签名身份，或设置 CODESIGN_ENV_FILE\n' >&2
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
if [[ -z "$CODESIGN_IDENTIFIER" ]]; then
  printf 'CODESIGN_IDENTIFIER 不能为空；请使用稳定 bundle identifier\n' >&2
  exit 1
fi
if [[ -n "$CODESIGN_KEYCHAIN_PASSWORD_FILE" ]]; then
  if [[ ! -f "$CODESIGN_KEYCHAIN_PASSWORD_FILE" ]]; then
    printf '找不到 CODESIGN_KEYCHAIN_PASSWORD_FILE 指向的密码文件: %s\n' "$CODESIGN_KEYCHAIN_PASSWORD_FILE" >&2
    exit 1
  fi
  if [[ -z "$CODESIGN_KEYCHAIN_PASSWORD" ]]; then
    CODESIGN_KEYCHAIN_PASSWORD="$(<"$CODESIGN_KEYCHAIN_PASSWORD_FILE")"
  fi
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
  codesign --force --sign "$CODESIGN_IDENTITY" --identifier "$CODESIGN_IDENTIFIER" --keychain "$CODESIGN_KEYCHAIN" --timestamp=none "$BIN_PATH"
else
  codesign --force --sign "$CODESIGN_IDENTITY" --identifier "$CODESIGN_IDENTIFIER" --timestamp=none "$BIN_PATH"
fi
codesign --verify --strict --verbose=2 "$BIN_PATH"
