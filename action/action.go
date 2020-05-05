package action

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Context struct {
	path   string
	target string
}

func (ctx *Context) Set(args ...contextFunc) error {
	for _, arg := range args {
		if err := arg(ctx); err != nil {
			return err
		}
	}
	return nil
}

type contextFunc func(*Context) error

// Path sets the action context path
func Path(path string) contextFunc {
	return func(ctx *Context) error {
		if _, err := os.Stat(path); err != nil {
			return errors.Wrapf(err, "invalid path: %v", path)
		}
		ctx.path = path

		return nil
	}
}

// Target sets the action context target
func Target(target string) contextFunc {
	return func(ctx *Context) error {
		if _, err := os.Stat(target); err != nil {
			return errors.Wrapf(err, "invalid target path: %v", target)
		}
		ctx.target = target
		return nil
	}
}

type action struct {
	ctx *Context
	fn  contextFunc
}

func FromContext(ctx *Context, fn func(*Context) error) *action {
	return &action{
		ctx: ctx,
		fn:  fn,
	}
}

func Move(ctx *Context) error {
	_, file := filepath.Split(ctx.path)
	tpath := filepath.Join(ctx.target, file)

	if err := os.Rename(ctx.path, tpath); err != nil {
		return errors.Wrapf(err, "failed to move file %s to %s", ctx.path, tpath)
	}
	return nil
}

// Dispatch executes the action function using the Context
func Dispatch(a *action) error {
	return a.fn(a.ctx)
}
