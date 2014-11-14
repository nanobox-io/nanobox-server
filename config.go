package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// config parses a provided config file, or uses a default.conf
func (n *Nanobox) config() error {

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
	opts, err := parseConfig(conf)
	if err != nil {
		return err
	}

	n.opts = opts

	return nil
}

// parseConfig will parse a config file, returning a 'opts' map of the resulting
// config options.
func parseConfig(file string) (map[string]string, error) {

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
