// A simple cat program for testing.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	for _, arg := range os.Args[1:] {
		bytes, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(bytes))
	}
}
