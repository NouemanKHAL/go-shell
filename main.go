package main

import (
	"os"

	"github.com/NouemanKHAL/go-shell/internal/shell"
)

func main() {
	sh, err := shell.NewShell()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	for {
		sh.Prompt()
	}
}
