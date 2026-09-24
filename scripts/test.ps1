[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# 跑转译器测试，再检查编辑器的语法定义
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

Write-Host "== 转译器测试 =="
go test ./...

Write-Host "`n== tree-sitter 语法测试 =="
if (Test-Path editors/tree-sitter-garnet/node_modules) {
    Push-Location editors/tree-sitter-garnet
    npx tree-sitter test
    Pop-Location
} else {
    Write-Host "跳过：先在 editors/tree-sitter-garnet 下执行 npm install"
}

Write-Host "`n== VS Code 高亮测试 =="
if (Test-Path editors/vscode/node_modules) {
    Push-Location editors/vscode
    npm test
    Pop-Location
} else {
    Write-Host "跳过：先在 editors/vscode 下执行 npm install"
}
