[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# 构建 unigo 和 unigo-lsp 到 bin/
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")
New-Item -ItemType Directory -Force -Path bin | Out-Null
go build -o bin/unigo.exe ./compiler/cmd/unigo
go build -o bin/unigo-lsp.exe ./lsp/cmd/unigo-lsp
Write-Host "已生成 bin/unigo.exe 和 bin/unigo-lsp.exe"
