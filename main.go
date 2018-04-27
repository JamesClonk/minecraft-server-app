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
	fmt.Print(env)
}
