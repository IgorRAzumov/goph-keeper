package main

import (
	"os"

	"goph-keeper/internal/client/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
