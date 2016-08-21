package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	driverScope = "global"
	driverName  = "voletc"
)

type DriverConfig struct {
	MountBaseDir string
	BackendType  string
	BackendAddr  string
	Prefix       string
}

func NewDriverConfig(backendUri, basedir, prefix string) *DriverConfig {
	idx := strings.Index(backendUri, "://")
	d := &DriverConfig{
		MountBaseDir: filepath.Join(basedir, prefix),
		BackendType:  backendUri[:idx],
		BackendAddr:  backendUri[idx+3:],
		Prefix:       prefix,
	}

	if !strings.HasSuffix(d.MountBaseDir, "/") {
		d.MountBaseDir = d.MountBaseDir + "/"
	}

	return d
}

type MyVolumeDriver struct {
	cfg *DriverConfig

	ve *VolEtc

	be Backend
}

func NewVolumeDriver(cfg *DriverConfig) (*MyVolumeDriver, error) {
	d := &MyVolumeDriver{cfg: cfg}
	os.MkdirAll(d.cfg.MountBaseDir, 0777)

	be, err := NewBackend(cfg)
	if err == nil {
		d.be = be
		d.ve = &VolEtc{be}
	}

	return d, err
}

// Instruct the plugin that the user wants to create a volume, given a user specified
// volume name. The plugin does not need to actually manifest the volume on the
// filesystem yet (until Mount is called). Opts is a map of driver specific options
// passed through from the user request.
func (m *MyVolumeDriver) Create(req volume.Request) volume.Response {
	log.Printf("[Create] Request: %+v\n", req)
	// Create kv structure on backend.
	c, err := NewAppConfigFromName(req.Name, m.be)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	mp, err := parseCreateReqOptions(req.Options)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	c.Set(mp)

	if err = c.Commit(); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{}
}

// Get the list of volumes registered with the plugin.
func (m *MyVolumeDriver) List(req volume.Request) volume.Response {
	log.Printf("[List] Request: %+v\n", req)

	ls, err := m.ve.List()
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	resp := volume.Response{Capabilities: volume.Capability{Scope: driverScope}}

	resp.Volumes = make([]*volume.Volume, len(ls))
	i := 0
	for _, v := range ls {
		resp.Volumes[i] = &volume.Volume{
			Name:       v.QualifiedName(),
			Mountpoint: m.cfg.MountBaseDir + v.getOpaque(v.Env),
		}
		i++
	}

	log.Printf("[List] Response: %+v", resp)
	return resp
}

// Get the volume info.
func (m *MyVolumeDriver) Get(req volume.Request) volume.Response {
	log.Printf("[Get] Request: %+v\n", req)

	c, err := m.ve.Get(req.Name)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	resp := volume.Response{
		Capabilities: volume.Capability{Scope: driverScope},
	}
	resp.Volume = &volume.Volume{
		Name:       req.Name,
		Mountpoint: m.cfg.MountBaseDir + c.getOpaque(c.Env),
		Status:     c.Metadata(),
	}

	log.Printf("[Get] Response: %+v\n", resp)
	return resp
}

// Delete the specified volume from disk. This request is issued when a user invokes
// docker rm -v to remove volumes associated with a container.
func (m *MyVolumeDriver) Remove(req volume.Request) volume.Response {
	log.Printf("[Remove] Request: %+v\n", req)

	c, err := m.ve.Get(req.Name)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	resp := volume.Response{}
	if err = c.Destroy(); err != nil {
		resp.Err = err.Error()
	}

	log.Printf("[Remove] Response: %+v\n", resp)
	return resp
}

// Respond with the path on the host filesystem where the volume has been made available,
// and/or a string error if an error occurred. Mountpoint is optional, however the plugin
// may be queried again later if one is not provided.
func (m *MyVolumeDriver) Path(req volume.Request) volume.Response {
	log.Printf("[Path] Request: %+v\n", req)

	c, err := m.ve.Get(req.Name)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	resp := volume.Response{Mountpoint: m.cfg.MountBaseDir + c.getOpaque(c.Env)}
	log.Printf("[Path] Response: %+v\n", resp)
	return resp
}

func (m *MyVolumeDriver) Mount(req volume.MountRequest) volume.Response {
	log.Printf("Mount: %+v\n", req)

	c, err := m.ve.Get(req.Name)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	dpath := m.cfg.MountBaseDir + c.getOpaque(c.Env)
	os.MkdirAll(dpath, 0777)

	if err = c.Generate(dpath); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Mountpoint: dpath}
}

// Indication that Docker no longer is using the named volume. This is called
// once per container stop. Plugin may deduce that it is safe to deprovision it at this point.
func (m *MyVolumeDriver) Unmount(req volume.UnmountRequest) volume.Response {
	log.Printf("[Unmount] Request: %+v\n", req)

	c, err := m.ve.Get(req.Name)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	dpath := m.cfg.MountBaseDir + c.getOpaque(c.Env)
	os.RemoveAll(dpath)

	return volume.Response{}
}

// Get the list of capabilities the driver supports. The driver is not required
// to implement this endpoint, however in such cases the default values will be taken.
func (m *MyVolumeDriver) Capabilities(req volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: driverScope}}
}

// convert template:<name> to templates/<name> for storage
func parseCreateReqOptions(m map[string]string) (map[string][]byte, error) {
	out := map[string][]byte{}
	for k, v := range m {
		if strings.HasPrefix(k, "template:") {
			l := strings.Index(k, ":") + 1
			var val []byte

			if strings.HasPrefix(v, "/") || strings.HasPrefix(v, "./") {
				b, err := ioutil.ReadFile(v)
				if err != nil {
					return nil, err
				}
				val = b
			} else {
				val = []byte(v)
			}
			out["templates/"+k[l:]] = val

		} else if strings.HasPrefix(k, "templates/") {
			return nil, fmt.Errorf("reserved prefix: 'templates/' in '%s'", k)
		} else {
			out[k] = []byte(v)
		}
	}
	return out, nil
}
