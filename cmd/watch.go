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

	"github.com/pi-impala/scribe/action"
	"github.com/pi-impala/scribe/notifier"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directories for speciied file types and move newly created ones to a target directory",
	Long: `The "watch" command takes a list of paths to watch, as arguments. 
	For example, to watch ~/Desktop and ~/Downloads for png, pcap, or expk files and move them to ~/Documents/my_target/,
	run:

	scribe watch ~/Desktop ~/Downloads -e png,pcap,expk -t ~/Documents/my_target
	`,
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
		validArgs := map[string]struct{}{
			"q":    struct{}{},
			"quit": struct{}{},
		}
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			s := sc.Text()
			if _, ok := validArgs[s]; ok {
				logrus.Println("received quit")
				endChan <- struct{}{}
				return
			}
		}
	}()

	logrus.Println("initializing notifer...")

	n, err := notifier.New()
	if err != nil {
		return errors.Wrap(err, "failed to initialize notifer")
	}
	defer n.Close()

	// split extensions into slice, handle possible trailing comma
	exts := strings.Split(strings.TrimRight(watchExt, ","), ",")
	evts, errs, err := n.Notify(exts, args...)

	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()
		pending := []*action.Context{}
		for {
			select {
			case event, ok := <-evts:
				if !ok {
					flushPendingActions(pending)
					return
				}
				p := event.Path()
				logrus.Println("detected new file: ", p)

				ctx, err := newActionContext(p, watchTarget)
				if err != nil {
					logrus.Errorf("failed to create action for %s: %v", p, err)
				}

				pending = append(pending, ctx)

			case err, ok := <-errs:
				if !ok {
					flushPendingActions(pending)
					return
				}
				logrus.Println(err)
			case <-endChan:
				flushPendingActions(pending)
				logrus.Println("stopping...")
				return
			}
		}
	}()

	wg.Wait()

	return nil
}

func newActionContext(path string, target string) (*action.Context, error) {
	ctx := action.Context{}
	err := ctx.Set(
		action.Path(path),
		action.Target(target),
	)
	if err != nil {
		return nil, errors.Wrap(err, "action dispatch failed")
	}
	return &ctx, nil
}

func flushPendingActions(actions []*action.Context) {
	for _, ctx := range actions {
		if err := dispatch(ctx); err != nil {
			logrus.Errorf("failed to dispatch action: %v", err)
		} else {

			logrus.Println("moved ", ctx)
		}
	}
}

func dispatch(ctx *action.Context) error {
	a := action.FromContext(ctx, action.Move)
	return action.Dispatch(a)
}
