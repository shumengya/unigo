package main

import (
	"fmt"
	"os"

	"unigo/compiler/compile"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "build" {
		fmt.Fprintln(os.Stderr, "用法: unigo build <file.ug>")
		os.Exit(2)
	}
	if err := compile.Build(os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
