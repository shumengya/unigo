package main

import (
	"fmt"
	"os"

	"unigo/lsp/server"
)

func main() {
	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
