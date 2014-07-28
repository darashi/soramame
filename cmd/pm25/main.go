package main

import (
	"flag"
	"log"

	"github.com/darashi/soramame"
)

func main() {
	flag.Parse()

	for _, code := range flag.Args() {
		observation, err := soramame.Fetch(code)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(observation)
	}
}
