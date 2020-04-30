/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args:    cobra.MinimumNArgs(1),
	PreRunE: watchPreRunE,
	RunE:    watchRunE,
}

// Target directory flag
var watchTarget string

// Extension flag
var watchExt string

func init() {
	rootCmd.AddCommand(watchCmd)

	// Flags
	watchCmd.Flags().StringVarP(&watchTarget, "target", "t", "", "the destination directory")
	watchCmd.Flags().StringVarP(&watchExt, "extensions", "e", "", "extensions to watch (comma delimited)")
}

// Check that the target path is valid, and if not then try to set it to the working directory
func watchPreRunE(cmd *cobra.Command, args []string) error {
	if watchTarget == "" {
		t, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "no target path specified and unable to get cwd")
		}
		watchTarget = t
		return nil
	}

	if _, err := os.Stat(watchTarget); os.IsNotExist(err) {
		return errors.Wrapf(err, "invalid target directory %s", watchTarget)
	}

	for _, arg := range args {
		if _, err := os.Stat(arg); os.IsNotExist(err) {
			return errors.Wrapf(err, "invalid path: %s", arg)
		}
	}

	return nil
}

func watchRunE(cmd *cobra.Command, args []string) error {

	endChan := make(chan struct{})
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			s := sc.Text()
			if strings.Compare("q", s) == 0 {
				logrus.Println("received quit")
				endChan <- struct{}{}
				return
			}
		}
	}()

	logrus.Println("initializing watcher...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "failed to initialize watcher")
	}
	defer watcher.Close()

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					logrus.Println("detected new file: ", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Println(err)
			case _, ok := <-endChan:
				if !ok {
					return
				}
				logrus.Println("stopping...")
				return
			}
		}
	}()

	for _, arg := range args {
		if err = watcher.Add(arg); err != nil {
			return errors.Wrapf(err, "unable to watch path %s", arg)
		}
	}

	wg.Wait()

	return nil
}
