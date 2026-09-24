#!/usr/bin/env sh
# 跑转译器测试，再检查编辑器的语法定义
set -e
cd "$(dirname "$0")/.."

echo "== 转译器测试 =="
go test ./...

echo
echo "== tree-sitter 语法测试 =="
if [ -d editors/tree-sitter-unigo/node_modules ]; then
    (cd editors/tree-sitter-unigo && npx tree-sitter test)
else
    echo "跳过：先在 editors/tree-sitter-unigo 下执行 npm install"
fi

echo
echo "== VS Code 高亮测试 =="
if [ -d editors/vscode/node_modules ]; then
    (cd editors/vscode && npm test)
else
    echo "跳过：先在 editors/vscode 下执行 npm install"
fi
