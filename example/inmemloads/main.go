package main

import (
	"log"

	bananaphone "github.com/awgh/BananaPhone/pkg/BananaPhone"
)

func main() {

	loads, err := bananaphone.InMemLoads()

	log.Println(loads, err)
}
