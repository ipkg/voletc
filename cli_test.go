package main

import (
	"testing"
)

func Test_cli(t *testing.T) {

	cl, err := newCli(testDrvCfg)
	if err != nil {
		t.Fatal(err)
	}

	//printUsage()
	if err := cl.Run([]string{"create", "test2-0.1.0-dev",
		"db/name=dbname", "template:config.json=./testdata/config.json"}); err != nil {
		t.Fatal(err)
	}

	if err := cl.Run([]string{"create", "test3-0.1.0-dev",
		"db/name=dbname", "template:config.json=./testdata/config.json", "-dryrun"}); err != nil {
		t.Fatal(err)
	}
	if err := cl.Run([]string{"info", "test3-0.1.0-dev"}); err == nil {
		t.Log("should fail")
		t.Fail()
	}

	if err := cl.Run([]string{"info", "test2-0.1.0-dev"}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cl.Run([]string{"render", "test2-0.1.0-dev"}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cl.Run([]string{"ls"}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cl.Run([]string{"edit", "test2-0.1.0-dev",
		"db/username=dbuser,db/password=dbpasswd"}); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := cl.Run([]string{"edit"}); err == nil {
		t.Log("should fail")
		t.Fail()
	}

	if err := cl.Run([]string{"rm", "test2-0.1.0-dev", "-y"}); err != nil {
		t.Fatal(err)
	}

}
