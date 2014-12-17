package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-core/hatchet"
	"github.com/nanobox-core/logtap"
	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/router"
	"github.com/nanobox-core/scribble"
)

//
var (
	APIPort  string
	Log      hatchet.Logger
	Logtap   *logtap.Logtap
	Mist     *mist.Mist
	Router   *router.Router
	Scribble *scribble.Driver
)

//
func Init() error {

	//
	Log = lumber.NewConsoleLogger(lumber.INFO)

	//
	config := struct {
		host                 string
		port                 string
		logtapSyslogPort     string
		logtapHistoricalPort string
		logtapHistoricalFile string
		mistPort             string
		routerPort           string
		scribbleDir          string
	}{
		host:                 "0.0.0.0",
		port:                 "1757",
		logtapSyslogPort:     "514",
		logtapHistoricalPort: "8080",
		logtapHistoricalFile: "./tmp/bolt.db",
		mistPort:             "1445",
		routerPort:           "80",
		scribbleDir:          "./tmp/db",
	}

	// command line args w/o program
	args := os.Args[1:]

	//
	if len(args) >= 1 {
		conf := args[0]

		Log.Info("Configuring nanobox using '%v'...\n", conf)

		// parse config file
		opts, err := parseFile(conf)
		if err != nil {
			return err
		}

		//
		for key, value := range opts {
			switch key {
			case "host":
				config.host = value
			case "port":
				config.port = value
			case "mist_port":
				config.mistPort = value
			case "router_port":
				config.routerPort = value
			case "logtap_port":
				config.logtapSyslogPort = value
			case "logtap_historical_port":
				config.logtapHistoricalPort = value
			case "logtap_historical_file":
				config.logtapHistoricalFile = value
			case "scribble_dir":
				config.scribbleDir = value
			default:
				Log.Info("No option: %v", value)
			}
		}
	}

	//
	APIPort = config.port

	// create new logtap
	// Logtap = logtap.New(config.logtapPort, Log)

	// create new mist
	Mist = mist.New(config.mistPort, Log)

	// create new router
	Router = router.New(config.routerPort, Log)

	// create new scribble
	Scribble = scribble.New(config.scribbleDir, Log)

	// create new logtap
	Logtap = logtap.New(Log)
	Logtap.Start()

	sysc := logtap.NewSyslogCollector(config.logtapSyslogPort)
	Logtap.AddCollector("syslog", sysc)
	sysc.Start()

	hist := logtap.NewHistoricalDrain(config.logtapHistoricalPort, config.logtapHistoricalFile, 1000)
	Logtap.AddDrain("history", hist)
	hist.Start()

	pub := logtap.NewPublishDrain(Mist)
	Logtap.AddDrain("mist", pub)

	return nil
}

// parseFile will parse a config file, returning a 'opts' map of the resulting
// config options.
func parseFile(file string) (map[string]string, error) {

	// attempt to open file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	opts := make(map[string]string)
	scanner := bufio.NewScanner(f)
	startLine := 1

	// Read line by line, sending lines to parseLine
	for scanner.Scan() {
		if err := parseLine(scanner.Text(), opts); err != nil {
			fmt.Println("Error reading line ", startLine)
			return nil, err
		}

		startLine++
	}

	return opts, err
}

// parseLine reads each line of the config file, extracting a key/value pair to
// insert into an 'opts' map.
func parseLine(line string, m map[string]string) error {

	// handle instances where we just want to skip the line and move on
	switch {

	// skip empty lines
	case len(line) <= 0:
		return nil

	// skip commented lines
	case strings.HasPrefix(line, "#"):
		return nil
	}

	// extract key/value pair
	fields := strings.Fields(line)

	// ensure expected length of 2
	if len(fields) != 2 {
		return errors.New("Incorrect format. Expecting 'key value', received: " + line)
	}

	// insert key/value pair into map
	m[fields[0]] = fields[1]

	return nil
}
