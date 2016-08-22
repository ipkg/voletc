package main

import (
	"encoding/json"
	"flag"
	"fmt"
	//"io/ioutil"
	//"log"
	"bufio"
	"os"
	//"path/filepath"
	"strings"
)

const (
	defaultConsulUri = "consul://localhost:8500"
	defaultBaseDir   = "/opt"
)

var (
	backendUri = flag.String("H", defaultConsulUri, "Backend URI")
	dataPrefix = flag.String("prefix", driverName, "Path prefix to store data under")
	listenAddr = flag.String("b", "127.0.0.1:8989", "Bind address [server mode only]")
	baseDir    = flag.String("dir", defaultBaseDir, "Data directory")
	// Set when cli is parsed
	dryrun    = false
	answerYes = new(bool)
)

var usageHeader = `
Usage:

  voletc [options] <cmd> [name] [key=value] [key=value]

  vol etc is a distributed, persistent configuration volume with Docker
  support.  It is a tool that runs as a service as well as a tool to 
  manage volumes.

Commands:

  ls        List volumes
  create    Create new volume
  edit      Edit volume configurations
  info      Show volume info
  rm        Destroy volume i.e. remove all keys
  render    Show rendered volume templates
  version   Show version

Global Options:

  -H        Backend URI                       (default: consul://localhost:8500)
  -prefix   Prefix on filesystem and backend  (default: voletc)

Service Options:
  
  -b        Address the service listens on    (default: 127.0.0.1:8989)
  -dir      Directory to store data under     (default: /opt)
`

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

	case "rm":
		if len(args) < 2 || args[1] == "" {
			err = errInvalidConfName
			break
		}

		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err == nil {
			printDataStructue(vol)
			parseCliKeyValues(args[2:])

			if !*answerYes {
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("Are you sure you want to destroy '%s' [y/n]? : ", vol.QualifiedName())
				ans, _ := reader.ReadString('\n')
				ans = strings.TrimSuffix(ans, "\n")

				if strings.ToLower(ans) != "y" && strings.ToLower(ans) != "yes" {
					break
				}
			}

			fmt.Printf("Destroying volume (%s)...\n", vol.QualifiedName())
			err = vol.Destroy()

		}

	case "edit":
		if len(args) < 2 || args[1] == "" {
			err = errInvalidConfName
			break
		}

		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err == nil {

			if len(args[2:]) < 1 {
				err = fmt.Errorf("no data provided")
				break
			}

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
			switch {
			case strings.HasSuffix(s, "-dryrun"):
				dryrun = true

			case strings.HasSuffix(s, "-y"):
				*answerYes = true
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

var usageFooter = `* Notes:

 - Volume name format: <name>-<version>-<env>
 - Template key format: template:<name_of_file>=<content_or_filepath>
 - File paths must begin with '/' or './'
`

func printUsage() {
	fmt.Println(usageHeader)
	fmt.Println(usageFooter)
}
