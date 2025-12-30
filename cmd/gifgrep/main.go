package main

import (
	"os"

	"github.com/steipete/gifgrep/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
