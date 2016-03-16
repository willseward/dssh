package main

import (
	"os"
)

func main() {
	app := NewCliApp()
	app.Run(os.Args)
}
