package db

import (
	"encoding/JSON"
	"fmt"
	"io/ioutil"
	"os"
)

// Write
func Write(collection, resourceID string, v interface{}, c chan<- int) {

	//
	dir, _ := os.Stat("./db/" + collection)

	if dir == nil {
		err := os.Mkdir("./db/"+collection, 0755)
		if err != nil {
			fmt.Printf("Unable to create dir '%s': %s", collection, err)
			os.Exit(1)
		}
	}

	//
	file, err := os.Create("./db/" + collection + "/" + resourceID)
	if err != nil {
		fmt.Printf("Unable to create file %s/%s: %s", collection, resourceID, err)
		os.Exit(1)
	}

	defer file.Close()

	//
	b := toJSONIndent(v)

	_, err = file.WriteString(string(b))
	if err != nil {
		fmt.Printf("Unable to write to file %s: %s", resourceID, err)
		os.Exit(1)
	}

  c <- 0
}

// Read
func Read(collection, resourceID string, v interface{}, c chan<- int) interface{} {
	b, err := ioutil.ReadFile("./db/" + collection + "/" + resourceID)
	if err != nil {
		fmt.Printf("Unable to read file %s/%s: %s", collection, resourceID, err)
		os.Exit(1)
	}

	if err := fromJSON(b, v); err != nil {
		panic(err)
	}

  c <- 0

	return v
}

// ReadAll
func ReadAll(collection string, v interface{}, c chan<- int) {
	files, err := ioutil.ReadDir("./db/" + collection)

	// if there is an error here it just means there are no evars so we wont do
	// anything
	if err != nil {
	}

	var f []string

	for _, file := range files {
		b, err := ioutil.ReadFile("./db/" + collection + "/" + file.Name())
		if err != nil {
			panic(err)
		}

		f = append(f, string(b))

	}

	b := toJSON(f)

	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}

  c <- 0
}

// Delete
func Delete(collection, resourceID string, c chan<- int) {
	err := os.Remove("./db/" + collection + "/" + resourceID)
	if err != nil {
		fmt.Printf("Unable to delete file %s/%s: %s", collection, resourceID, err)
		os.Exit(1)
	}

  c <- 0
}

// private

// toJSON converts an interface (v) into JSON bytecode
func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

// toJSONIndent
func toJSONIndent(v interface{}) []byte {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}

	return b
}

// fromJSON converts an interface (v) into JSON bytecode
func fromJSON(body []byte, v interface{}) error {
	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}

	return nil
}
