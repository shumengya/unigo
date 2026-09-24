#!/usr/bin/env sh
# 构建 garnet 和 garnet-lsp 到 bin/
set -e
cd "$(dirname "$0")/.."
mkdir -p bin
go build -o bin/garnet ./compiler/cmd/garnet
go build -o bin/garnet-lsp ./lsp/cmd/garnet-lsp
echo "已生成 bin/garnet 和 bin/garnet-lsp"
