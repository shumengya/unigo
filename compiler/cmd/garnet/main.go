package main

import (
	"fmt"
	"os"

	"garnet/compiler/compile"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "build" {
		fmt.Fprintln(os.Stderr, "用法: garnet build <file.gn>")
		os.Exit(2)
	}
	if err := compile.Build(os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
