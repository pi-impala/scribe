package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
)

var App struct {
	Watch watchCmd `cmd help:"Watch designated folder."`
}

type watchCmd struct {
	Path   string `arg name:"path" type:"path" help:"Path to watch."`
	Target string `type:"path" name:"target"`
}

func (w *watchCmd) Run() error {
	/*
		TODO: 	 Get information for $PATH and $TARGET directory
				 * Check if target dir is empty, default to zero-value?
				 * Create watch function, first validate input and
				 * then pass to watch function
	*/
	var target string
	if w.Target == "" {
		t, err := os.Getwd()
		if err != nil {
			return err
		}
		target = t
	}
	var curPath string
	if w.Path == "" {
		cp, err := os.Getwd()
		if err != nil {
			return err
		}
		curPath = cp
	}
	startWatch(curPath, target)
	return nil
}

func startWatch(pathStr string, targetStr string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(pathStr)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func main() {
	// Context
	ctx := kong.Parse(&App)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
