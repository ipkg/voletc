package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/docker/go-plugins-helpers/volume"
)

var (
	driverConfig *DriverConfig
)

func init() {
	flag.Usage = printUsage
	flag.Parse()
	setDefaultVersionInfo()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	driverConfig = NewDriverConfig(*backendUri, *baseDir, *dataPrefix)
	driverConfig.EncryptionKey = *encDec
	cl, err := newCli(driverConfig)
	if err != nil {
		log.Fatal(err)
	}

	extOnRet, err := cl.Run(flag.Args())
	if err != nil {
		fmt.Println("Volume", err)
		os.Exit(1)
	} else if extOnRet {
		os.Exit(0)
	}
}

func main() {
	// New Driver
	driver, err := NewVolumeDriver(driverConfig)
	if err != nil {
		log.Fatal(err)
	}
	// New docker volume driver handler
	handler := volume.NewHandler(driver)

	log.Println("Starting sevice on:", *listenAddr)
	if err = handler.ServeTCP(driverName, *listenAddr); err != nil {
		log.Fatal(err)
	}

	select {}
}
