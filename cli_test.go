package main

import (
	"testing"
)

func Test_cli(t *testing.T) {

	cl, err := newCli(testDrvCfg)
	if err != nil {
		t.Fatal(err)
	}

	printUsage()

	if _, err := cl.Run([]string{"create", "test2-0.1.0-dev",
		"db/name=dbname", "template:config.json=./testdata/config.json"}); err != nil {
		t.Fatal(err)
	}

	if _, err := cl.Run([]string{"create", "test3-0.1.0-dev",
		"db/name=dbname", "template:config.json=./testdata/config.json", "-dryrun"}); err != nil {
		t.Fatal(err)
	}
	if _, err := cl.Run([]string{"info", "test3-0.1.0-dev"}); err == nil {
		t.Fatal("info should fail")
	}

	if _, err := cl.Run([]string{"info", "test2-0.1.0-dev"}); err != nil {
		t.Fatal(err)
	}

	if _, err := cl.Run([]string{"render", "test2-0.1.0-dev"}); err != nil {
		t.Fatal(err)
	}

	if _, err = cl.Run([]string{"ls"}); err != nil {
		t.Fatal(err)
	}

}
