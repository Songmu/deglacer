package main

import (
	"log"
	"os"

	"github.com/MH4GF/notion-deglacer"
)

func main() {
	if err := deglacer.Run(os.Args[1:]); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
