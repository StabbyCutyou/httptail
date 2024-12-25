package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/StabbyCutyou/httptail"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("You must pass a URL to tail")
	}
	log.Println(os.Args[1])
	t, err := httptail.NewHttpTailer(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if err := t.Tail(); err != nil {
		log.Fatal(err)
	}
	b, err := json.MarshalIndent(t.Results(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(string(b))
}
