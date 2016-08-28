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
}

func runServer() error {
	driver, err := NewVolumeDriver(driverConfig)
	if err != nil {
		return err
	}
	// New docker volume driver handler
	handler := volume.NewHandler(driver)

	log.Println("Starting sevice on:", *listenAddr)
	if err = handler.ServeTCP(driverName, *listenAddr); err != nil {
		return err
	}

	select {}
}

func runClient() {
	cl, err := newCli(driverConfig)
	if err == nil {
		err = cl.Run(flag.Args())
	}

	if err != nil {
		printUsage()
		fmt.Println("Volume:", err)
		os.Exit(1)
	}
}

func main() {
	//var err error
	if *serverMode {
		if err := runServer(); err != nil {
			log.Fatal(err)
		}
		return
	}

	runClient()

}
