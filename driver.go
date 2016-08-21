package main

import (
	"fmt"
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

	be Backend
}

func NewVolumeDriver(cfg *DriverConfig) (*MyVolumeDriver, error) {
	d := &MyVolumeDriver{cfg: cfg}
	os.MkdirAll(d.cfg.MountBaseDir, 0777)

	be, err := d.newBackend()
	if err == nil {
		d.be = be
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

	mp := parseCreateReqOptions(req.Options)
	c.Set(mp)

	if err = c.Init(); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{}
}

// convert template:<name> to templates/<name> for storage
func parseCreateReqOptions(m map[string]string) map[string][]byte {
	out := map[string][]byte{}
	for k, v := range m {
		//log.Println(k)
		if strings.HasPrefix(k, "template:") {
			l := strings.Index(k, ":") + 1
			out["templates/"+k[l:]] = []byte(v)
		} else {
			out[k] = []byte(v)
		}
	}
	return out
}

// Get the list of volumes registered with the plugin.
func (m *MyVolumeDriver) List(req volume.Request) volume.Response {
	log.Printf("[List] Request: %+v\n", req)

	mp, err := m.be.GetMap("")
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	resp := volume.Response{Capabilities: volume.Capability{Scope: driverScope}}
	resp.Volumes = []*volume.Volume{}

	for k, _ := range mp {
		pp := strings.Split(k, "/")
		// TODO: Support app and version without environment
		// this would be needed to check for templates but no keys
		if pp[2] == "templates" {
			continue
		}

		n := fmt.Sprintf("%s-%s-%s", pp[0], pp[1], pp[2])
		vol := &volume.Volume{
			Name:       n,
			Mountpoint: m.cfg.MountBaseDir + n,
		}
		resp.Volumes = append(resp.Volumes, vol)

	}

	log.Printf("[List] Response: %+v", resp)
	return resp
}

// Get the volume info.
func (m *MyVolumeDriver) Get(req volume.Request) volume.Response {
	log.Printf("[Get] Request: %+v\n", req)

	c, err := NewAppConfigFromName(req.Name, m.be)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	//if len(c.Keys) < 1 && len(c.Templates) < 1 {
	if len(c.Keys) < 1 {
		return volume.Response{Err: "not found: " + req.Name}
	}

	resp := volume.Response{
		Capabilities: volume.Capability{Scope: driverScope},
	}
	resp.Volume = &volume.Volume{
		Name:       req.Name,
		Mountpoint: m.cfg.MountBaseDir + c.getOpaque(c.Env),
		Status:     c.Meta(),
	}

	log.Printf("[Get] Response: %+v\n", resp)
	return resp
}

// Delete the specified volume from disk. This request is issued when a user invokes
// docker rm -v to remove volumes associated with a container.
func (m *MyVolumeDriver) Remove(req volume.Request) volume.Response {
	log.Printf("[Remove] Request: %+v\n", req)

	resp := volume.Response{}
	c, err := NewAppConfigFromName(req.Name, m.be)
	if err == nil {
		err = c.Destroy()
	}

	if err != nil {
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

	resp := volume.Response{}
	c, err := NewAppConfigFromName(req.Name, m.be)
	if err == nil {
		if len(c.Keys) < 1 && len(c.Templates) < 1 {
			resp.Err = fmt.Sprintf("not found: %s", req.Name)
		} else {
			resp.Mountpoint = m.cfg.MountBaseDir + c.getOpaque(c.Env)
		}
	} else {
		resp.Err = err.Error()
	}

	log.Printf("[Path] Response: %+v\n", resp)
	return resp
}

func (m *MyVolumeDriver) Mount(req volume.MountRequest) volume.Response {
	log.Printf("Mount: %+v\n", req)

	c, err := NewAppConfigFromName(req.Name, m.be)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	dpath := m.cfg.MountBaseDir + c.getOpaque(c.Env)
	os.MkdirAll(dpath, 0777)

	if err = c.Load(dpath); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Mountpoint: dpath}
}

// Indication that Docker no longer is using the named volume. This is called
// once per container stop. Plugin may deduce that it is safe to deprovision it at this point.
func (m *MyVolumeDriver) Unmount(req volume.UnmountRequest) volume.Response {
	log.Printf("[Unmount] Request: %+v\n", req)

	c, err := NewAppConfigFromName(req.Name, m.be)
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

func (m *MyVolumeDriver) newBackend() (Backend, error) {
	var (
		be  Backend
		err error
	)

	switch m.cfg.BackendType {
	case "consul":
		be, err = NewConsulBackend(m.cfg.BackendAddr, m.cfg.Prefix)

	default:
		err = fmt.Errorf("backend not supported: %s", m.cfg.BackendType)
	}

	return be, err
}
