package main

import (
	"io/ioutil"
	"os"

	"github.com/dutchcoders/evtxparser"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	evtxparser.Parse(data)
	return
}
