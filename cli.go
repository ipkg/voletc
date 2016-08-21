package main

import (
	"encoding/json"
	"flag"
	"fmt"
	//"io/ioutil"
	//"log"
	//"path/filepath"
	"strings"
)

const (
	defaultConsulUri = "consul://localhost:8500"
	defaultBaseDir   = "/opt"
)

var (
	baseDir    = flag.String("dir", defaultBaseDir, "Data directory")
	dataPrefix = flag.String("prefix", driverName, "Path prefix to store data under")
	backendUri = flag.String("uri", defaultConsulUri, "Backend uri")
	listenAddr = flag.String("b", "127.0.0.1:8989", "Service bind address")
	//showVersion = flag.Bool("version", false, "Show version")

	// Use for template operations
	//tfile      = flag.String("t", "", "Template file")
	//appName    = flag.String("a", "", "Name of app ( <name>-<verison>-<env> )")
	//render     = flag.Bool("r", false, "Render template")
	//commitConf = flag.Bool("commit", false, "Commit app config to backend")

	dryrun = false
)

type cli struct {
	ve *VolEtc
}

func newCli(dc *DriverConfig) (*cli, error) {
	be, err := NewBackend(dc)
	if err != nil {
		return nil, err
	}
	return &cli{ve: &VolEtc{be: be}}, nil
}

func (c *cli) Run(args []string) (bool, error) {
	var err error

	if args == nil || len(args) < 1 {
		return false, nil
	}

	switch args[0] {

	case "version":
		printRelease()

	case "render":
		if len(args) < 2 || args[1] == "" {
			err = errInvalidConfName
			break
		}

		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err == nil {

			for _, t := range vol.Templates {
				fmt.Printf("- %s:\n", t.Name)

				var rndrd []byte
				if rndrd, err = t.Render(vol.Keys.ToString()); err == nil {
					fmt.Printf("%s\n", rndrd)
				}
			}
		}

	case "create":
		if len(args) < 2 || args[1] == "" {
			err = errInvalidConfName
			break
		}

		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err == nil {
			err = fmt.Errorf("exists: '%s'", args[1])
			break
		}

		if vol, err = NewAppConfigFromName(args[1], c.ve.be); err == nil {
			if len(args[2:]) > 0 {
				ckvs := parseCliKeyValues(args[2:])
				var reqOpts map[string][]byte
				if reqOpts, err = parseCreateReqOptions(ckvs); err == nil {
					vol.Set(reqOpts)
					if !dryrun {
						err = vol.Commit()
					}
					printDataStructue(vol)
				}
			}
		}

	case "info":
		if len(args) < 2 || args[1] == "" {
			err = errInvalidConfName
			break
		}
		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err == nil {
			printDataStructue(vol)
		}

	case "ls":
		var vols map[string]*AppConfig
		if vols, err = c.ve.List(); err == nil {
			mls := make([]map[string]interface{}, len(vols))
			i := 0
			for _, vol := range vols {
				mls[i] = vol.Metadata()
				i++
			}
			printDataStructue(mls)
		}

	default:
		err = fmt.Errorf("command invalid: '%s'", args[0])

	}

	return true, err
}

// Parse cli key values into a map
func parseCliKeyValues(arr []string) map[string]string {
	m := map[string]string{}
	for _, s := range arr {
		// Treat keys starting with - specially.
		if strings.HasPrefix(s, "-") {
			if strings.HasSuffix(s, "-dryrun") {
				dryrun = true
			}
			continue
		}

		pp := strings.Split(s, "=")

		k := pp[0]
		v := strings.Join(pp[1:], "=")
		m[k] = v
	}

	return m
}

func printDataStructue(v interface{}) {
	b, _ := json.MarshalIndent(v, " ", "  ")
	fmt.Printf("%s\n", b)
}

func printUsage() {
	fmt.Println(`
Usage:

  voletc [options] <cmd> [name] [key=value] [key=value]

  A tool to manage application configuration volumes. 

Commands:

  ls        List volumes
  create    Create new volume
  info      Show volume info
  render    Show rendered volume templates
  version   Show version

Options:
`)
	flag.PrintDefaults()
	fmt.Println(`Rules:

  - Volume names: <name>-<version>-<env>
  - Template keys: template:<name_of_file>=<content_or_filepath>
  - File paths must begin with '/' or './' in order to be recognized.
`)
	//os.Exit(0)
}
