#!/bin/bash
# Adapted from the Deno installer: Copyright 2019 the Deno authors. All rights reserved. MIT license.
# Ref: https://github.com/denoland/deno_install
# TODO(everyone): Keep this script simple and easily auditable.

# TODO(mf): this should work on Linux and macOS. Not intended for Windows.

set -e

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

if [ "$arch" = "aarch64" ]; then
	arch="arm64"
fi

if [ $# -eq 0 ]; then
	agfapi_uri="https://github.com/s-hammon/agfapi/releases/latest/download/agfapi_${os}_${arch}"
else
	agfapi_uri="https://github.com/s-hammon/agfapi/releases/download/${1}/agfapi_${os}_${arch}"
fi

agfapi_install="${AGFAPI_INSTALL:-/usr/local}"
bin_dir="${agfapi_install}/bin"
exe="${bin_dir}/agfapi"

if [ ! -d "${bin_dir}" ]; then
	mkdir -p "${bin_dir}"
fi

curl --silent --show-error --location --fail --location --output "${exe}" "$agfapi_uri"
chmod +x "${exe}"

echo "agfapi was installed successfully to ${exe}"
if command -v agfapi >/dev/null; then
	echo "Run 'agfapi --help' to get started"
fi
