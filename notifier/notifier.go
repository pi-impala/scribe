package notifier

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type Notifier interface {
	io.Closer
	Notify(exts []string, paths ...string) (<-chan Event, <-chan error, error)
}

type Event interface {
	Path() string
}

type event struct {
	path string
}

func newEvent(path string) *event {
	return &event{
		path,
	}
}

func (e *event) Path() string {
	return e.path
}

type notifier struct {
	watcher *fsnotify.Watcher
	exts    []string
	events  chan Event
	errs    chan error
}

func New() (*notifier, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}

	return &notifier{
		watcher: watcher,
		exts:    []string{},
		events:  make(chan Event),
		errs:    make(chan error),
	}, nil
}

func (n *notifier) Close() error {
	return n.watcher.Close()
}

func (n *notifier) Notify(exts []string, paths ...string) (<-chan Event, <-chan error, error) {
	n.exts = exts
	for _, path := range paths {
		if err := n.watcher.Add(path); err != nil {
			return n.events, n.errs, errors.Wrapf(err, "failed to  add path %s", path)
		}
	}
	// ...TODO: flush out
	go func() {
		for {
			select {
			case event, ok := <-n.watcher.Events:
				if !ok {
					n.errs <- errors.New("channel error")
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					_, filename := filepath.Split(event.Name)
					for _, ext := range n.exts {
						if strings.HasSuffix(filename, ext) || len(n.exts) == 0 {
							evt := newEvent(event.Name)
							n.events <- evt
							break
						}
					}
				}
			case err, ok := <-n.watcher.Errors:
				if !ok {
					n.errs <- errors.New("channel error")
					return
				}
				n.errs <- err
			}
		}
	}()
	return n.events, n.errs, nil
}
