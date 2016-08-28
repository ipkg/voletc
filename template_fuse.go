package main

import (
	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

//
// FUSE
//
var _ fs.Node = (*Template)(nil)

func (t *Template) Attr(ctx context.Context, a *fuse.Attr) error {
	// TODO: set inode based on sha1
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(t.Body))
	return nil
}

func (t *Template) ReadAll(ctx context.Context) ([]byte, error) {
	// TODO: return rendered ???
	return t.rendered, nil
}
