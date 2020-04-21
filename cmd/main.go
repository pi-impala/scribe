package main

import (
	"github.com/alecthomas/kong"
)

var App struct {

}

type watch struct {
	Path   string `arg name:"path" type:"path" help:"Path to watch."`
	target string `type:"path" name:"target"`
} `cmd help: "Watch designated folder."`

func (w *watch) Run() error {
	/*
	TODO: 	 Get information for $PATH and $TARGET directory
			 * Check if target dir is empty, default to zero-value?
			 * Create watch function, first validate input and
			 * then pass to watch function
	*/

	return nil
}

func main() {
	// Context
	ctx := kong.Parse(&App)
	switch ctx.Command() {
	case "watch":
	default:
		panic(ctx.Command())
	}
}
