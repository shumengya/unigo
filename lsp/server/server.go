// Package server 是 UniGo 语言服务器（LSP）的实现。
//
// 目前只是骨架。计划复用 unigo/compiler 下的 parser 和 check，
// 按下面的顺序补齐能力：
//
//  1. 诊断：打开和修改 .ug 文件时跑 parser + check，把错误推给编辑器
//  2. 悬停：显示标识符的类型
//  3. 跳转定义、查找引用
//  4. 补全：关键字、局部变量、结构体字段、导入包的成员
//  5. 格式化：接 unigo fmt
package server

import (
	"errors"
	"io"
)

// ErrNotImplemented 表示语言服务器还没有实现。
var ErrNotImplemented = errors.New("unigo-lsp 还没有实现")

// Run 在 in/out 上以 JSON-RPC 方式提供 LSP 服务，直到输入结束。
func Run(in io.Reader, out io.Writer) error {
	return ErrNotImplemented
}
