package action

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	var Fs = afero.NewOsFs()
	_, err := Fs.Create("/tmp/foo")
	if err != nil {
		t.Fail()
	}

	_, err = Fs.Create("/tmp/foo2")
	if err != nil {
		t.Fail()
	}

	var ctx Context
	err = ctx.Set(
		Path("/tmp/foo"),
		Target("/tmp/foo2"),
	)
	if err != nil {
		t.Fail()
	}

	expect := Context{path: "/tmp/foo", target: "/tmp/foo2"}
	assert.Equal(t, expect, ctx, "context does not match")
	_ = Fs.Remove("/tmp/foo")
	_ = Fs.Remove("/tmp/foo2")
}
