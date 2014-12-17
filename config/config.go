package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	// "github.com/nanobox-core/logtap"
	"github.com/nanobox-core/hatchet"
	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/router"
	"github.com/nanobox-core/scribble"
)

//
type (

	//
	Config struct {
		sync.Mutex

		host        string //
		Port        string //
		logtapDir   string //
		logtapPort  string //
		mistPort    string //
		routerPort  string //
		scribbleDir string //
	}
)

//
const (
	DefaultHost       = "0.0.0.0"
	DefaultPort       = "1757"
	DefaultLogTapPort = "514"
	DefaultLogtapDir  = "./tmp/logs"
	DefaultRouterPort = "80"
)

//
var (
	Log hatchet.Logger
	// Logtap 		 logtap.Logtag
	Mist     *mist.Mist
	Router   *router.Router
	Scribble *scribble.Driver
)

//
func Init() *Config {

	//
	config := &Config{
		host:        DefaultHost,
		Port:        DefaultPort,
		logtapPort:  DefaultLogTapPort,
		logtapDir:   DefaultLogtapDir,
		mistPort:    mist.DefaultPort,
		routerPort:  DefaultRouterPort,
		scribbleDir: scribble.DefaultDir,
	}

	config.parse()

	//
	Log = hatchet.New()

	// create new scribble
	Scribble = scribble.New(config.scribbleDir, Log)

	// create new logtap
	// Logtap = logtap.New(config.logtapPort, Log)

	// create new router
	Router = router.New(config.routerPort, Log)

	// create new mist
	Mist = mist.New(config.mistPort, Log)

	return config
}

// config parses a provided config file, or uses a default.conf
func (c *Config) parse() error {

	// set default config file
	conf := "default.conf"

	// command line args w/o program
	args := os.Args[1:]

	// override default if config file provided
	if len(args) >= 1 {
		conf = args[0]
	}

	fmt.Printf("Configuring nanobox using '%v'...\n", conf)

	// parse config file
	opts, err := c.parseFile(conf)
	if err != nil {
		return err
	}

	fmt.Println("do something with these", opts)

	return nil
}

// parseFile will parse a config file, returning a 'opts' map of the resulting
// config options.
func (c *Config) parseFile(file string) (map[string]string, error) {

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
		if err := c.parseLine(scanner.Text(), opts); err != nil {
			fmt.Println("Error reading line ", startLine)
			return nil, err
		}

		startLine++
	}

	return opts, err
}

// parseLine reads each line of the config file, extracting a key/value pair to
// insert into an 'opts' map.
func (c *Config) parseLine(line string, m map[string]string) error {

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
	c.Lock()
	m[fields[0]] = fields[1]
	c.Unlock()

	return nil
}
