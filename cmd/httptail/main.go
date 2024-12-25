package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/StabbyCutyou/httptail"
)

func main() {
	// Make sure they passed something
	if len(os.Args) == 1 {
		log.Fatal("You must pass a URL to tail")
	}
	// TODO validate that it is a url, but the hard-fault they'll hit
	// if they don't pass a valud url will also verify that lol
	t, err := httptail.NewHttpTailer(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if err := t.Tail(); err != nil {
		log.Fatal(err)
	}
	// Pretty print the results, could make this an optional flag
	// if folks don't wanna waste the bytes
	b, err := json.MarshalIndent(t.Results(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(string(b))
}
