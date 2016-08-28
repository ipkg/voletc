package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

var (
	testConsulUri = "consul://consul:8500"
	//testConsulUri = "consul://127.0.0.1:8500"

	testDrvCfg *DriverConfig
	testDriver *MyVolumeDriver

	testAppCfg = &AppConfig{Name: "test", Version: "0.1.1", Env: "dev"}
	testName   = testAppCfg.QualifiedName()
)

func init() {
	testDrvCfg = NewDriverConfig(testConsulUri, "./testrun", "test-driver")

	var err error
	if testDriver, err = NewVolumeDriver(testDrvCfg); err != nil {
		log.Fatal(err)
	}
}

func Test_VolumeDriver_Error_Create(t *testing.T) {
	req := volume.Request{
		Name: testName,
		Options: map[string]string{
			"n1/k1":                 "v1",
			"templates/config.json": `{"key": "${n1/k1}"}`,
		},
	}
	resp := testDriver.Create(req)
	if resp.Err == "" {
		t.Fatal("should fail")
	}
}

func Test_VolumeDriver_Create(t *testing.T) {

	req := volume.Request{
		Name: testName,
		Options: map[string]string{
			"n1/k1":                "v1",
			"template:inline.json": `{"key": "${n1/k1}"}`,
			"template:config.json": "./testdata/config.json",
		},
	}

	resp := testDriver.Create(req)
	if resp.Err != "" {
		t.Fatal(resp.Err)
	}

	c, err := NewAppConfigFromName(testName, testDriver.be)
	if err != nil {
		t.Fatal(err)
	}

	if len(c.Templates) < 1 {
		t.Log("no templates")
		t.Fail()
	}

	if v, ok := c.Keys["n1/k1"]; !ok || string(v) != "v1" {
		t.Fatal("missing keys")
	}

	if !t.Failed() {
		if len(c.Templates[0].Body) < 1 {
			t.Log("no template data")
			t.Fail()
		}
	}
}

func Test_VolumeDriver_Get(t *testing.T) {

	req1 := volume.Request{Name: testName}
	r2 := testDriver.Get(req1)
	if r2.Err != "" {
		t.Fatal(r2.Err)
	}

	t.Logf("%+v\n", *r2.Volume)

	req2 := volume.Request{Name: "does-not-exist"}
	r3 := testDriver.Get(req2)
	if r3.Err == "" {
		t.Fatal("get should fail")
	}
}

func Test_VolumeDriver_List(t *testing.T) {
	rsp := testDriver.List(volume.Request{})
	if rsp.Err != "" {
		t.Log(rsp.Err)
		t.Fail()
	}
	if len(rsp.Volumes) < 1 {
		t.Log("empty listing")
		t.Fail()
	}

	found := false
	for _, vol := range rsp.Volumes {
		if vol.Name == testName {
			found = true
			break

		}
	}

	if !found {
		t.Log("volume not found")
		t.Fail()
	}

	//t.Logf("%+v\n", rsp.Volumes)
}

func Test_VolumeDriver_Path(t *testing.T) {

	req1 := volume.Request{Name: testName}
	r2 := testDriver.Path(req1)
	if r2.Err != "" {
		t.Fatal(r2.Err)
	}

	if r2.Mountpoint != testDriver.cfg.MountBaseDir+testAppCfg.getOpaque(testAppCfg.Env) {
		t.Log("wrong mountpoint")
		t.Fail()
	}

	t.Logf("%+v\n", r2)

	req1 = volume.Request{Name: "does-not-exist"}
	r2 = testDriver.Path(req1)
	if r2.Err == "" {
		t.Fatal("path should fail")
	}
}

func Test_VolumeDriver_Mount(t *testing.T) {

	req2 := volume.MountRequest{Name: testName}
	r3 := testDriver.Mount(req2)
	if r3.Err != "" {
		t.Fatal(r3.Err)
	}

	b, err := ioutil.ReadFile(testDriver.cfg.MountBaseDir + testAppCfg.getOpaque(testAppCfg.Env+"/inline.json"))
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != `{"key": "v1"}` {
		t.Logf("wrong payload: '%s'\n", b)
		t.Fail()
	}

}

func Test_VolumeDriver_Unmount(t *testing.T) {
	req := volume.UnmountRequest{Name: testName}
	rsp := testDriver.Unmount(req)
	if rsp.Err != "" {
		t.Log(rsp.Err)
		t.Fail()
	}

	_, err := os.Stat(testDriver.cfg.MountBaseDir + testAppCfg.getOpaque(testAppCfg.Env))
	if err == nil {
		t.Log("file should not exist")
		t.Fail()
	}

}

func Test_VolumeDriver_Remove(t *testing.T) {
	req1 := volume.Request{Name: testName}
	r3 := testDriver.Remove(req1)
	if r3.Err != "" {
		t.Fatal(r3.Err)
	}

	req1 = volume.Request{Name: "does-not-exist"}
	r3 = testDriver.Remove(req1)
	if r3.Err == "" {
		t.Fatal("remove should fail")
	}

	// Cleanup
	cbe := testDriver.be.(*ConsulBackend)
	cbe.DeleteMap("")
}
