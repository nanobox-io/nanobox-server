package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func parseLine(line string, m map[string]string) error {
	//Empty line
	if len(line) <= 0 {
		return nil
	}

	line = strings.TrimSpace(line)
	//Commented line
	if ([]rune(line))[0] == '#' {
		return nil
    }

	//Extract key-value pair
	splitLine := strings.Fields(line)
	if len(splitLine) != 2 {
		return errors.New("Incorrectly formatted line (expected key value): " + line)
	}
	m[splitLine[0]] = splitLine[1]
	return nil
}

//
func parseConfig(fileName string) (map[string]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		file.Close()
		return nil, err
	}

	//Read line by line, sending lines to parseLine
	m := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		err = parseLine(scanner.Text(), m)
		if err != nil {
			fmt.Println("Error on line", lineNumber)
			return nil, err
		}
		lineNumber++
	}

	return m, err
}
