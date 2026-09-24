[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# 构建 garnet 和 garnet-lsp 到 bin/
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")
New-Item -ItemType Directory -Force -Path bin | Out-Null
go build -o bin/garnet.exe ./compiler/cmd/garnet
go build -o bin/garnet-lsp.exe ./lsp/cmd/garnet-lsp
Write-Host "已生成 bin/garnet.exe 和 bin/garnet-lsp.exe"
