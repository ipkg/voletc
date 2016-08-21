package main

import (
	"flag"
	"log"
	"os"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	defaultConsulUri = "consul://localhost:8500"
	defaultBaseDir   = "/opt"
)

var (
	baseDir    = flag.String("dir", defaultBaseDir, "Data directory")
	dataPrefix = flag.String("p", driverName, "Path prefix to store data under")
	backendUri = flag.String("uri", defaultConsulUri, "Backend uri")
	listenAddr = flag.String("b", "127.0.0.1:8989", "Bind address")
	version    = flag.Bool("version", false, "Show version")
)

func init() {
	flag.Parse()
	setDefaultVersionInfo()

	if *version {
		printRelease()
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Init driver config
	dc := NewDriverConfig(*backendUri, *baseDir, *dataPrefix)
	// New Driver
	driver, err := NewVolumeDriver(dc)
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
