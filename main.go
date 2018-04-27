package main

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
)

func main() {
	env, err := cfenv.Current()
	if err != nil {
		log.Fatalf("could not parse VCAP environment: %s\n", err)
	}

	service, err := env.Services.WithName("world-backup")
	if err != nil {
		log.Fatalf("could not get world-backup service from VCAP environment: %s\n", err)
	}
	fmt.Println(service)
}
