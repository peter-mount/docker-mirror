package main

import (
	mirror "github.com/peter-mount/docker-mirror"
	"log"
)

func main() {
	err := mirror.Main()

	if err != nil {
		log.Fatal(err)
	}
}
