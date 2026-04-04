#!/usr/bin/env bash
#
# 本地构建 Android AAR（gomobile）。在仓库根目录执行：
#   chmod +x scripts/build_android_local.sh
#   ./scripts/build_android_local.sh
#
# 可通过环境变量覆盖默认路径，例如：
#   ANDROID_HOME=... ANDROID_NDK_HOME=... ./scripts/build_android_local.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_BIN="${ROOT_DIR}/.tools/bin"

mkdir -p "${TOOLS_BIN}"

# macOS 默认 SDK 路径；Linux 常见为 $HOME/Android/Sdk
if [[ "$(uname -s)" == "Darwin" ]]; then
	_DEFAULT_SDK="${HOME}/Library/Android/sdk"
else
	_DEFAULT_SDK="${HOME}/Android/Sdk"
fi

export ANDROID_HOME="${ANDROID_HOME:-${_DEFAULT_SDK}}"
export ANDROID_SDK_ROOT="${ANDROID_SDK_ROOT:-${ANDROID_HOME}}"

# 若未指定 NDK，尝试使用 ANDROID_HOME/ndk 下唯一版本目录
if [[ -z "${ANDROID_NDK_HOME:-}" ]]; then
	if [[ -d "${ANDROID_HOME}/ndk" ]]; then
		# shellcheck disable=SC2012
		_NDK_CANDIDATE="$(ls -1 "${ANDROID_HOME}/ndk" 2>/dev/null | head -n1)"
		if [[ -n "${_NDK_CANDIDATE}" ]]; then
			export ANDROID_NDK_HOME="${ANDROID_HOME}/ndk/${_NDK_CANDIDATE}"
		fi
	fi
fi

export PATH="${TOOLS_BIN}:${PATH}"
export GOBIN="${TOOLS_BIN}"
export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

echo "ROOT_DIR=${ROOT_DIR}"
echo "ANDROID_HOME=${ANDROID_HOME}"
echo "ANDROID_NDK_HOME=${ANDROID_NDK_HOME:-<unset>}"
echo "GOBIN=${GOBIN}"

if [[ -z "${ANDROID_NDK_HOME:-}" ]]; then
	echo "错误: 未找到 ANDROID_NDK_HOME。请安装 Android NDK（SDK Manager），或手动设置：" >&2
	echo "  export ANDROID_NDK_HOME=/path/to/ndk/<version>" >&2
	exit 1
fi

cd "${ROOT_DIR}"

go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init -v

make android
