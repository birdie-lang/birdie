package main

import (
	"os"

	"github.com/birdie-lang/birdie/cmd/birdie/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
