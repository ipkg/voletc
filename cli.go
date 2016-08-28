package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/olekukonko/tablewriter"
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
	serverMode = flag.Bool("server", false, "Server mode")

	// These are client tool options
	encDec    = flag.String("e", "", "Encryption/Decryption key")
	dryrun    = false
	answerYes = new(bool)
)

var usageHeader = `
Usage:

  voletc [options] <cmd> [name] [key=value] [key=value]

  vol etc is a distributed, persistent configuration volume with Docker
  support.  It is a tool that runs as a service as well as a tool to 
  manage volumes.

  Key-value pairs are are specified in the following format: path/to/key=value.
  Templates and template files are also specified in the same format but must be
  prefixed with 'template:'.  When using file paths, absolute or relative paths
  must be specified

  Key-Value Examples:

  - Template key with file path as value.  The contents of the file are used.

    template:config.json=./etc/config.json

  - Template key with content as value

    template:config.json='{"k": "${path/to/key}"}'

  - Key-Value

    db/host=127.0.0.1

Commands:

  ls        List volumes
  create    Create new volume
  edit      Edit volume configurations
  info      Show volume info
  rm        Destroy volume i.e. remove all keys
  render    Show rendered volume templates
  mount     Mount config volume via fuse (experimental)
  version   Show version

Global Options:

  -H        Backend URI                       (default: consul://localhost:8500)
  -prefix   Prefix on filesystem and backend  (default: voletc)
  -server   Start docker plugin service

Service Options:
  
  -b        Address the service listens on    (default: 127.0.0.1:8989)
  -dir      Directory to store data under     (default: /opt)

Client Options:

  -e        Key to encrypt/decrypt data.  Must be atleast 16
            characters in length. 
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

func (c *cli) Run(args []string) error {
	var err error

	if args == nil || len(args) < 1 {
		return fmt.Errorf("command missing")
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
		if vol, err = c.ve.Get(args[1]); err != nil {
			fmt.Println("Volume", err)
			vol, err = c.buildAppConfig(args[1], args[2:])
		}

		if err == nil {
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

		if vol, err = c.buildAppConfig(args[1], args[2:]); err == nil {
			if !dryrun {
				err = vol.Commit()
				printDataStructue(vol)
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

	case "mount":
		if len(args) < 3 || args[1] == "" || args[2] == "" {
			err = errInvalidConfName
			break
		}

		var vol *AppConfig
		if vol, err = c.ve.Get(args[1]); err != nil {
			break
		}

		acfs := AppConfigFS{acfg: vol, mntPoint: args[2]}
		done := make(chan bool)

		go func() {
			if e := acfs.Mount(); e != nil {
				fmt.Println("ERR", e)
			}
			done <- true
		}()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sigs:
			err = acfs.Unmount()

		case <-done:
			err = acfs.Unmount()
		}

	case "ls":
		var vols map[string]*AppConfig
		if vols, err = c.ve.List(); err == nil {
			printVolumeTable(vols)
		}

	default:
		err = fmt.Errorf("command invalid: '%s'", args[0])

	}

	return err
}

func (c *cli) buildAppConfig(name string, args []string) (*AppConfig, error) {
	vol, err := NewAppConfigFromName(name, c.ve.be)
	if err == nil {
		if len(args) > 0 {
			ckvs := parseCliKeyValues(args)
			var reqOpts map[string][]byte
			if reqOpts, err = parseCreateReqOptions(ckvs); err == nil {
				err = vol.Set(reqOpts)
			}
		}
	}
	return vol, err

}

func printVolumeTable(vols map[string]*AppConfig) {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"volume id", "name", "version", "env", "keys", "files"})

	for _, vol := range vols {
		md := vol.Metadata()
		tw.Append([]string{
			md["id"].(string),
			md["name"].(string),
			md["version"].(string),
			md["env"].(string),
			fmt.Sprintf("%d", md["keys"]),
			fmt.Sprintf("%d", md["files"]),
		})
	}

	tw.SetHeaderLine(false)
	tw.SetColumnSeparator("")
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetBorder(false)
	tw.Render()
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

/*
var usageFooter = `* Notes:

 - Volume name format: <name>-<version>-<env>
 - Template key format: template:<name_of_file>=<content_or_filepath>
 - File paths must begin with '/' or './'
`
*/
func printUsage() {
	fmt.Println(usageHeader)
	//fmt.Println(usageFooter)
}
