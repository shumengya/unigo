#!/usr/bin/env sh
# 构建 unigo 和 unigo-lsp 到 bin/
set -e
cd "$(dirname "$0")/.."
mkdir -p bin
go build -o bin/unigo ./compiler/cmd/unigo
go build -o bin/unigo-lsp ./lsp/cmd/unigo-lsp
echo "已生成 bin/unigo 和 bin/unigo-lsp"
