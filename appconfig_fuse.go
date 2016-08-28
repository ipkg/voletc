package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

//
// FUSE
//

// FS implements the hello world file system.
type AppConfigFS struct {
	//be *ConsulBackend
	acfg     *AppConfig
	mntPoint string
}

func (a *AppConfigFS) Root() (fs.Node, error) {
	// Create and load appconfig
	return a.acfg, nil
}

func (a *AppConfigFS) Unmount() error {
	return fuse.Unmount(a.mntPoint)
}

func (a *AppConfigFS) Mount() error {
	a.acfg.cacheRender()

	c, err := fuse.Mount(
		a.mntPoint,
		fuse.FSName("voletc"),
		fuse.Subtype("consul"),
		fuse.LocalVolume(),
		fuse.VolumeName(fmt.Sprintf("%s %s %s", a.acfg.Name, a.acfg.Version, a.acfg.Env)),
	)
	if err != nil {
		return err
	}

	defer func() {
		if e := c.Close(); e != nil {
			log.Println("ERR", e)
		}
	}()

	if err = fs.Serve(c, a); err == nil {
		<-c.Ready
		err = c.MountError
	}

	return err
}

var _ fs.Node = (*AppConfig)(nil)

func (ac *AppConfig) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("Dir.Attr")
	attr.Inode = 1
	attr.Mode = os.ModeDir | 0555
	return nil
}

var _ = fs.NodeRequestLookuper(&AppConfig{})

//func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
func (a *AppConfig) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	log.Println("Lookup", req.Name)

	for _, v := range a.Templates {

		// File
		if req.Name == v.Name {
			return v, nil
		}

		arr := strings.Split(v.Name, "/")

		if req.Name == arr[len(arr)-1] {
			return v, nil
		}

		// Directory
		if req.Name == arr[0] {
			return a, nil
		}

		//log.Println(a.QualifiedName(), dirDirs[i])
	}

	return nil, fuse.ENOENT

}

func (a *AppConfig) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAll")
	dirDirs := []fuse.Dirent{}

	for i, v := range a.Templates {
		fst := strings.Split(v.Name, "/")

		if fst[0] == v.Name {
			// File
			dirDirs = append(dirDirs, fuse.Dirent{Inode: uint64(i) + 1, Name: v.Name, Type: fuse.DT_File})
		} else {
			// Dirs
			//for j, a := range fst[:len(fst)-1] {
			//	dirDirs = append(dirDirs, fuse.Dirent{Inode: uint64(i + j), Name: fst[0], Type: fuse.DT_Dir})
			//}
			//// Add file
			//dirDirs = append(dirDirs, fuse.Dirent{Inode: uint64(i + len(fst)), Name: v.Name, Type: fuse.DT_File})
			dirDirs[i] = fuse.Dirent{Inode: uint64(i) + 1, Name: fst[0], Type: fuse.DT_Dir}
		}

		//log.Println(a.QualifiedName(), dirDirs[i])
	}

	return dirDirs, nil
}
